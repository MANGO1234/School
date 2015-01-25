int create_cache_file(buffered_reader* in, char* host, char* resource);

int if_cache_file_exists_send_to_client(int sd, char* host, char* resource);