package ftp.commands;

import ftp.ErrorMessages;
import ftp.exception.DataIOException;
import ftp.exception.FTPException;
import ftp.exception.ReplyTimeoutException;

import java.io.*;
import java.net.Socket;
import java.net.SocketTimeoutException;

public class FTPConnection {
	private Socket s;
	private BufferedReader i;
	private Writer o;
	private HostPort hostPort;
	private DataConnection dataCon;
	private String lastReply;
	private boolean loggedIn;

	// *********************************************************************************************
	// **** GETTERS/SETTERS ************************************************************************
	public void setHostPort(HostPort hostPort) {
		this.hostPort = hostPort;
	}

	public boolean isConnected() {
		return s != null;
	}

	public Socket getConnection() {
		return s;
	}

	public BufferedReader getReader() {
		return i;
	}

	public Writer getWriter() {
		return o;
	}

	public HostPort getHostPort() {
		return hostPort;
	}

	public String getLastReply() {
		return lastReply;
	}

	public boolean isLoggedIn() {
		return loggedIn;
	}

	public void setLoggedIn(boolean b) {
		loggedIn = b;
	}

	// *********************************************************************************************
	// **** DATA CONNECTION ************************************************************************
	public void establishDataConnection() throws DataIOException {
		if (hostPort == null) {
			throw new FTPException("This should never happen");
		}
		try {
			dataCon = new DataConnection(hostPort.host, hostPort.port);
		}
		catch (SocketTimeoutException e) {
			FTPUtil.dataConnectionFailure(hostPort.host, hostPort.port);
		}
		catch (IOException e) {
			FTPUtil.dataConnectionFailure(hostPort.host, hostPort.port);
		}
		finally {
			hostPort = null;
		}
	}

	public DataConnection getDataConnection() {
		return dataCon;
	}

	public void closeDataConnection() {
		try {
			if (dataCon != null) dataCon.close();
		}
		catch (IOException e) {
			throw new FTPException(ErrorMessages.createMessage(999, "Closing data connection encountered error. Data connection closed."));
		}
		finally {
			dataCon = null;
		}
	}

	// *********************************************************************************************
	// *** CONNECTION ******************************************************************************
	public void setConnection(Socket _s) throws IOException {
		s = _s;
		i = new BufferedReader(new InputStreamReader(s.getInputStream()));
		o = new BufferedWriter(new OutputStreamWriter(s.getOutputStream()));
	}

	public void close() {
		try {
			closeDataConnection();
		}
		catch (FTPException e) {} // swallow
		try {
			s.close();
		}
		catch (IOException e) {
			throw new FTPException(ErrorMessages.createMessage(999, "Closing connection encountered error. Connection closed."));
		}
		finally {
			setLoggedIn(false);
			s = null;
			i = null;
			o = null;
			hostPort = null;
			dataCon = null;
		}
	}

	// *********************************************************************************************
	// *** I/O *************************************************************************************
	public void sendCommand(String cmd) throws IOException {
		o.write(cmd);
		o.write("\r\n");
		o.flush();
		messageSent(cmd);
	}

	public String readReplyAndReturnFirstLine() throws IOException {
		String reply = null;
		long time = System.currentTimeMillis();
		while (!i.ready() && System.currentTimeMillis() - time < 10000); // timeout  of 10s
		if (System.currentTimeMillis() - time >= 10000) throw new ReplyTimeoutException("");
		// the above doesn't do much, we treat ReplyTimeoutException as connectionIOException
		if (i.ready()) reply = i.readLine();
		if (reply == null) {
			// this can happen when the server shut down
			// so readLine returns null
			FTPUtil.connectionIOError(this);
		}

		String firstLine = reply;
		messageRecieved(reply);
		String code = Parse.readCode(reply);
		while (!(reply.startsWith(code) && reply.charAt(3) == ' ')) {
			reply = i.readLine();
			messageRecieved(reply);
		}
		lastReply = firstLine;
		return firstLine;
	}

	public void messageSent(String str) {
		System.out.println("--> " + str);
	}

	public void messageRecieved(String str) {
		System.out.println("<-- " + str);
	}
}