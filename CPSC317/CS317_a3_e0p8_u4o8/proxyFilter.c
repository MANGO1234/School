#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/types.h>
#include <sys/socket.h>
#include <netdb.h>
#include <netinet/in.h>
#include <unistd.h>
#include "queue.h"
#include "filter.h"
#include "parse.h"
#include "response.h"
#include "cache.h"
#include "constants.h"
#include "Thread.h"

#define SERVER_TCP_PORT 8000
#define REQUEST_LENGTH 6072
#define HOST_LENGTH 2024
#define RESOURCE_LENGTH 4048
#define DATA_LENGTH 8096
#define NUM_OF_THREADS 4

void* process_client(void* sd);
int read_request(int sd, char* buf, int len);
void process_request(int sd, char* request, int request_len);
int proxy_get(char* request, int request_len, char* host, char* resource, int clientsd);
int send_message(int sd, char* request, int request_len);
int recieve_data_and_pipe_to_client(int sd, int clientsd, char* host, char* resource);
void handle_error_status(int clientsd, int status);
int code_has_no_body(char* code);

int main(int argc, char **argv) {
	// determine port and start server
	int port;
	switch (argc) {
		case 3:
			port = atoi(argv[1]);
			if (read_filter_list(argv[2]) != 0) {
				fprintf(stderr, "File of filter list does not exist or IO error.");
				return -1;
			}
			break;
		default:
			fprintf(stderr, "Usage: %s [port] [file of filter list]\n", argv[0]);
			return -1;
	}
	return startServer(port);
}


int startServer(int port) {
	struct sockaddr_in server;
	int sd;

	if ((sd = socket(AF_INET, SOCK_STREAM, 0)) == -1) {
		fprintf(stderr, "Can't create a socket.\n");
		return -1;
	}

	memset((char*) &server, 0, sizeof(struct sockaddr_in));
	server.sin_family = AF_UNSPEC;  // use IPv4 or IPv6, whichever
	server.sin_port = htons(port);
	server.sin_addr.s_addr = htonl(INADDR_ANY);
	int yes=1;

	// timeout while recieveing data/accepting connection
	struct timeval timeout;
	timeout.tv_sec = 120;
	timeout.tv_usec = 0;

	if (setsockopt (sd, SOL_SOCKET, SO_RCVTIMEO, (char*) &timeout, sizeof(timeout)) < 0) {
		fprintf(stderr, "Can't set timeout on recieve\n");
		return -1;
	}//*/

	// stop the "can't bind to port" and the need to wait for 1 minute after closing
	if (setsockopt(sd, SOL_SOCKET, SO_REUSEADDR, &yes,sizeof(int)) == -1) {
		printf("Can't set to reusable address\n");
	} 

	if (bind(sd, (struct sockaddr*) &server, sizeof(server)) == -1) {
		fprintf(stderr, "Can't bind to port.\n");
		return -1;
	}

	if (listen(sd, 1) == -1) {
		fprintf(stderr, "Can't listen on socket.\n");
		return -1;
	}

	// single threaded version, use for debugging
	//process_client(&sd);

	// create a number of threads, run them, then join (which blocks)
	// the threads run forever so the main thread just sleeps forever
	// with a decent amount of threads (e.g. 10) can see a pretty high improvement on server throughoutput
	// of course debugging messages are all messed up, use single thread version if need to debug
	int i;
	void* threads[NUM_OF_THREADS];
	for (i = 0; i < NUM_OF_THREADS; ++i) {
		threads[i] = createThread(&process_client, (void*) &sd);
		runThread(threads[i], NULL);
	}
	void *ret;
	for (i = 0; i < NUM_OF_THREADS; ++i) {
		joinThread(threads[i], &ret);
		free(threads[i]);
	}//*/

	close(sd);
	return 0;
}

void* process_client(void* sd_add) {
	while (1) {
		// sleep(5); // require #include <unistd.h> (linux)
		int sd = *((int*) sd_add);
		struct sockaddr_in client;

		int new_sd;
		int client_len = sizeof(client);
		if ((new_sd = accept(sd, (struct sockaddr*) &client, &client_len)) == -1) {
			continue; // keep listening
		}

		// timeout if we can't read/send to client in 10s, so we don't have a thread waiting forever
		struct timeval timeout;
		timeout.tv_sec = 120;
		timeout.tv_usec = 0;

		if (setsockopt(new_sd, SOL_SOCKET, SO_RCVTIMEO, (char*) &timeout, sizeof(timeout)) < 0) {
			close(new_sd);
			continue;
		}

		if (setsockopt(new_sd, SOL_SOCKET, SO_SNDTIMEO, (char*) &timeout, sizeof(timeout)) < 0) {
			close(new_sd);
			continue;
		}//*/

		printf("Client accepted\n");
		char request[REQUEST_LENGTH + 1];
		int request_len = read_request(new_sd, request, REQUEST_LENGTH);
		//printf("*** REQUEST ***\n%s\n", request);
		printf("Finish reading request of length %d.\n", request_len);
		if (request_len == 0) {
			printf("Client has closed connection.\n");
		}
		else {
			process_request(new_sd, request, request_len);
		}
		printf("-------------------------------------------------------\n");
		close(new_sd);
	}
}

// Maybe TODO. Read a request from the client
int read_request(int sd, char* buf, int len) {
	int bytes_read = 0;
	int n = recv(sd, buf, len, 0);
	while (n > 0 && len > 0) {
		buf += n;
		bytes_read += n;
		len -= n;
		if (*(buf - 1) == '\n' && (*(buf - 2) == '\n' || (*(buf - 2) == '\r' && *(buf - 3) == '\n'))) break;
		n = recv(sd, buf, len, 0);
	}
	*buf = 0;
	return bytes_read;
}

// process request to get host/resource, send the appropriate response when it's not a GET request
// or there's error connecting to host/sending from client etc. Also try to read from cache if possible.
void process_request(int sd, char* request, int request_len) {
	char* host = (char*) malloc(HOST_LENGTH);
	char* resource = (char*) malloc(RESOURCE_LENGTH);
	int success = check_GET_and_get_resource(request, host, HOST_LENGTH, resource, RESOURCE_LENGTH);
	if (success != 0) {
		handle_error_status(sd, success);
		return;
	}

	if (filter_host(host) == -1) {
		send_response(sd, "403");
		return;
	}

	success = if_cache_file_exists_send_to_client(sd, host, resource);
	if (success == 0) {
		printf("Cache hit. Use cache instead of making request to host.\n");
		return;
	}
	else if (success != FILE_NOT_EXIST && success != CACHE_CONFLICT) { // todo
		printf("Cache Error.\n");
		return;
	}//*/

	printf("Getting resource %s at host %s.\n", resource, host);
	success = proxy_get(request, request_len, host, resource, sd);
	free(host);
	free(resource);
	if (success != 0) {
		handle_error_status(sd, HOST_IO_ERROR);
		return;
	}
}

// creates a socket to host, send request, and then process the response
// returns 0 on success and -1 if any errors occured (except file IO caching)
int proxy_get(char* request, int request_len, char* host, char* resource, int clientsd) {
	int sd;
	struct addrinfo hints, *res;
	res = (struct addrinfo*) malloc(sizeof(struct addrinfo));

	// initalized address structs
	memset(&hints, 0, sizeof hints);
	hints.ai_family = AF_UNSPEC;
	hints.ai_socktype = SOCK_STREAM;
	hints.ai_protocol = IPPROTO_TCP;
	char* port = get_port(host); // e.g www.hu.com:80 -> www.hu.com and 80
	if (port != NULL) printf("Using specified port %s.\n", port);
	getaddrinfo(host, port == NULL ? "80" : port, &hints, &res);

	// connect to host
	if ((sd = socket(res->ai_family, res->ai_socktype, res->ai_protocol)) == -1) {
		free(res);
		fprintf(stderr, "Can't create a socket to host.\n");
		return -1;
	}

	// timeout if we can't read/send to host in 10s
	struct timeval timeout;
	timeout.tv_sec = 10;
	timeout.tv_usec = 0;

	if (setsockopt(sd, SOL_SOCKET, SO_RCVTIMEO, (char*) &timeout, sizeof(timeout)) < 0) {
		free(res);
		return -1;
	}

	if (setsockopt(sd, SOL_SOCKET, SO_SNDTIMEO, (char*) &timeout, sizeof(timeout)) < 0) {
		free(res);
		return -1;
	}//*/

	// connect
	if (connect(sd, res->ai_addr, res->ai_addrlen) != 0) {
		printf("Error connecting to host %s.\n", host);
		free(res);
		return -1;
	}
	printf("Connected to host %s\n", host);

	// send request
	if (send_message(sd, request, request_len) != 0) {
		printf("Error sending request.\n");
		free(res);
		return -1;
	}

	return recieve_data_and_pipe_to_client(sd, clientsd, host, resource);
}


// sends a message to the socket given, return 0 on success and -1 if any error occurs
int send_message(int sd, char* message, int message_len) {
	int bytes_sent = 0;
	while (bytes_sent < message_len) {
		// MSG_NOSIGNAL -> prevent broken pipe error (happens rarely using browser that it's hard to test for sure)
		int status = send(sd, message, message_len, MSG_NOSIGNAL);
		if (status <= 0) return -1;
		bytes_sent += status;
	}
	return 0;
}


// get data from the host socket and pipe it to the client, returns 0 on success and -1 otherwise
// (except file IO caching, which doesn't effect whether the piping is succeessful or not).
int recieve_data_and_pipe_to_client(int sd, int clientsd, char* host, char* resource) {
	buffered_reader* in = malloc(sizeof(buffered_reader));
	initialize(in, DATA_LENGTH, sd, -1);

	// read code, if error occurs, handle it
	char code[4];
	if (read_code(in, code) != 0) {
		send_response(clientsd, "502");
		clean(in);
		return -1;
	}
	printf("Code: %s\n", code);

	// all the heavy work is in buffered reader, which automatically sends the data to the client
	// whenever the buffer runs out if given a clientsd, otherwise it stores the bytes in data
	// to be sent later
	// so the following are just parsing and reading and we don't have to worry other stuff
	if (code_has_no_body(code) == 0) {
		printf("Code %s has no body.\n", code);
		read_to_empty_line(in);
		flush(in); // send the remaining bytes in buffer to client
	}
	else {
		int len = read_content_length(in);
		if (len >= 0) {
			printf("Content-Length: %d.\n", len);
			readline(in);
			read_to_empty_line(in);
			advance(in, len); // read num of chars equal to len
			flush(in);
		}
		else if (len == IS_TRANSFER_ENCODING) { // -1 is chunk transfer
				printf("Is chunked transfer.\n");
				readline(in);
				read_to_empty_line(in);
				int len = read_hex(in);
				while (len != 0) {
					readline(in);
					advance(in, len);
					readline(in);
					len = read_hex(in);
				}
				read_to_empty_line(in);
				flush(in);
		}
		else if (len == NO_CONTENT_LENGTH_OR_TRANSFER_ENCODING) {
			if (in->status == 0) { // if reading is fine, then it's something wrong in response
				printf("No content length or transfer encoding found in response.\n");
				// specified section 4.4., 5
				// assume server will end connection, timeout is treated as closed by host
				while (in->status == 0) {
					flush(in);
					peek(in);
				}
				in->status = 0;
			}
			else {
				handle_error_status(clientsd, in->status);
			}
			clean(in);
			return -1;
		}
	}

	if (in->status != 0) {
		handle_error_status(clientsd, in->status);
		clean(in);
		return -1;
	}

	// send the data
	queue_t* chunk = in->data->start;
	int success = 0;
	while (chunk->next != NULL && success == 0) {
		chunk = chunk->next;
		success = send_message(clientsd, chunk->buf, chunk->len);
	}

	if (success == 0) {
		printf("Succeessful request.\n");
	}
	else {
		printf("Error occurred while sending to client.\n");
	}

	// cache
	if (strcmp(code, "200") == 0) {
		int st = create_cache_file(in, host, resource);
		if (st == 0) {
			printf("Cache file created succeessfully.\n");
		}
		else {
			printf("Something went wrong while making cache file. %d\n", st);
		}
	}//*/

	clean(in);
	return success;
}

// handles error status by sending client the associated response
void handle_error_status(int clientsd, int status) {
	switch (status) {
	case HOST_IO_ERROR:
		send_response(clientsd, "500");
		break;
	case HOST_TIMEOUT_ERROR:
		printf("Host timeout.\n");
		send_response(clientsd, "504");
		break;
	case IS_NOT_GET:
		printf("Is not get request.\n");
		send_response(clientsd, "405");
		break;
	case NO_HOST:
		send_response(clientsd, "400");
		break;
	}
	printf("Error %d while serving request.\n", status);
}

// returns 0 if code is one of 1xx, 204, 3, -1 otherwise
int code_has_no_body(char* code) {
	if (code[0] == '1') return 0;
	if (strncmp(code, "204", 3) == 0) return 0;
	if (strncmp(code, "304", 3) == 0) return 0;
	return -1;
}
