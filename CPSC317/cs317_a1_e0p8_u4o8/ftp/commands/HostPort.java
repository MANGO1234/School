package ftp.commands;

public class HostPort {
	public final String host;
	public final int port;

	public HostPort(String host, int port) {
		this.host = host;
		this.port = port;
	}

	@Override
	public String toString() {
		final StringBuilder sb = new StringBuilder("HostPort{");
		sb.append("host='").append(host).append('\'');
		sb.append(", port=").append(port);
		sb.append('}');
		return sb.toString();
	}
}
