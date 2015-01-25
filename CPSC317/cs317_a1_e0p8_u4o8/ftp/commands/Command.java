package ftp.commands;

import ftp.exception.FTPException;

public interface Command {
	public void execute(FTPConnection con, String[] args) throws FTPException;
}
