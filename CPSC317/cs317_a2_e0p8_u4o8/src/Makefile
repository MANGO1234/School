all: rtpClient.jar
rtpClient.jar: ubc/cs317/rtsp/client/ui/MainWindow.java
	find . -name "*.java" -print | xargs javac
	jar cvfe rtpClient.jar ubc/cs317/rtsp/client/ui/MainWindow ubc

run: rtpClient.jar
	 java -jar rtpClient.jar

clean:
	rm -f */*/*/*/*/*.class
	rm -f */*/*/*/*/*/*.class
	rm -f */*/*/*/*/*/*/*.class
	rm -f rtpClient.jar
