package ftp.commands;

import java.util.HashMap;
import java.util.Map;

public class Commands {
	// implementation commands
	public static final Command PASV = new Pasv();
	public static final Command LIST = new List();
	public static final Command PASS = new Pass();
	public static final Command TYPE = new Type();
	public static final Command STOR = new Stor();
	public static final Command RETR = new Retr();

	// user commands
	private static Map<String, Command> commands = new HashMap<>();
	static {
		commands.put("open", new Open());
		commands.put("user", new User());
		commands.put("close", new Close());
		commands.put("quit", new Quit());
		commands.put("dir", new Dir());
		commands.put("cd", new Cwd());
		commands.put("put", new Put());
		commands.put("get", new Get());
	}

	// return command, null if command doesn't exist
	public static Command getCommand(String cmd) {
		return commands.get(cmd);
	}
}