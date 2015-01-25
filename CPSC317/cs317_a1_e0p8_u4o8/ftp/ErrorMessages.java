package ftp;

import java.util.Arrays;
import java.util.HashMap;
import java.util.Map;
import java.util.Objects;

public class ErrorMessages {
	private static Map<Integer, String> msgs = new HashMap<>();
	static {
		msgs.put(900, "900 Invalid command.");
		msgs.put(901, "901 Incorrect number of arguments.");
		msgs.put(902, "902 Invalid argument.");
		msgs.put(903, "903 Supplied command not expected at this time.");
		msgs.put(910, "910 Access to local file % denied.");
		msgs.put(920, "920 Control connection to % on port % failed to open.");
		msgs.put(925, "925 Control connection I/O error, closing control connection.");
		msgs.put(930, "930 Data transfer connection to % on port % failed to open.");
		msgs.put(935, "935 Data transfer connection I/O error, closing data connection.");
		msgs.put(998, "998 Input error while reading commands, terminating.");
		msgs.put(999, "999 Processing error. %");
	}

	public static String createMessage(int code, String... args) {
		// inefficient
		String m = msgs.get(code);
		Objects.requireNonNull(m);
		int i = 0;
		while (m.contains("%")) {
			m = m.replaceFirst("%", args[i++]);
		}
		return m;
	}
}