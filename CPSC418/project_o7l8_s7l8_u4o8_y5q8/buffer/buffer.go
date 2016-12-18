package buffer

import (
	"bytes"
)

const NO_OPERATION = byte(0)
const INSERT = byte(1)
const REMOTE_INSERT = byte(2)
const DELETE = byte(3)

type BufferOperation struct {
	Type byte
	Pos  int
	Atom byte
}

type Line struct {
	Prev  *Line
	Next  *Line
	Bytes []byte
}

type Buffer struct {
	width           int
	numberOfChars   int
	numberOfLines   int
	lines           *Line
	currentLine     *Line
	currentPosition int
	currentX        int
	currentY        int
}

func StringToBuffer(str string, width int) *Buffer {
	buf := &Buffer{width: width}
	buf.lines, buf.numberOfLines, buf.numberOfChars = SeqToLines(NewStringSequence(str), width)
	buf.currentLine = buf.lines
	buf.currentPosition = 0
	buf.currentX = 0
	buf.currentY = 0
	return buf
}

func SeqToLines(seq WordSequence, width int) (*Line, int, int) {
	line := &Line{Bytes: make([]byte, 0, width+2)}
	startLine := line
	i := 0
	numberOfLines := 1
	numberOfChars := 0
	for {
		word := seq.NextWord()
		if word == nil {
			break
		}
		numberOfChars += len(word)

		if word[0] == '\n' {
			line.Bytes = append(line.Bytes, '\n')
			lastLine := line
			line = &Line{Bytes: make([]byte, 0, width), Prev: lastLine}
			lastLine.Next = line
			i = 0
			numberOfLines++
			continue
		}

		wordLen := wordLength(word)
		if i+wordLen <= width {
			line.Bytes = append(line.Bytes, word...)
			i += wordLen
			continue
		}

		if wordLen <= width {
			lastLine := line
			line = &Line{Bytes: make([]byte, 0, width+2), Prev: lastLine}
			lastLine.Next = line
			line.Bytes = append(line.Bytes, word...)
			i = wordLen
			numberOfLines++
			continue
		} else {
			line.Bytes = append(line.Bytes, word[:width-i]...)
			k := width - i
			for k < wordLen {
				lastLine := line
				line = &Line{Bytes: make([]byte, 0, width+2), Prev: lastLine}
				lastLine.Next = line
				i = width
				if wordLen-k < width {
					i = wordLen - k
				}
				line.Bytes = append(line.Bytes, word[k:k+i]...)
				numberOfLines++
				k += i
			}
			continue
		}
	}
	// a sentinel \n at the end to remove edge cases from various methods
	line.Bytes = append(line.Bytes, '\n')
	return startLine, numberOfLines, numberOfChars
}

func (buf *Buffer) Resize(width int) {
	buf.lines, buf.numberOfLines, buf.numberOfChars = SeqToLines(NewLineSequence(buf.lines), width)
	buf.SetPosition(buf.currentPosition)
	buf.width = width
}

func (buf *Buffer) MoveLeft() {
	if buf.currentX > 0 {
		buf.currentX--
		buf.currentPosition--
	} else if buf.currentLine.Prev != nil {
		buf.currentLine = buf.currentLine.Prev
		buf.currentY--
		buf.currentX = len(buf.currentLine.Bytes) - 1
		buf.currentPosition--
	}
}

func (buf *Buffer) MoveRight() {
	if buf.currentX < len(buf.currentLine.Bytes)-1 {
		buf.currentX++
		buf.currentPosition++
	} else if buf.currentLine.Next != nil {
		buf.currentLine = buf.currentLine.Next
		buf.currentY++
		buf.currentX = 0
		buf.currentPosition++
	}
}

func (buf *Buffer) MoveUp() {
	if buf.currentLine.Prev != nil {
		target := sliceLength(buf.currentLine.Bytes[:buf.currentX])
		buf.currentPosition -= buf.currentX
		buf.currentLine = buf.currentLine.Prev
		buf.currentY--
		buf.currentX = 0
		s := 0
		for _, ch := range buf.currentLine.Bytes {
			s += charLength(ch)
			if s > target || ch == '\n' {
				break
			}
			buf.currentX++
			if s == target {
				break
			}
		}
		buf.currentPosition -= len(buf.currentLine.Bytes) - buf.currentX
	}
}

func (buf *Buffer) MoveDown() {
	if buf.currentLine.Next != nil {
		target := sliceLength(buf.currentLine.Bytes[:buf.currentX])
		buf.currentPosition += len(buf.currentLine.Bytes) - buf.currentX
		buf.currentLine = buf.currentLine.Next
		buf.currentY++
		buf.currentX = 0
		s := 0
		for _, ch := range buf.currentLine.Bytes {
			s += charLength(ch)
			if s > target || ch == '\n' {
				break
			}
			buf.currentX++
			if s == target {
				break
			}
		}
		buf.currentPosition += buf.currentX
	}
}

func (buf *Buffer) findPos(pos int) (int, int, int, *Line) {
	currentPosition := pos
	currentLine := buf.lines
	currentY := 0
	currentX := 0
	acc := 0
	for {
		if acc+len(currentLine.Bytes) > pos {
			currentX = pos - acc
			break
		} else {
			acc += len(currentLine.Bytes)
			currentLine = currentLine.Next
			currentY++
		}
	}
	return currentPosition, currentX, currentY, currentLine
}

func (buf *Buffer) InsertAtCurrent(ch byte) {
	buf.Insert(buf.currentPosition, ch)
}

func (buf *Buffer) Insert(pos int, ch byte) {
	_, currentX, _, currentLine := buf.findPos(pos)
	currentLine.Bytes = append(currentLine.Bytes, 0)
	copy(currentLine.Bytes[currentX+1:], currentLine.Bytes[currentX:])
	currentLine.Bytes[currentX] = ch
	if pos <= buf.currentPosition {
		buf.currentPosition++
	}
	buf.Resize(buf.width)
}

// only difference is < vs <= that removes anomaly when a remote insertion happen at the cursor location
// the cursor moving forward cause interleaving inserts between two sites
func (buf *Buffer) RemoteInsert(pos int, ch byte) {
	_, currentX, _, currentLine := buf.findPos(pos)
	currentLine.Bytes = append(currentLine.Bytes, 0)
	copy(currentLine.Bytes[currentX+1:], currentLine.Bytes[currentX:])
	currentLine.Bytes[currentX] = ch
	if pos < buf.currentPosition {
		buf.currentPosition++
	}
	buf.Resize(buf.width)
}

func (buf *Buffer) DeleteAtCurrent() {
	if buf.numberOfChars > 0 {
		buf.Delete(buf.currentPosition)
	}
}

func (buf *Buffer) BackspaceAtCurrent() {
	if buf.currentPosition > 0 {
		buf.Delete(buf.currentPosition - 1)
	}
}

func (buf *Buffer) Delete(pos int) {
	_, currentX, _, currentLine := buf.findPos(pos)
	currentLine.Bytes = append(currentLine.Bytes[:currentX], currentLine.Bytes[currentX+1:]...)
	if pos < buf.currentPosition {
		buf.currentPosition--
	}
	buf.Resize(buf.width)
}

func (buf *Buffer) ApplyOperation(bufOp BufferOperation) {
	if bufOp.Type == INSERT {
		buf.Insert(bufOp.Pos, bufOp.Atom)
	} else if bufOp.Type == REMOTE_INSERT {
		buf.RemoteInsert(bufOp.Pos, bufOp.Atom)
	} else if bufOp.Type == DELETE {
		buf.Delete(bufOp.Pos)
	}
}

func (buf *Buffer) SetPosition(pos int) {
	buf.currentPosition = pos
	buf.currentLine = buf.lines
	buf.currentY = 0
	buf.currentX = 0
	acc := 0
	for {
		if acc+len(buf.currentLine.Bytes) > pos {
			buf.currentX = pos - acc
			break
		} else {
			acc += len(buf.currentLine.Bytes)
			buf.currentLine = buf.currentLine.Next
			buf.currentY++
		}
	}
}

func (buf *Buffer) GetPosition() int {
	return buf.currentPosition
}

func (buf *Buffer) GetSize() int {
	return buf.numberOfChars
}

func (buf *Buffer) GetDisplayInformation(screenY, height int) (int, int, int, *Line) {
	if buf.currentY < screenY {
		screenY = buf.currentY
	} else if buf.currentY >= screenY+height {
		screenY = buf.currentY - height + 1
	}
	cursorY := buf.currentY - screenY
	cursorX := sliceLength(buf.currentLine.Bytes[:buf.currentX])
	line := buf.currentLine
	for i := 0; i < cursorY; i++ {
		line = line.Prev
	}
	return screenY, cursorX, cursorY, line
}

func (buf *Buffer) Lines() *Line {
	return buf.lines
}

func (buf *Buffer) ToString() string {
	builder := bytes.Buffer{}
	lines := buf.lines
	builder.Write(lines.Bytes[0:])
	lines = lines.Next
	for lines != nil {
		builder.Write(lines.Bytes)
		lines = lines.Next
	}
	builder.Truncate(builder.Len() - 1)
	return builder.String()
}

func NumberOfLines(lines *Line) int {
	i := 0
	for lines != nil {
		i++
		lines = lines.Next
	}
	return i
}
