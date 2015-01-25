#include <stdlib.h>
#include "queue.h"

// a queue that holds char buffers that hold responses
// a more efficient version can be made (by packing the bytes)

void initialize_queue(queue* q) {
	queue_t* t = create_queue_t(NULL, 0);
	q->start = t;
	q->end = t;
}

// enqueue a queue_t
void enqueue(queue* q, queue_t* t) {
	q->end->next = t;
	q->end = t;
}

// create a queue_t from a given buffer
queue_t* create_queue_t(char* buf, int len) {
	queue_t* t = (queue_t*) malloc(sizeof(queue_t));
	t->len = len;
	t->buf = buf;
	t->next = NULL;
}

// free all resources with queue
void free_queue(queue* q) {
	queue_t* start = q->start;
	while (start->next != NULL) {
		queue_t* temp = start;
		start = start->next;
		free(temp->buf);
		free(temp);
	}
	free(start);
	free(q);
}