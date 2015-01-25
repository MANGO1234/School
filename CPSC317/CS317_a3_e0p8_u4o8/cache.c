#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/stat.h>
#include "queue.h"
#include "parse.h"
#include "constants.h"
#define BUFFER_LENGTH 8048

int send_message(int sd, char* request, int request_len);

// simple string hash, set n = 0, every 4 char gets treated as an unsigned int x, set new n = n * 31 + x
// repeat until the end. If the string length is not a multiple of 4, then the last (1/2/3) bytes are padded
// with 0 at the right. Use the int as the name of the file.
char* hash(char* host, char* resource) {
	unsigned int n = 0;
	while (resource[0] && resource[1] && resource[2] && resource[3]) {
		n = n * 31 + *((unsigned int*) resource);
		resource += 4;
	}

	char rem[4];
	int i = 0;
	while (resource[i] != 0) {
		rem[i] = resource[i];
		i++;
	}
	if (i != 0) {
		memset(rem + i, 0, 4 - i);
		n = n * 31 + *((unsigned int*) rem);
	}

	char* name = (char*) malloc(32); // an unsigned int cannot have more than 10 characters
	memcpy(name, "./cache/", 8);
	sprintf(name + 8, "%u\0", n);
	return name;
}

void create_cache_directory() {
	mkdir("./cache", S_IRWXU | S_IRWXG | S_IROTH | S_IXOTH);
}

// write all data in the buff to the file
int write_all(int fd, char* buf, int len) {
	int n = 0;
	while (n < len) {
		int k = write(fd, buf, len);
		if (k < 0) return k;
		n += k;
	}
	return n;
}

// create a given cache file from the resource name
int create_cache_file(buffered_reader* in, char* host, char* resource) {
	// open
	create_cache_directory();
	char name[] = "./cache/tempXXXXXXX";
	int fd = mkstemp(name);
	if (fd == -1) {
		close(name);
		remove(name);
		return FILE_CANNOT_BE_CREATED;
	}

	// write data
	// writing resource name in case of collision (however unlikely)
	int n = 0;
	if (n >= 0) n = write_all(fd, resource, strlen(resource));
	if (n >= 0) n = write_all(fd, "\r\n", 2);

	queue_t* chunk = in->data->start;
	while (chunk->next != NULL && n >= 0) {
		chunk = chunk->next;
		n = write_all(fd, chunk->buf, chunk->len);
	}

	// fail to write
	if (chunk->next != NULL || n < 0) {
		close(name);
		remove(name);
		return FILE_IO_ERROR;
	}

	// fail to rename
	char* cache_name = hash(host, resource);
	if (rename(name, cache_name) == -1) {
		close(name);
		remove(name);
		free(cache_name);
		return FILE_CANNOT_BE_RENAMED;
	}
	printf("File name: %s\n", cache_name);
	free(cache_name);
	return 0;
}

// sends a cache file to client, return 0 success, an error code otherwise
int if_cache_file_exists_send_to_client(int sd, char* host, char* resource) {
	// open file
	create_cache_directory();
	char* cache_name = hash(host, resource);
	FILE* file = fopen(cache_name, "r");
	if (file == NULL) {
		free(cache_name);
		return FILE_NOT_EXIST;
	}
	printf("File name: %s\n", cache_name);

	// check it's the right resource
	size_t s = 4048;
	char* cache_resource = malloc(4048);
	getline(&cache_resource, &s, file);
	*(cache_resource + strlen(cache_resource) - 2) = 0;
	if (strcmp(resource, cache_resource) != 0) {
		printf("Cache conflict: %s, %d\n", cache_resource, s);
		return CACHE_CONFLICT;
	}

	// read and sent to client
	char* buf = (char*) malloc(BUFFER_LENGTH);
	while (!feof(file)) {
		int n = fread(buf, 1, BUFFER_LENGTH, file);
		if (ferror(file)) {
			close(file);
			return FILE_IO_ERROR;
		}
		if (n != 0 && send_message(sd, buf, n) < 0) {
			close(file);
			return CLIENT_IO_ERROR;
		}
	}
	free(buf);
	free(cache_name);
	return 0;
}
