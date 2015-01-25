package ftp.commands;

import ftp.ErrorMessages;
import ftp.exception.FTPException;

import java.io.File;
import java.io.IOException;

public class FTPUtil {
	public static void requireConnected(FTPConnection con) throws FTPException {
		if (!con.isConnected()) {
			throw new FTPException(ErrorMessages.createMessage(903));
		}
	}

	public static void requireNotConnected(FTPConnection con) throws FTPException {
		if (con.isConnected()) {
			throw new FTPException(ErrorMessages.createMessage(903));
		}
	}

	public static void requireNotLoggedIn(FTPConnection con) {
		if (con.isLoggedIn()) {
			throw new FTPException(ErrorMessages.createMessage(903));
		}
	}

	public static void requireLoggedIn(FTPConnection con) {
		if (!con.isLoggedIn()) {
			throw new FTPException(ErrorMessages.createMessage(903));
		}
	}

	public static void requireNumberOfArguments(String[] args, int n) throws FTPException {
		if (args.length-1 != n) { // -1 b/c args[0] should be the command
			throw new FTPException(ErrorMessages.createMessage(901));
		}
	}

	public static void connectionFailure(String host, int port) throws FTPException {
		throw new FTPException(ErrorMessages.createMessage(920, host, Integer.toString(port)));
	}

	public static void dataConnectionFailure(String host, int port) throws FTPException {
		throw new FTPException(ErrorMessages.createMessage(930, host, Integer.toString(port)));
	}

	public static void unexpectedError(String err) throws FTPException {
		throw new FTPException(ErrorMessages.createMessage(999, err));
	}

	public static void connectionIOError(FTPConnection con) throws FTPException {
		con.close();
		throw new FTPException(ErrorMessages.createMessage(925));
	}

	public static void dataConnectionIOError(FTPConnection con) throws FTPException {
		con.closeDataConnection();
		throw new FTPException(ErrorMessages.createMessage(935));
	}

	public static void fileAccessDeniedError(File file) throws FTPException {
		throw new FTPException(ErrorMessages.createMessage(910, file.getAbsolutePath()));
	}

	public static void handleLocalFileAccessError(FTPConnection con, File file) throws FTPException {
		// server will still reply, read it to prevent replies getting out of sync
		try {
			con.closeDataConnection();
			con.readReplyAndReturnFirstLine();
		}
		catch (IOException e1) {
			FTPUtil.connectionIOError(con);
		}
		finally {
			FTPUtil.fileAccessDeniedError(file);
		}
	}

	public static void handleDataIOException(FTPConnection con) throws FTPException {
		// server will probably respond with server code when it occur, read reply to prevent
		// it from getting out of it (it has happened to us)
		try {
			con.closeDataConnection();
			con.readReplyAndReturnFirstLine();
		}
		catch (IOException e1) {
			FTPUtil.connectionIOError(con);
		}
		finally {
			FTPUtil.dataConnectionIOError(con);
		}
	}

	public static void handleCommonErrorCode(String code, FTPConnection con) throws FTPException {
		switch (code) {
			case "332":
				FTPUtil.unexpectedError("Need ACCT. Not required to implement.");
				break;
			case "500":
			case "501":
				FTPUtil.unexpectedError("Message sent to server ill-formatted.");
				break;
			case "502":
			case "504":
				FTPUtil.unexpectedError("Server did not implement command.");
				break;
			case "421":
				FTPUtil.connectionIOError(con);
				break;
			case "503":
				throw new FTPException(ErrorMessages.createMessage(903));
			default:
				// case "530", case "200", case "230", case "331"
				// nothing
		}
	}
}