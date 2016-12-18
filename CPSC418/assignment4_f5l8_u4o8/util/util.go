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

func CheckError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error ", err.Error())
		os.Exit(-1)
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

func KvReadLineBuffer(reader *bufio.Reader) (*bytes.Buffer, error) {
	toRead, err := reader.ReadString(' ')
	if err != nil {
		return nil, err
	}
	n := StrToInt(toRead[:len(toRead)-1])
	i := 0
	var buf bytes.Buffer
	for i < n {
		k, err := reader.ReadSlice('\n')
		if err != nil {
			return nil, err
		}
		buf.Write(k)
		i += len(k)
	}
	return &buf, nil
}

func KvReadLineSlice(reader *bufio.Reader) ([]byte, error) {
	buf, err := KvReadLineBuffer(reader)
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

func KvReadLine(reader *bufio.Reader) (string, error) {
	buf, err := KvReadLineBuffer(reader)
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

func KvWriteLine(writer *bufio.Writer, str string) error {
	n := len(str)
	_, err := writer.WriteString(strconv.Itoa(n))
	if err != nil {
		return err
	}
	err = writer.WriteByte(' ')
	if err != nil {
		return err
	}
	_, err = writer.WriteString(str)
	if err != nil {
		return err
	}
	err = writer.WriteByte('\n')
	if err != nil {
		return err
	}
	err = writer.Flush()
	if err != nil {
		return err
	}
	return err
}

func KvWriteLineSlice(writer *bufio.Writer, str []byte) error {
	n := len(str)
	_, err := writer.WriteString(strconv.Itoa(n))
	if err != nil {
		return err
	}
	err = writer.WriteByte(' ')
	if err != nil {
		return err
	}
	_, err = writer.Write(str)
	if err != nil {
		return err
	}
	err = writer.WriteByte('\n')
	if err != nil {
		return err
	}
	err = writer.Flush()
	if err != nil {
		return err
	}
	return err
}
