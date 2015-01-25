#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <ctype.h>

char *banlist[100];
int len_banlist = 0;

// turn a given string to its lower case version and return pointer to string
char* to_lower_str(char *bans){
	int f = 0;
	while (bans[f]) {
        bans[f] = tolower(bans[f]);
       	f++;
    }
    return bans;
}

// read the given file and load the lsit of strings into memory
// returns 0 on success, -1 if something went wrong with IO
int read_filter_list(char *fname) {
	FILE *file;
	if ((file = fopen(fname, "r")) == NULL) {
		return -1;
	}

	int x = 0;
	while (!feof(file)) {
		char *tempstore = (char*) malloc(103); // 101 = 100 chars + null at end + 2 possible /r/n
		if (fgets(tempstore, 103, file) != NULL) {
			int len = strlen(tempstore);
			if (*(tempstore + len - 1) == '\n') { // remove new line
				*(tempstore + len - 1) = 0;
			}
			if (*(tempstore + len - 2) == '\r') { // slightly redundant but meh
				*(tempstore + len - 2) = 0;
			}
			banlist[x] = to_lower_str(tempstore);
			x++;
		}
	}
	len_banlist = x;
	return 0;
}

// given a host, go through the filter list and checked if it's banned
// return 0 if it's not banned and -1 if it's banned
int filter_host(char *str) {
	char lower[101]; // no /r/n here
	int c = 0;
	while (str[c]) {
		lower[c] = tolower(str[c]);
		c++;
	}
	lower[c] = 0;

	int z = 0;
	while (z < len_banlist) {
		if (strstr(lower, banlist[z]) != NULL) {
			printf("Host %s has been banned (banned word %s).\n", str, banlist[z]);	
			return -1;
		}
		z++;
	}
	return 0;
}