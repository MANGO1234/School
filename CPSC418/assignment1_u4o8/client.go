/*
Implements the solution to assignment 1 for UBC CS 416 2015 W2.

Usage:
$ go run client.go [local UDP ip:port] [aserver UDP ip:port] [secret]

Example:
$ go run client.go 127.0.0.1:2020 127.0.0.1:7070 1984

*/

package main

import (
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
)

/////////// Msgs used by both auth and fortune servers:

// An error message from the server.
type ErrMessage struct {
	Error string
}

/////////// Auth server msgs:

// Message containing a nonce from auth-server.
type NonceMessage struct {
	Nonce int64
}

// Message containing an MD5 hash from client to auth-server.
type HashMessage struct {
	Hash string
}

// Message with details for contacting the fortune-server.
type FortuneInfoMessage struct {
	FortuneServer string
	FortuneNonce  int64
}

/////////// Fortune server msgs:

// Message requesting a fortune from the fortune-server.
type FortuneReqMessage struct {
	FortuneNonce int64
}

// Response from the fortune-server containing the fortune.
type FortuneMessage struct {
	Fortune string
}

func main() {
	lServerIp := os.Args[1]
	lAddr, err := net.ResolveUDPAddr("udp", lServerIp)
	errExit(err, "Unknown local address")
	lDialer := net.Dialer{LocalAddr: lAddr}
	aServerIp := os.Args[2]
	secret, err := strconv.ParseInt(os.Args[3], 10, 64)
	errExit(err, "Secret is not an integer")

	var buf [1024]byte

	// connect to aserver, ask for nonce
	aConn, err := lDialer.Dial("udp", aServerIp)
	errExit(err, "Cannot connect to aserver")

	_, err = aConn.Write(buf[:])
	errExit(err, "Error communicating with aserver")

	// get nonce
	readN, err := aConn.Read(buf[:])
	errExit(err, "Error communicating with aserver")
	checkForErrMessage(buf[0:readN])

	var nonce *NonceMessage
	err = json.Unmarshal(buf[0:readN], &nonce)
	errExit(err, "Cannot parse NonceMessage")

	// read md5, send it to aserver
	readN = binary.PutVarint(buf[:], nonce.Nonce+secret)
	md5 := md5.Sum(buf[0:readN])
	tmpBuf, err := json.Marshal(HashMessage{Hash: hex.EncodeToString(md5[:])})
	_, err = aConn.Write(tmpBuf[:])
	errExit(err, "Error communicating with aserver")

	// read information on fserver
	readN, err = aConn.Read(buf[:])
	errExit(err, "Error communicating with aserver")
	checkForErrMessage(buf[0:readN])

	var fortuneInfo *FortuneInfoMessage
	err = json.Unmarshal(buf[0:readN], &fortuneInfo)
	errExit(err, "Cannot parse FortuneInfoMessage")

	err = aConn.Close()
	errExit(err, "Cannot disconnect from aserver")

	// connect to fserver, ask for fortune message
	fConn, err := lDialer.Dial("udp", fortuneInfo.FortuneServer)
	errExit(err, "Cannot connect to fserver")

	tmpBuf, err = json.Marshal(FortuneReqMessage{FortuneNonce: fortuneInfo.FortuneNonce})
	fConn.Write(tmpBuf[:])
	errExit(err, "Error communicating with fserver")

	// get fortune message
	readN, err = fConn.Read(buf[:])
	errExit(err, "Error communicating with fserver")
	checkForErrMessage(buf[0:readN])

	var fortuneMessage *FortuneMessage
	err = json.Unmarshal(buf[0:readN], &fortuneMessage)
	errExit(err, "Cannot parse FortuneMessage")

	fmt.Println(fortuneMessage.Fortune)
}

func errExit(err error, message string) {
	if err != nil {
		fmt.Println(err)
		fmt.Println(message)
		os.Exit(-1)
	}
}

func checkForErrMessage(bytes []byte) {
	var errMessage *ErrMessage
	json.Unmarshal(bytes, &errMessage)
	if errMessage.Error != "" {
		fmt.Println("Error from server: ", errMessage.Error)
		os.Exit(-1)
	}
}
