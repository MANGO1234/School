package ftp.commands;

import ftp.ErrorMessages;
import ftp.exception.DataIOException;
import ftp.exception.FTPException;
import java.io.BufferedReader;
import java.io.IOException;

public class List implements Command {
	@Override
	public void execute(FTPConnection con, String[] args) throws FTPException {
		FTPUtil.requireConnected(con);
		FTPUtil.requireLoggedIn(con);
		FTPUtil.requireNumberOfArguments(args, 0);

		try {
			con.sendCommand("LIST");
			con.establishDataConnection(); // need data connection before reading reply
			String reply = con.readReplyAndReturnFirstLine();
			handleCode(con, Parse.readCode(reply));
			readDirectory(con);
			con.closeDataConnection();

			reply = con.readReplyAndReturnFirstLine();
			handleCode(con, Parse.readCode(reply));
		}
		catch (DataIOException e) {
			FTPUtil.handleDataIOException(con);
		}
		catch (IOException e) {
			FTPUtil.connectionIOError(con);
		}
	}

	private void readDirectory(FTPConnection con) {
		try {
			BufferedReader r = con.getDataConnection().getReader();
			String f;
			while ((f = r.readLine()) != null) {
				con.messageRecieved(f);
			}
		}
		catch (IOException e) {
			FTPUtil.dataConnectionIOError(con);
		}
	}

	private void handleCode(FTPConnection con, String code) {
		DataConnection dataCon = con.getDataConnection();
		switch (code) {
			case "226": // success
			case "150":
			case "250":
				break;
			case "125": //Shouldn't happen
				FTPUtil.unexpectedError("Data connection somehow opened.");
				break;
			case "425":
				throw new FTPException(ErrorMessages.createMessage(920, dataCon.host, Integer.toString(dataCon.port)));
			case "426":
			case "451": // error message from server should be enough
			case "450":
				break;
			default:
				FTPUtil.handleCommonErrorCode(code, con);
		}
	}
}