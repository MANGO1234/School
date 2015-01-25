
For part A please use branch: master
For part B please use branch: master

ADDITIONAL COMMENTS

There are 3 timers in additional to the main thread. As soon as the user set up the video, a PLAY request is sent to start buffering the video.

The 1st timer reads the packets, parse them to frames and put them in the buffer. The buffer is a priority queue to simplify handling of out of order packets (a linked list queue may work but a large buffer can cause problem when an out or order packet arrives).

The 2nd timer will peridically read a frame from the buffer and set up the 3rd timer to play it at the correct time. It also handles the buffering of the buffer. When there are more than 2s of video in the buffer it will start playing and continue until the buffer is empty, at which point it waits until the buffer is filled up again. Any frames that arrived too late (i.e. a frame with higher sequence number that has already been played) are discarded. The 2nd timer polls the buffer and monitor for any out of order frame. If an out of order frame is recieved, it will cancel the current timer (to play the old frame) and set up a new timer to play the new frame. When a frame has been played, it will be removed from the buffer during the next cycle of the 2nd timer.

The third timer is set up by the 2nd timer to play a frame. It is responsible for updating the timestamp and sequence number of the last frame played along with a system timestamp of when it was played by the client. It also corrects the system timestamp to prevent accumulation of error which ensure the frames are played at the correct time.

(Testing on linux desktops at school, the client playback will be laggy, but youtube/twtich etc. will also be laggy. However, it seems the lin01-lin25 machine at room 005 is not laggy.)

(F and G has a lot of buffering.)

(I think I could have implemented this with 2 timers: the 1st timer reschedule when it receives an out of order packet which eliminated the polling from the 2nd timer, the 2nd timer plays the frame and schedule the next frame to play. I have a version for 2 timers but it's a little buggy and this version is stable and working.)

(Makefile in src should complie in case anything goes wrong)