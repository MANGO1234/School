package ftp.commands;

import ftp.ErrorMessages;
import ftp.exception.DataIOException;
import ftp.exception.FTPException;
import ftp.exception.LocalIOException;
import ftp.exception.ReplyTimeoutException;

import java.io.*;

public class Retr implements Command {
	@Override
	public void execute(FTPConnection con, String[] args) throws FTPException {
		FTPUtil.requireConnected(con);
		FTPUtil.requireLoggedIn(con);
		FTPUtil.requireNumberOfArguments(args, 1);

		try {
			con.sendCommand("RETR " + args[1]);
			con.establishDataConnection(); 
			String reply = con.readReplyAndReturnFirstLine();

			if (handleCode(con, Parse.readCode(reply))) { // if code ok, proceed
				File parent = file.getParentFile();
				if (parent != null) parent.mkdirs();

				InputStream in = con.getDataConnection().getInputStream();
				try (BufferedOutputStream out = new BufferedOutputStream(new FileOutputStream(file))) {
					pipe(in, out);
				}
				con.closeDataConnection(); // cose immediately or no server reply

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
			con.closeDataConnection(); // just in case
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
				throw new DataIOException("");
			}
			try {
				out.write(buffer, 0, len);
				out.flush();
			}
			catch (IOException e) {
				throw new LocalIOException("");
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
			case "450":
			case "550":
				return false;
			default:
				FTPUtil.handleCommonErrorCode(code, con);
		}
		return false; // pessimistic
	}

	private File file;
	public void setFile(File _file) {
		file = _file;
	}
}
