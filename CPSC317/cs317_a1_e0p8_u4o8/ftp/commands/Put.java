package ftp.commands;

import ftp.exception.FTPException;

import java.io.File;

import static ftp.commands.Commands.STOR;
import static ftp.commands.Commands.TYPE;
import static ftp.commands.Commands.PASV;

public class Put implements Command {
	@Override
	public void execute(FTPConnection con, String[] args) throws FTPException {
		FTPUtil.requireConnected(con);
		FTPUtil.requireLoggedIn(con);
		FTPUtil.requireNumberOfArguments(args, 2);
		File file = new File(args[1]);
		if (!file.exists()) {
			FTPUtil.fileAccessDeniedError(file);
		}
		PASV.execute(con, new String[]{"pasv"});
		TYPE.execute(con, new String[]{"type", "I"}); // send in binary
		Stor stor = (Stor) STOR;
		stor.setFile(file);
		stor.execute(con, new String[]{"stor", args[2]});
	}
}