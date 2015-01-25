package ftp.commands;

import ftp.exception.FTPException;
import java.io.IOException;

public class Pass implements Command {
	@Override
	public void execute(FTPConnection con, String[] args) throws FTPException {
		FTPUtil.requireConnected(con);
		FTPUtil.requireNotLoggedIn(con);
		FTPUtil.requireNumberOfArguments(args, 1);

		try {
			String cmd = "PASS " + args[1];
			con.sendCommand(cmd);
			String reply = con.readReplyAndReturnFirstLine();
			if (reply.startsWith("230")) con.setLoggedIn(true);
			if (!reply.startsWith("503")) {// swallow up 503
				FTPUtil.handleCommonErrorCode(Parse.readCode(reply), con);
			}
		} catch (IOException e) {
			FTPUtil.connectionIOError(con);
		}
	}
}