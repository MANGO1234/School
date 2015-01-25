package ftp.commands;

import ftp.exception.FTPException;
import java.io.IOException;

public class Pasv implements Command {
	@Override
	public void execute(FTPConnection con, String[] args) throws FTPException {
		FTPUtil.requireConnected(con);
		FTPUtil.requireLoggedIn(con);
		FTPUtil.requireNumberOfArguments(args, 0);

		try {
			String cmd = "PASV";
			con.sendCommand(cmd);
			String reply = con.readReplyAndReturnFirstLine();
			//case "227": good
			FTPUtil.handleCommonErrorCode(Parse.readCode(reply), con);
			con.setHostPort(Parse.getHostPort(reply));
		}
		catch (IOException e) {
			FTPUtil.connectionIOError(con);
		}
	}
}