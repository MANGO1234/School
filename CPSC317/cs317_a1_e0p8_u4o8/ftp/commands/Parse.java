package ftp.commands;

import ftp.ErrorMessages;
import ftp.exception.FTPException;

import java.util.regex.MatchResult;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

public class Parse {
	public static String readCode(String str) {
		return str.trim().substring(0, 3);
	}

	static final Pattern hostPort = Pattern.compile("(\\d{1,3},){5}\\d{1,3}");
	public static HostPort getHostPort(String str) {
		Matcher m = hostPort.matcher(str);
		while (m.find()) {
			MatchResult r = m.toMatchResult();
			String[] hp = r.group().split(",");
			int[] hpi = new int[6];

			// every integer should be between 0 and 255
			boolean correct = true;
			for (int i = 0; i < 6; i++) {
				int n = hpi[i] = Integer.parseInt(hp[i]);
				if (!(0 <= n && n <= 255)) {
					correct = false;
					break;
				}
			}
			if (!correct) continue;

			StringBuilder host = new StringBuilder();
			for (int i = 0; i < 3; i++) {
				host.append(hp[i]).append('.');
			}
			host.append(hp[3]);

			int port = hpi[4]*256 + hpi[5];

			return new HostPort(host.toString(), port);
		}
		throw new FTPException(ErrorMessages.createMessage(999, "Server did not return a host/port for data connection"));
	}
}