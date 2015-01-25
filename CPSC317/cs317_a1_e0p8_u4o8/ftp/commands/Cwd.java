package ftp.commands;

import ftp.exception.FTPException;
import java.io.IOException;

public class Cwd implements Command {
	@Override
	public void execute(FTPConnection con, String[] args) throws FTPException {
		FTPUtil.requireConnected(con);
		FTPUtil.requireLoggedIn(con);
		FTPUtil.requireNumberOfArguments(args, 1);
		
		try{
			String cmd = "cwd " + args[1];
			con.sendCommand(cmd);
			String reply = con.readReplyAndReturnFirstLine();
			
			switch(Parse.readCode(reply)){
				//case "200": case "250":
				case "550": // server should return enough info
					break;
				default:
					FTPUtil.handleCommonErrorCode(Parse.readCode(reply), con);
			}
		}
		catch(IOException e){
			FTPUtil.connectionIOError(con);
		}
	}
}