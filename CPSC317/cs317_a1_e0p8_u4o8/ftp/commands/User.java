package ftp.commands;

import ftp.exception.FTPException;
import java.io.IOException;

public class User implements Command {
	@Override
	public void execute(FTPConnection con, String[] args) throws FTPException {
		FTPUtil.requireConnected(con);
		FTPUtil.requireNotLoggedIn(con);
		FTPUtil.requireNumberOfArguments(args, 1);

		try {
			String cmd = "USER " + args[1];
			con.sendCommand(cmd);
			String reply = con.readReplyAndReturnFirstLine();
			if (reply.startsWith("230")) con.setLoggedIn(true);
			FTPUtil.handleCommonErrorCode(Parse.readCode(reply), con);
		}
		catch (IOException e) {
			FTPUtil.connectionIOError(con);
		}
	}
}