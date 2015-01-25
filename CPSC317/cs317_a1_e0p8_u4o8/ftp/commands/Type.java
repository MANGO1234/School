package ftp.commands;

import ftp.exception.FTPException;
import java.io.IOException;

public class Type implements Command {
	@Override
	public void execute(FTPConnection con, String[] args) throws FTPException {
		FTPUtil.requireConnected(con);
		FTPUtil.requireLoggedIn(con);
		FTPUtil.requireNumberOfArguments(args, 1);
		// note that type can carry 2 arg but this is enough for assignment since it's not directly invoked by user

		try {
			String cmd = "TYPE " + args[1];
			con.sendCommand(cmd);
			String reply = con.readReplyAndReturnFirstLine();
			FTPUtil.handleCommonErrorCode(Parse.readCode(reply), con);
		}
		catch (IOException e) {
			FTPUtil.connectionIOError(con);
		}
	}
}
