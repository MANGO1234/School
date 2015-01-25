package ftp.commands;

import java.io.*;
import java.net.InetSocketAddress;
import java.net.Socket;
import java.net.SocketTimeoutException;

public class DataConnection {
	private Socket s;
	private InputStream i;
	private OutputStream o;
	private BufferedReader r;
	public final String host;
	public final int port;

	public DataConnection(String host, int port) throws IOException, SocketTimeoutException {
		this.host = host;
		this.port = port;
		s = new Socket();
		s.connect(new InetSocketAddress(host, port), 30000);
		InputStream temp = s.getInputStream();
		i = new BufferedInputStream(temp);
		o = s.getOutputStream();
		r = new BufferedReader(new InputStreamReader(temp));
	}

	public BufferedReader getReader() {
		return r;
	}

	public InputStream getInputStream() {
		return i;
	}

	public OutputStream getOutputStream() {
		return o;
	}

	public void close() throws IOException {
		i = null;
		o = null;
		r = null;
		s.close();
	}
}