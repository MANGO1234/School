typedef struct queue_t_t {
	int len;
	char* buf;
	struct queue_t_t* next;
} queue_t;

typedef struct {
	queue_t* start;
	queue_t* end;
} queue;


void initialize_queue(queue* q);

void enqueue(queue* q, queue_t* t);

void free_queue(queue* q);

queue_t* create_queue_t(char* buf, int len);