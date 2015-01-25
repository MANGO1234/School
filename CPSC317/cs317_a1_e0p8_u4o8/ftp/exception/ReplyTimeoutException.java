package ftp.exception;

import java.io.IOException;

public class ReplyTimeoutException extends IOException {
	public ReplyTimeoutException(String message) {
		super(message);
	}
}
