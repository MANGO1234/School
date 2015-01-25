package ftp.commands;

import ftp.exception.FTPException;

public class Quit implements Command {
	@Override
	public void execute(FTPConnection con, String[] args) throws FTPException {
		FTPUtil.requireNumberOfArguments(args, 0);
		if (con.isConnected()) {
			Commands.getCommand("close").execute(con, new String[]{"close"});
		}
		System.exit(0);
	}
}