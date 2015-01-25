package ubc.cs317.rtsp.client.net;

import ubc.cs317.rtsp.client.model.Frame;

public class FrameLog {
	public final Frame frame;

	public FrameLog(Frame frame, long timestamp) {
		this.frame = frame;
		this.timestamp = timestamp;
	}

	public final long timestamp;
}
