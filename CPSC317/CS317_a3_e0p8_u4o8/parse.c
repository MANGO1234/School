#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "queue.h"
#include "parse.h"
#include "constants.h"

// sends a message to the socket given, return 0 on success and -1/-2 if any error occurs
int br_send_message(int sd, char* message, int message_len) {
	int bytes_sent = 0;
	while (bytes_sent < message_len) {
		int status = send(sd, message, message_len, 0);
		if (status == 0) return CLIENT_TIMEOUT_ERROR;
		if (status < 0) return CLIENT_IO_ERROR;
		bytes_sent += status;
	}
	return 0;
}


// initialize a buffered reader
void initialize(buffered_reader* in, int buflen, int host_sd, int client_sd) {
	memset(in, 0, sizeof(buffered_reader));
	in->buflen = buflen;
	in->buf = malloc(buflen);
	in->host_sd = host_sd;
	in->client_sd = client_sd;
	queue* q = (queue*) malloc(sizeof(queue));
	initialize_queue(q);
	in->data = q;
}

// read data in buffer, set the corresponding members in buffered reader
void readbuf(buffered_reader* in) {
	// a better way of doing this is to uuse callbacks so buffered reader code is not polluterd, but meh
	if (in->status == 0 && in->client_sd >= 0 && in->bytes_in_buf > 0) {
		// send the data to client before reading new data
		// can return CLIENT_TIMEOUT_ERROR/CLIENT_IO_ERROR
		in->status = br_send_message(in->client_sd, in->buf, in->bytes_in_buf);
	}
	else if (in->status == 0 && in->client_sd < 0 && in->bytes_in_buf > 0) {
		// else store it in memory, really though the more efficient way is to stream it
		char* buf = (char*) malloc(in->bytes_in_buf);
		memcpy(buf, in->buf, in->bytes_in_buf);
		queue_t* more_bytes = create_queue_t(buf, in->bytes_in_buf);
		enqueue(in->data, more_bytes);
	}


	int bytes_read;
	bytes_read = recv(in->host_sd, in->buf, in->buflen, 0);
	if (bytes_read == 0) {
		in->status = HOST_TIMEOUT_ERROR;
	}
	else if (bytes_read < 0) {
		in->status = HOST_IO_ERROR;
	}

	in->bytes_in_buf = bytes_read <= 0 ? 0 : bytes_read;
	in->total_bytes += in->bytes_in_buf;
	in->pos = 0;
}

// read next byte in buffered reader
char next(buffered_reader* in) {
	if (in->status == 0) {
		if (in->pos == in->bytes_in_buf) {
			readbuf(in);
		}
		//printf("%c", in->buf[in->pos]);
		return in->buf[in->pos++];
	}
	else {
		printf("check status before next()\n");
		return 0;
	}
}

// peek at next byte in buffered reader without advancing
char peek(buffered_reader* in) {
	if (in->status == 0) {
		if (in->pos == in->bytes_in_buf) {
			readbuf(in);
		}
		return in->buf[in->pos];
	}
	else {
		printf("check status before peek()\n");
		return 0;
	}
}

// read a set amount of data into a given buffer, return the number of bytes read
int read(buffered_reader* in, char* buf, int num) {
	int i = 0;
	while (in->status == 0 && i < num) {
		*(buf++) = next(in);
		i++;
	}
	return i;
}

// flush the buffer, i.e. send any data to host/store buffer in in->data
void flush(buffered_reader* in) {
	in->pos = in->bytes_in_buf;
	if (in->client_sd >= 0 && in->bytes_in_buf > 0) {
		// can return CLIENT_TIMEOUT_ERROR/CLIENT_IO_ERROR
		in->status = br_send_message(in->client_sd, in->buf, in->bytes_in_buf);
	}
	else if (in->status == 0 && in->client_sd < 0 && in->bytes_in_buf > 0) {
		char* buf = (char*) malloc(in->bytes_in_buf);
		memcpy(buf, in->buf, in->bytes_in_buf);
		queue_t* more_bytes = create_queue_t(buf, in->bytes_in_buf);
		enqueue(in->data, more_bytes);
	}
}

// advance a set number of bytes (used for Content-Length, chunked encoding)
char advance(buffered_reader* in, int num) {
	while (in->status == 0 && num > 0) {
		if (in->pos == in->bytes_in_buf) {
			readbuf(in);
			if (in->bytes_in_buf == 0) {
				// no bytes read, in->status wouls've been set by readbuf(), so just quit
				return 0;
			}
		}

		// keep reading till we advance all 
		//printf("%d, %d, %d, ", in->pos, num, in->bytes_in_buf);
		if (in->bytes_in_buf - in->pos < num) { // not enough bytes in buffer
			num -= (in->bytes_in_buf - in->pos);
			in->pos = in->bytes_in_buf;
		}
		else { // enough bytes in buffer
			in->pos += num;
			num = 0;
		}
		//printf("%d, %d\n", in->pos, num);
	}
}

// free up all resources used by buffered_reader
void clean(buffered_reader* in) {
	free_queue(in->data);
	free(in->buf);
	free(in);
}


// ******************************************
// use buffered_reader to parse response

int is_space(char ch) {
	return ch == ' ' || ch == '\t';
}

int is_digit(char ch) {
	return '0' <= ch && ch <= '9';
}

int is_hex_digit(char ch) {
	return ('0' <= ch && ch <= '9') || ('A' <= ch && ch <= 'F') || ('a' <= ch && ch <= 'f');
}

// reads until end of line, next() will return first char of next line
void readline(buffered_reader* in) {
	while (in->status == 0 && peek(in) != '\n') {
		next(in);
	}
	if (in->status == 0) next(in);
}

// reads all spaces, next() will return char after space
void readspaces(buffered_reader* in) {
	while (in->status == 0 && is_space(peek(in))) {
		next(in);
	}
}

// try to read the matched string as much as possible, return numbers of char matched
int try_match(buffered_reader* in, char* str) {
	int i = 0;
	while (in->status == 0 && *str != 0 && peek(in) == *str) {
		next(in);
		i++; *str++;
	}
	return i;
}

// read http code, returns 0 on success, -1 else
int read_code(buffered_reader* in, char* str) {
	if (try_match(in, "HTTP/1.") != 7) return -1;
	if (try_match(in, "1") != 1 && try_match(in, "0") != 1) return -1;
	next(in);
	readspaces(in);
	if (read(in, str, 3) != 3) return -1;
	if (is_digit(str[0]) && is_digit(str[1]) && is_digit(str[2])) {
		str[3] = 0;
		return 0;
	}
	return -1;
}

// reads integer, used to read Content-Length
int read_int(buffered_reader* in) {
	char num[32];
	int i = 0;
	while (i < 31 && in->status == 0 && is_digit(peek(in))) {
		num[i++] = next(in);
	}
	num[i] = 0;
	return atoi(num);
}

// read the content_length, returns IS_TRANSFER_ENCODING if it's in chunked encoding, NO_CONTENT_LENGTH_OR_TRANSFER_ENCODING if it's neither
int read_content_length(buffered_reader* in) {
	while (in->status == 0) {
		if (peek(in) == '\n') {
			return NO_CONTENT_LENGTH_OR_TRANSFER_ENCODING;
		}
		else if (peek(in) == '\r') {
			if (try_match(in, "\r\n") == 2) {
				return NO_CONTENT_LENGTH_OR_TRANSFER_ENCODING;
			}
		}
		else if (peek(in) == 'C') {
			if (try_match(in, "Content-Length:") == 15) {
				readspaces(in);
				int len = read_int(in);
				return len;
			}
		}
		else if (peek(in) == 'T') {
			if (try_match(in, "Transfer-Encoding:") == 18) {
				readspaces(in);
				if (try_match(in, "chunked") == 7) return IS_TRANSFER_ENCODING;
				else return NO_CONTENT_LENGTH_OR_TRANSFER_ENCODING;
			}
		}
		readline(in);
	}
	return NO_CONTENT_LENGTH_OR_TRANSFER_ENCODING;
}

// read to empty line in response, assuming readline() is called before this, returns 0 on success, -1 otherwise
int read_to_empty_line(buffered_reader* in) {
	while (in->status == 0) {
		if (peek(in) == '\n') {
			return 0;
		}
		else if (peek(in) == '\r') {
			if (try_match(in, "\r\n") == 2) {
				return 0;
			}
		}
		readline(in);
	}
	return -1;
}

// reads hex, use to read chunked encoding
int read_hex(buffered_reader* in) {
	char num[32];
	int i = 0;
	while (i < 31 && in->status == 0 && is_hex_digit(peek(in))) {
		num[i++] = next(in);
	}
	num[i] = 0;
	return strtol(num, NULL, 16);
}




// may revamp this later, for now it works ok
char* read_till_space(char* buffer) {
	while (!is_space(*buffer) && *buffer != '\0') buffer++;
	return buffer;
}

char* read_till_non_space(char* buffer) {
	while (is_space(*buffer) && *buffer != '\0') buffer++;
	return buffer;
}

char* read_to_new_line(char* buffer) {
	while (*buffer != '\n' && *buffer != '\0') buffer++;
	return buffer + 1;
}

// parse a request and put the host and resource into the given buffer, returns 0 on success and -1 otherwise
int check_GET_and_get_resource(char* buffer, char* host, int host_len, char* resource, int resource_len) {
	if (strncmp("GET ", buffer, 4) == 0) {
		buffer += 4;
		buffer = read_till_non_space(buffer);

		char* end = read_till_space(buffer);
		memcpy(resource, buffer, end - buffer);
		*(resource + (end - buffer)) = 0;

		while (*buffer != '\n' && !(*buffer == '\r' && *(buffer + 1) == '\r')) {
			buffer = read_to_new_line(buffer);
			if (strncmp("Host:", buffer, 5) == 0) {
				buffer += 5;
				buffer = read_till_non_space(buffer);

				end = read_to_new_line(buffer);
				end--;
				if (*(end - 1) == '\r') end--;
				memcpy(host, buffer, end - buffer);
				*(host + (end - buffer)) = 0;
				return 0;
			}
		}
		return NO_HOST;
	}
	else {
		return IS_NOT_GET;
	}
}

// find the empty line, returns the pointer at empty line on success, NULL otherwise
char* read_to_after_empty_line(char* buffer) {
	while (*buffer != '\0') {
		buffer = read_to_new_line(buffer);
		if (*(buffer - 1) == '\0') return NULL;

		if (buffer[0] == '\n')                           return buffer + 1;
		else if (buffer[0] == '\r' && buffer[1] == '\n') return buffer + 2;
	}
	return NULL;
}

// get the port in a host name i.e. www.hu.com:80, return NULL is none is fine
char* get_port(char* host) {
	// read to end, go back till we reach a non digit and check if it's ':'
	while (*host != 0) host++;
	host--;
	if (!is_digit(*host)) return NULL;
	while (is_digit(*host)) host--;
	if (*host == ':') {
		*host = 0;
		return host + 1;
	}
	else {
		return NULL;
	}
}