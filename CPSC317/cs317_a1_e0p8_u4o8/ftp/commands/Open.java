package ftp.commands;

import ftp.ErrorMessages;
import ftp.exception.FTPException;
import java.io.IOException;
import java.net.InetSocketAddress;
import java.net.Socket;
import java.net.SocketTimeoutException;
import java.net.UnknownHostException;

public class Open implements Command {
	@Override
	public void execute(FTPConnection con, String[] args) throws FTPException {
		FTPUtil.requireNotConnected(con);
		if (args.length < 2 || args.length > 3) {
			throw new FTPException(ErrorMessages.createMessage(901));
		}

		String host = args[1];
		int port;
		try {
			port = args.length == 2 ? 21 : Integer.parseInt(args[2]);
			if (port < 0 || port > 65535) throw new FTPException(ErrorMessages.createMessage(902));
		}
		catch (NumberFormatException e) {
			throw new FTPException(ErrorMessages.createMessage(902));
		}

		try {
			Socket s = new Socket();
			s.connect(new InetSocketAddress(host, port), 30000);
			con.setConnection(s);
			String reply = con.readReplyAndReturnFirstLine();

			switch (Parse.readCode(reply)) {
				//case "220":
				case "421":
				case "120": // service ready in nnn minutes, processing error?
					FTPUtil.connectionFailure(host, port);
					break;
			}
		}
		catch (UnknownHostException|SocketTimeoutException e) {
			FTPUtil.connectionFailure(host, port);
		}
		catch (IOException e) {
			FTPUtil.connectionFailure(host, port);
		}
	}
}