/*
 * University of British Columbia
 * Department of Computer Science
 * CPSC317 - Internet Programming
 * Assignment 2
 * 
 * Author: Jonatan Schroeder
 * January 2013
 * 
 * This code may not be used without written consent of the authors, except for 
 * current and future projects and assignments of the CPSC317 course at UBC.
 */

package ubc.cs317.rtsp.client.net;

import java.io.*;
import java.net.*;
import java.util.*;

import ubc.cs317.rtsp.client.exception.RTSPException;
import ubc.cs317.rtsp.client.model.Frame;
import ubc.cs317.rtsp.client.model.Session;

/**
 * This class represents a connection with an RTSP server.
 */
public class RTSPConnection {
	private static final int BUFFER_LENGTH = 15000;
	private static final long MINIMUM_DELAY_READ_PACKETS_MS =1;
	private static final int MINIMUM_FRAME_BUFFER_SIZE = 0;
	private static final int RECOMMENDED_FRAME_BUFFER_SIZE = 50;
	private static final int POLL_BUFFER_DELAY_MS = 15;
	private static final int INIT = 0;
	private static final int READY = 1;
	private static final int PLAYING = 2;

	private Session session;
	private Timer rtpTimer;

	private Socket connection;
	private DatagramSocket rtpConnection;
	private BufferedWriter out;
	private BufferedReader in;

	private String sessionNo;
	private String vidName;
	private int seq = 1;
	private int status = INIT; // status of the server
	private boolean isPlaying;

	private Timer rtpPlayTimer;
	private Timer rtpPlayFrameTimer;
	private PriorityQueue<Frame> frameBuffer;
	private boolean isBuffering;
	private volatile int consecutivePacketLoss;
	private short lastSequenceNumber;
	private int lastTimestamp;
	private long lastSystemTimestamp;
	private short tentativeSequenceNumber;

	private static final boolean LOG = false;
	private List<FrameLog> frameLog;

	/**
	 * Establishes a new connection with an RTSP server. No message is sent at
	 * this point, and no stream is set up.
	 * 
	 * @param session
	 *            The Session object to be used for connectivity with the UI.
	 * @param server
	 *            The hostname or IP address of the server.
	 * @param port
	 *            The TCP port number where the server is listening to.
	 * @throws RTSPException
	 *             If the connection couldn't be accepted, such as if the host
	 *             name or port number are invalid or there is no connectivity.
	 */
	public RTSPConnection(Session session, String server, int port) throws RTSPException {
		this.session = session;
		try {
			connection = new Socket(server, port);
			connection.setSoTimeout(5000);
			out = new BufferedWriter(new OutputStreamWriter(connection.getOutputStream()));
			in = new BufferedReader(new InputStreamReader(connection.getInputStream()));
		}
		catch (UnknownHostException e) {
			throw new RTSPException("Unknown host.");
		}
		catch (IOException e) {
			throw new RTSPException("Unable to establish connection to RTSP server.");
		}
	}

	/**
	 * Sends a SETUP request to the server if session not already sent.
	 * Sends a PLAY request if it had not been sent.
	 * 
	 * @param videoName
	 *            The name of the video to be setup.
	 * @throws RTSPException
	 *             If there was an error sending or receiving the RTSP data, or
	 *             if the RTP socket could not be created, or if the server did
	 *             not return a successful response.
	 */
	public synchronized void setup(String videoName) throws RTSPException {
		sendSetupRequest(videoName);
		sendPlayRequest(); // immediately start buffering if possible
	}

	/**
	 * Plays the video (will be stored in buffer).
	 *
	 * @throws RTSPException
	 *             If there was an error sending or receiving the RTSP data, or
	 *             if the server did not return a successful response.
	 */
	public synchronized void play() throws RTSPException {
		sendPlayRequest(); // note that play request will not be sent if server is already in PLAY state
		if (status != INIT && !isPlaying) {
			isPlaying = true;
			lastSystemTimestamp = -1;
			startRTPPlayTimer();
		}
	}

	/**
	 * Pause playing. Cancel thread for playing the video.
	 *
	 * @throws RTSPException
	 *             If there was an error sending or receiving the RTSP data, or
	 *             if the server did not return a successful response.
	 */
	public synchronized void pause() throws RTSPException {
		isPlaying = false;
		cancelRTPPlayTimer();
	}

	/**
	 * Starts a timer that reads RTP packets repeatedly. The timer will wait at
	 * least MINIMUM_DELAY_READ_PACKETS_MS after receiving a packet to read the
	 * next one.
	 */
	private void startRTPTimer() {
		rtpTimer = new Timer();
		rtpTimer.schedule(new TimerTask() {
			@Override
			public void run() {
				receiveRTPPacket();
			}
		}, 0, MINIMUM_DELAY_READ_PACKETS_MS);
	}

	/**
	 * Receives a single RTP packet and processes the corresponding frame. The
	 * data received from the datagram socket is assumed to be no larger than
	 * BUFFER_LENGTH bytes. This data is then parsed into a Frame object (using
	 * the parseRTPPacket method) and the method session.processReceivedFrame is
	 * called with the resulting packet. In case of timeout no exception should
	 * be thrown and no frame should be processed.
	 */
	private void receiveRTPPacket() {
		try {
			byte[] buffer = new byte[BUFFER_LENGTH];
			DatagramPacket packet = new DatagramPacket(buffer, 0, buffer.length);
			rtpConnection.receive(packet);
			consecutivePacketLoss = 0;
			Frame frame = parseRTPPacket(buffer, packet.getLength());
			synchronized (this) {
				frameBuffer.offer(frame);
			}
			if (LOG) {
				System.out.println("received frame " + frame.getSequenceNumber() + " " + frame.getTimestamp());
				frameLog.add(new FrameLog(frame, System.currentTimeMillis()));
			}
		}
		catch (IOException e) {
			consecutivePacketLoss++;
		}
	}

	/**
	 * Starts a timer that reads RTP packets repeatedly. The timer will wait at
	 * least MINIMUM_DELAY_READ_PACKETS_MS after receiving a packet to read the
	 * next one.
	 */
	private void startRTPPlayTimer() {
		rtpPlayTimer = new Timer();
		rtpPlayTimer.schedule(new TimerTask() {
			@Override
			public void run() {
				playRTPPacket();
			}
		}, 0, POLL_BUFFER_DELAY_MS);
	}

	private void playRTPPacket() {
		// buffering control, play when size >= RECOMMENDED_FRAME_BUFFER_SIZE, stop when not enough frame
		int size;
		synchronized (this) { size = frameBuffer == null ? 0 : frameBuffer.size(); }
		if (size >= RECOMMENDED_FRAME_BUFFER_SIZE || consecutivePacketLoss >= 3) {
			// some frames may get stuck in buffer, since there's no way to know how long the vid is, we will
			// assume 3 consecutive packet loss (or 3 seconds where no packets arrived) is the end of vid
			isBuffering = false;
		}
		else if (size <= MINIMUM_FRAME_BUFFER_SIZE) {
			isBuffering = true;
			synchronized (session) { lastSystemTimestamp = -1; } // important so timing is correct
		}
		if (isBuffering) return;

		// synchronized to prevent reading invalid data when an ooo packet arrives and we reschedule timer 3
		// although the scenario doesn't happen for the assignment
		synchronized (session) {
			final Frame frame = findFrameToPlay();
			if (frame == null) return;
			// schedule it make sure it's run at roughly the right time, with correction made in playFrame()
			// of course normal application would just use a fps instead of doing this...
			if (tentativeSequenceNumber == frame.getSequenceNumber()) return;
			cancelPlayFrameTimer();
			rtpPlayFrameTimer = new Timer();
			tentativeSequenceNumber = frame.getSequenceNumber();
			long delay = lastSystemTimestamp < 0 ? 0 : // < 0 means user just presssed play then play frame immediately
					     frame.getTimestamp() - lastTimestamp - (System.currentTimeMillis() - lastSystemTimestamp);
			if (delay > 0) {
				rtpPlayFrameTimer.schedule(new TimerTask() {
					@Override
					public void run() {
						playFrame(frame);
					}
				}, delay);
			}
			else {
				playFrame(frame);
			}
		}
	}

	// finds a frame that we can play, if not then returns null
	public Frame findFrameToPlay() {
		synchronized (this) {
			// throws away frames that arrived too late
			Frame frame = frameBuffer.peek();
			// <= is important, if we played a frame, it will be discarded here
			while (frame != null && frame.getSequenceNumber() <= lastSequenceNumber) {
				frameBuffer.poll();
				frame = frameBuffer.peek();
			}
			return frame; // frame = null if all the packets is late, prob never happening
		}
	}

	// play a frame, record info with frame
	public void playFrame(Frame frame) {
		// synchronized to prevent timer 2 to read out of data/reschedule when ooo
		// packet came in
		synchronized (session) {
			long currentTimestamp = System.currentTimeMillis();
			session.processReceivedFrame(frame);
			if (lastSystemTimestamp < 0) {
				lastSystemTimestamp = currentTimestamp;
			}
			else {
				// this gap is needed to correct accumulation in error of the timestamps
				int gap = (int) (currentTimestamp - lastSystemTimestamp) - (frame.getTimestamp() - lastTimestamp);
				lastSystemTimestamp = currentTimestamp - gap;
			}
			lastTimestamp = frame.getTimestamp();
			lastSequenceNumber = frame.getSequenceNumber();
			if (LOG) System.out.println("play frame " + frame.getSequenceNumber() + " at " + currentTimestamp);
		}
	}

	/**
	 * Sends a TEARDOWN request to the server. This method is responsible for
	 * sending the request, receiving the response and, in case of a successful
	 * response, closing the RTP socket. This method does not close the RTSP
	 * connection, and a further SETUP in the same connection should be
	 * accepted. Also this method can be called both for a paused and for a
	 * playing stream, so the timer responsible for receiving RTP packets will
	 * also be cancelled.
	 * (Also sends PAUSE request if video playing)
	 * 
	 * @throws RTSPException
	 *             If there was an error sending or receiving the RTSP data, or
	 *             if the server did not return a successful response.
	 */
	public synchronized void teardown() throws RTSPException {
		sendPauseRequest();
		if (status == INIT) return;
		try {
			String request = "TEARDOWN " + vidName + " RTSP/1.0\n" +
					"CSeq: " + nextSequenceNumber() + "\n" +
					"Session: " + sessionNo + "\n\n";

			RTSPResponse response = sendRequestAndWaitForResponse(request);
			switch (response.getResponseCode()) {
			case 200:
				closeRTPConnection();
				if (LOG) displayStat();
				status = INIT;
				break;
			default:
				handleError(response.getResponseCode());
			}
		}
		catch (IOException e) {
			throw new RTSPException("Error I/O to server. Please try again.");
		}
	}

	/**
	 * sends a setup request, start a session with server
	 */
	public void sendSetupRequest(String videoName) throws RTSPException {
		if (status != INIT) return;
		try {
			rtpConnection = new DatagramSocket();
			rtpConnection.setSoTimeout(1000);
			int port = rtpConnection.getLocalPort();
			String request = "SETUP " + videoName + " RTSP/1.0\n" +
					"CSeq: " + nextSequenceNumber() + "\n" +
					"Transport: RTP/UDP; client_port= " + port + ";\n\n";

			RTSPResponse response = sendRequestAndWaitForResponse(request);
			switch (response.getResponseCode()) {
			case 200:
				// set up everything needed for playing
				sessionNo = response.getHeaderValue("Session");
				vidName = videoName;
				frameBuffer = new PriorityQueue<Frame>(1000);
				isBuffering = true;
				isPlaying = false;
				lastSequenceNumber = 0;
				consecutivePacketLoss = 0;
				if (LOG) frameLog = new ArrayList<>();
				status = READY;
				break;
			default:
				handleError(response.getResponseCode());
			}
		}
		catch (SocketException e) {
			throw new RTSPException("Unable to establish RTP socket.");
		}
		catch (IOException e) {
			throw new RTSPException("Error I/O to server. Please try again.");
		}
	}

	/**
	 * sends a play request, start reading data into buffer from server
	 */
	public void sendPlayRequest() throws RTSPException {
		if (status != READY) return;
		try {
			String request = "PLAY " + vidName + " RTSP/1.0\n" +
					"CSeq: " + nextSequenceNumber() + "\n" +
					"Session: " + sessionNo + "\n\n";

			RTSPResponse response = sendRequestAndWaitForResponse(request);
			switch (response.getResponseCode()) {
			case 200:
				startRTPTimer();
				status = PLAYING;
				break;
			default:
				handleError(response.getResponseCode());
			}
		}
		catch (IOException e) {
			throw new RTSPException("Error I/O to server. Please try again.");
		}
	}

	/**
	 * sends a pause request, cancelling all rtp related resource
 	 */
	public void sendPauseRequest() throws RTSPException {
		if (status != PLAYING) return;
		try {
			String request = "PAUSE " + vidName + " RTSP/1.0\n" +
					"CSeq: " + nextSequenceNumber() + "\n" +
					"Session: " + sessionNo + "\n\n";

			RTSPResponse response = sendRequestAndWaitForResponse(request);
			switch (response.getResponseCode()) {
			case 200:
				rtpTimer.cancel();
				cancelPlayFrameTimer();
				cancelRTPPlayTimer();
				status = READY;
				break;
			default:
				handleError(response.getResponseCode());
			}
		}
		catch (IOException e) {
			throw new RTSPException("Error I/O to server. Please try again.");
		}
	}


	// display stat with a session of streaming
	private void displayStat() {
		if (frameLog.size() == 0) return;
		int maxSeq = 0;
		int orderCount = 0;
		for (int i = 0; i < frameLog.size(); i++) {
			int seq = frameLog.get(i).frame.getSequenceNumber();
			if (seq < maxSeq) orderCount++;
			maxSeq = Math.max(maxSeq, seq);
		}
		int loss = maxSeq - (frameLog.size() - 1);

		long dt = (frameLog.get(frameLog.size() - 1).timestamp - frameLog.get(0).timestamp);
		System.out.println(frameLog.size() * 1000.0 / dt + " pkts/s");
		System.out.println(loss * 1000.0 / dt + " loss/s");
		System.out.println(orderCount * 1000.0 / dt + " ooo/s");
	}

	/**
	 * Closes the connection with the RTSP server. This method should also close
	 * any open resource associated to this connection, such as the RTP
	 * connection, if it is still open.
	 */
	public synchronized void closeConnection() {
		closeRTPConnection();
		try {
			connection.close();
		}
		catch (IOException e) {} // swallow
	}


	// close all resources related to the current rtpConnection
	private void closeRTPConnection() {
		if (rtpTimer != null) rtpTimer.cancel();
		cancelPlayFrameTimer();
		cancelRTPPlayTimer();
		synchronized (this) { if (frameBuffer != null) frameBuffer.clear(); }
		if (rtpConnection != null) rtpConnection.close();
	}

	// cancel play timer
	private void cancelRTPPlayTimer() {
		if (rtpPlayTimer != null) rtpPlayTimer.cancel();
	}

	// cancel play frame timer
	private void cancelPlayFrameTimer() {
		if (rtpPlayFrameTimer != null) rtpPlayFrameTimer.cancel();
	}


	/**
	 * Parses an RTP packet into a Frame object.
	 * 
	 * @param packet
	 *            the byte representation of a frame, corresponding to the RTP
	 *            packet.
	 * @return A Frame object.
	 */
	private static Frame parseRTPPacket(byte[] packet, int length) {
		byte payloadType = (byte) (packet[1] & 0b1111111);
		boolean marker = packet[1] >> 7 == 1;
		short sequenceNumber = (short) ((packet[2] << 8 & 0x0000ff00) + (packet[3] & 0xff));
		int timestamp = (packet[4] << 24 & 0xff000000) + (packet[5] << 16 & 0x00ff0000) +
		                (packet[6] << 8 & 0x0000ff00) + (packet[7] & 0x000000ff);
		int CSRCcount = packet[0] & 0b1111; // always 0 in this assignment, but w/e
		int offset = 12 + CSRCcount * 4;
		int payloadLen = length - offset;
		return new Frame(payloadType, marker, sequenceNumber, timestamp, packet, offset, payloadLen);
	}

	private int nextSequenceNumber() {
		return seq++;
	}

	// sends a request and wait for response
	private RTSPResponse sendRequestAndWaitForResponse(String request) throws IOException, RTSPException {
		System.out.print(request);
		try {
			out.write(request);
			out.flush();
			RTSPResponse response = RTSPResponse.readRTSPResponse(in);
			System.out.println(response.getResponseCode());
			System.out.println(response.getResponseMessage());
			return response;
		}
		catch (SocketTimeoutException e) {
			throw new RTSPException("Server request timeout, please try again.");
		}
	}

	// handle error code, but the way the program is coded these shouldn't happen
	private void handleError(int code) throws RTSPException {
		switch (code) {
		case 400:
			throw new RTSPException("Bad Request, no video loaded.");
		case 404:
			throw new RTSPException("File not found.");
		case 454:
			throw new RTSPException("Session not found");
		}
	}
}