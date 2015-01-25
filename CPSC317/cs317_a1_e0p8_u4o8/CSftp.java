import ftp.ErrorMessages;
import ftp.commands.*;
import ftp.exception.FTPException;

import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStreamReader;

public class CSftp {
	public static void main(String[] args) {
		FTPConnection con = new FTPConnection();
		BufferedReader r = new BufferedReader(new InputStreamReader(System.in));
		try {
			while (true) {
				System.out.print("csftp> ");
				String line = r.readLine();
				if (line == null) continue;
				if (line.equals("")) continue;
				if (line.charAt(0) == '#') continue;
				// infinite loop, cmd 'quit' quits the application

				try {
					String[] words = line.trim().split("\\s+");
					Command cmd = Commands.getCommand(words[0]);
					if (cmd == null) {
						System.out.println(ErrorMessages.createMessage(900));
					}
					else {
						cmd.execute(con, words);
					}

					// special interaction with user
					if (words[0].equals("user")) {
						switch (Parse.readCode(con.getLastReply())) {
							case "230":
								System.out.println("Log in successful");
							case "530":
								System.out.println("Log in denied");
								break;
							case "331":
								System.out.println("Please enter password: ");
								System.out.print("csftp> ");
								String pw = r.readLine();
								// do we need this if we are actually going to disply the PASS request anyway
								//String pw = String.valueOf(System.console().readPassword());
								Commands.PASS.execute(con, new String[]{"pass", pw});
						}
					}
				}
				catch (FTPException e) {
					System.out.println(e.getMessage());
				}
			}
		}
		catch (IOException e) {
			System.out.println(ErrorMessages.createMessage(998));
		}
	}
}