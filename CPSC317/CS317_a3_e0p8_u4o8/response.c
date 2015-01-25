#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/types.h>
#include <sys/socket.h>
#include <netdb.h>
#include <netinet/in.h>

#define CODE_400 "HTTP/1.1 400 Bad Request\r\nContent-Length: 0\r\n\r\n"
#define CODE_400_LEN 47
#define CODE_403 "HTTP/1.1 403 Forbidden\r\nContent-Length: 0\r\n\r\n"
#define CODE_403_LEN 45
#define CODE_404 "HTTP/1.1 404 Not Found\r\nContent-Length: 0\r\n\r\n"
#define CODE_404_LEN 45
#define CODE_405 "HTTP/1.1 405 Method Not Allowed\r\nContent-Length: 0\r\nAccept: GET\r\n\r\n"
#define CODE_405_LEN 67
#define CODE_500 "HTTP/1.1 500 Internal Server Error\r\nContent-Length: 0\r\n\r\n"
#define CODE_500_LEN 57
#define CODE_502 "HTTP/1.1 502 Bad Gateway\r\nContent-Length: 0\r\n\r\n"
#define CODE_502_LEN 47
#define CODE_504 "HTTP/1.1 504 Gateway Timeout\r\nContent-Length: 0\r\n\r\n"
#define CODE_504_LEN 51

// send a message
int res_send_message(int sd, char* message, int message_len) {
	int bytes_sent = 0;
	while (bytes_sent < message_len) {
		int status = send(sd, message, message_len, 0);
		if (status < 0) return -1;
		bytes_sent += status;
	}
	return 0;
}

// sends a response depending on given code
int send_response(int sd, char* response_code) {
	if (response_code[0] == '4') {
		if (strcmp(response_code, "400") == 0) {
			return res_send_message(sd, CODE_400, CODE_404_LEN);
		}
		if (strcmp(response_code, "403") == 0) {
			return res_send_message(sd, CODE_403, CODE_403_LEN);
		}
		if (strcmp(response_code, "404") == 0) {
			return res_send_message(sd, CODE_404, CODE_404_LEN);
		}
		if (strcmp(response_code, "405") == 0) {
			return res_send_message(sd, CODE_405, CODE_405_LEN);
		}
	}
	else if (response_code[0] == '5') {
		if (strcmp(response_code, "500") == 0) {
			return res_send_message(sd, CODE_500, CODE_500_LEN);
		}
		if (strcmp(response_code, "502") == 0) {
			return res_send_message(sd, CODE_502, CODE_502_LEN);
		}
		if (strcmp(response_code, "504") == 0) {
			return res_send_message(sd, CODE_504, CODE_504_LEN);
		}
	}
}
