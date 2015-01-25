typedef struct {
	char* buf;
	int buflen;
	int pos;
	int bytes_in_buf;
	int host_sd;
	queue* data; // contains all the bytes read in a queue
	int total_bytes;
	int client_sd;
	int status;
} buffered_reader;

void initialize(buffered_reader* in, int buflen, int host_sd, int client_sd);

char next(buffered_reader* in);

char peek(buffered_reader* in);

void flush(buffered_reader* in);

char advance(buffered_reader* in, int num);

void clean(buffered_reader* in);

void readline(buffered_reader* in);

int try_match(buffered_reader* in, char* str);

int read_code(buffered_reader* in, char* str);

int read_content_length(buffered_reader* in);

int read_to_empty_line(buffered_reader* in);

int read_hex(buffered_reader* in);




char* read_to_after_empty_line(char* buffer);

int check_GET_and_get_resource(char* buffer, char* host, int host_len, char* resource, int resource_len);

char* get_port(char* host);