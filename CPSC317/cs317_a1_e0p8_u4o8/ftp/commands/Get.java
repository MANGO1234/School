package ftp.commands;

import ftp.exception.FTPException;

import java.io.File;

import static ftp.commands.Commands.RETR;
import static ftp.commands.Commands.TYPE;
import static ftp.commands.Commands.PASV;

public class Get implements Command {
	@Override
	public void execute(FTPConnection con, String[] args) throws FTPException {
		FTPUtil.requireConnected(con);
		FTPUtil.requireLoggedIn(con);
		FTPUtil.requireNumberOfArguments(args, 2);
		PASV.execute(con, new String[]{"pasv"});
		TYPE.execute(con, new String[]{"type", "I"});
		File file = new File(args[2]);
		Retr retr = (Retr) RETR;
		retr.setFile(file);
		retr.execute(con, new String[]{"retr", args[1]});
	}
}