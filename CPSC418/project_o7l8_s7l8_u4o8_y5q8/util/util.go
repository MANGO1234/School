package util

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strconv"
)

const DebugOn = true

func Debug(v ...interface{}) {
	if DebugOn {
		fmt.Println(v)
	}
}

// TODO: implement a pretty print function

func CheckError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error ", err.Error())
		os.Exit(-1)
	}
}

func PrintError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error ", err.Error())
	}
}

func UInt32ToStr(n uint32) string {
	return strconv.FormatUint(uint64(n), 10)
}

func StrToUInt32(str string) uint32 {
	n, _ := strconv.ParseUint(str, 10, 32)
	return uint32(n)
}

func StrToInt(str string) int {
	n, _ := strconv.Atoi(str)
	return n
}

type MessageReader struct {
	Reader *bufio.Reader
}

func (reader *MessageReader) ReadMessageBuffer() (*bytes.Buffer, error) {
	toRead, err := reader.Reader.ReadString(' ')
	if err != nil {
		return nil, err
	}
	n := StrToInt(toRead[:len(toRead)-1])
	i := 0
	var buf bytes.Buffer
	for i < n {
		k, err := reader.Reader.ReadBytes('\n')
		if err != nil {
			return nil, err
		}
		buf.Write(k)
		i += len(k)
	}
	return &buf, nil
}

func (reader *MessageReader) ReadMessageSlice() ([]byte, error) {
	buf, err := reader.ReadMessageBuffer()
	if err == nil {
		s := buf.Bytes()
		if len(s) > 0 {
			return s[:len(s)-1], nil
		} else {
			return s[:0], nil
		}
	}
	return nil, err
}

func (reader *MessageReader) ReadMessage() (string, error) {
	buf, err := reader.ReadMessageBuffer()
	if err == nil {
		s := buf.String()
		if len(s) > 0 {
			return s[:len(s)-1], nil
		} else {
			return s[:0], nil
		}
	}
	return "", err
}

type MessageWriter struct {
	Writer *bufio.Writer
}

func (writer *MessageWriter) WriteMessage(str string) error {
	n := len(str)
	_, err := writer.Writer.WriteString(strconv.Itoa(n))
	if err != nil {
		return err
	}
	err = writer.Writer.WriteByte(' ')
	if err != nil {
		return err
	}
	_, err = writer.Writer.WriteString(str)
	if err != nil {
		return err
	}
	err = writer.Writer.WriteByte('\n')
	if err != nil {
		return err
	}
	err = writer.Writer.Flush()
	if err != nil {
		return err
	}
	return err
}

func (writer *MessageWriter) WriteMessageSlice(str []byte) error {
	n := len(str)
	_, err := writer.Writer.WriteString(strconv.Itoa(n))
	if err != nil {
		return err
	}
	err = writer.Writer.WriteByte(' ')
	if err != nil {
		return err
	}
	_, err = writer.Writer.Write(str)
	if err != nil {
		return err
	}
	err = writer.Writer.WriteByte('\n')
	if err != nil {
		return err
	}
	err = writer.Writer.Flush()
	if err != nil {
		return err
	}
	return err
}
