package ftp.commands;

import ftp.ErrorMessages;
import ftp.exception.DataIOException;
import ftp.exception.FTPException;
import ftp.exception.LocalIOException;
import java.io.*;

public class Stor implements Command {
	@Override
	public void execute(FTPConnection con, String[] args) throws FTPException {
		FTPUtil.requireConnected(con);
		FTPUtil.requireLoggedIn(con);
		FTPUtil.requireNumberOfArguments(args, 1);

		try {
			con.sendCommand("STOR " + args[1]);
			con.establishDataConnection(); // need data connection before reading reply
			String reply = con.readReplyAndReturnFirstLine();

			if (handleCode(con, Parse.readCode(reply))) { // if ok, proceed to store file
				OutputStream out = con.getDataConnection().getOutputStream();
				try (InputStream in = new FileInputStream(file)) {
					pipe(in, out);
				}
				con.closeDataConnection(); // close immediately or no reply from servers

				reply = con.readReplyAndReturnFirstLine();
				handleCode(con, Parse.readCode(reply));
			}
		}
		catch (FileNotFoundException|LocalIOException e) {
			FTPUtil.handleLocalFileAccessError(con, file);
		}
		catch (DataIOException e) {
			FTPUtil.handleDataIOException(con);
		}
		catch (IOException e) {
			FTPUtil.connectionIOError(con);
		}
		finally {
			con.closeDataConnection(); // probably should named closeIfNotClosed
			file = null;
		}
	}

	private void pipe(InputStream in, OutputStream out) throws LocalIOException, DataIOException {
		byte[] buffer = new byte[1024 * 100]; // 10kb
		while (true) {
			int len;
			try {
				len = in.read(buffer);
				if (len == -1) break;
			}
			catch (IOException e) {
				throw new LocalIOException("");
			}
			try {
				out.write(buffer, 0, len);
				out.flush();
			}
			catch (IOException e) {
				throw new DataIOException("");
			}
		}
	}

	private boolean handleCode(FTPConnection con, String code) {
		DataConnection dataCon = con.getDataConnection();
		switch (code) {
			case "226": // success
			case "150":
			case "250":
				return true;
			case "125": //Shouldn't happen
				FTPUtil.unexpectedError("Data connection somehow opened.");
				break;
			case "425":
				// this shouldn't really happen
				throw new FTPException(ErrorMessages.createMessage(920, dataCon.host, Integer.toString(dataCon.port)));
			case "110":
			case "426":
			case "451":
			case "551":
			case "552":
			case "452":
			case "532":
			case "553":
			case "550":
			case "450": // server should return sufficient info for display
				return false;
			default:
				FTPUtil.handleCommonErrorCode(code, con);
		}
		return false; // pessimistic
	}

	// slightly hacky
	private File file;
	public void setFile(File _file) {
		file = _file;
	}
}
