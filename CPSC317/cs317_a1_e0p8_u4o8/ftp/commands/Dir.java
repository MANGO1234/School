package ftp.commands;

import ftp.exception.FTPException;
import static ftp.commands.Commands.*;

public class Dir implements Command {
	@Override
	public void execute(FTPConnection con, String[] args) throws FTPException {
		FTPUtil.requireConnected(con);
		FTPUtil.requireLoggedIn(con);
		FTPUtil.requireNumberOfArguments(args, 0);
		PASV.execute(con, new String[]{"pasv"});
		LIST.execute(con, new String[]{"list"});
	}
}