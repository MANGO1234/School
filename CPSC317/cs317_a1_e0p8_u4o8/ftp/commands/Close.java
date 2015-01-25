package ftp.commands;

import ftp.exception.FTPException;
import java.io.IOException;

public class Close implements Command {
	@Override
	public void execute(FTPConnection con, String[] args) throws FTPException {
		FTPUtil.requireConnected(con);
		FTPUtil.requireNumberOfArguments(args, 0);

		try {
			String cmd = "QUIT";
			con.sendCommand(cmd);
			String reply = con.readReplyAndReturnFirstLine();
			FTPUtil.handleCommonErrorCode(Parse.readCode(reply), con);
		}
		catch (IOException e) {
			FTPUtil.connectionIOError(con);
		}
		finally {
			con.close();
		}
	}
}