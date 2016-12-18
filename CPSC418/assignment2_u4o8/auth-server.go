package main

import (
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"net/rpc"
	"os"
	"strconv"
	"sync"
	"time"
)

type ErrMessage struct {
	Error string
}

type NonceMessage struct {
	Nonce int64
}

type HashMessage struct {
	Hash string
}

type FortuneInfoMessage struct {
	FortuneServer string
	FortuneNonce  int64
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	udpIp := os.Args[1]
	rpcIp := os.Args[2]
	secret, err := strconv.ParseInt(os.Args[3], 10, 64)
	errExit(err, "Secret is not an int64")

	udpAddr, err := net.ResolveUDPAddr("udp", udpIp)
	errExit(err, "Cannot resolve "+udpIp)
	udpListener, err := net.ListenUDP("udp", udpAddr)
	errExit(err, "Cannot listen at "+udpIp)

	rpcClient, err := rpc.Dial("tcp", rpcIp)
	errExit(err, "Cannot listen at "+rpcIp)

	for {
		var buf [1024]byte
		nRead, addr, err := udpListener.ReadFrom(buf[:])
		go handleAuthRequest(udpListener, addr, buf[0:nRead], err, rpcClient, secret)
	}
}

func handleAuthRequest(udpListener *net.UDPConn, addr net.Addr, request []byte, err error, rpcClient *rpc.Client, secret int64) {
	if err == nil {
		fmt.Println("Request from " + addr.String())
		err = handleAuthRequestHelper(udpListener, addr, request, rpcClient, secret)
	}
	if err != nil {
		fmt.Print("Error for " + addr.String() + ": ")
		fmt.Println(err)
	}
}

var addrNonceMap = struct {
	sync.RWMutex
	m map[string]int64
}{m: make(map[string]int64)}

func handleAuthRequestHelper(udpListener *net.UDPConn, addr net.Addr, request []byte, rpcClient *rpc.Client, secret int64) error {
	var hashMessage *HashMessage
	err := json.Unmarshal(request, &hashMessage)
	if err == nil {
		// check if hash is wrong and send correct error message
		errMessage := checkHash(addr, hashMessage, secret)
		if errMessage != nil {
			return sendUdpMessage(udpListener, addr, errMessage)
		}

		// try calling rpc
		fortuneInfo := new(FortuneInfoMessage)
		err = rpcClient.Call("FortuneServerRPC.GetFortuneInfo", addr.String(), &fortuneInfo)
		if err != nil {
			return err
		}

		// send back to client
		return sendUdpMessage(udpListener, addr, fortuneInfo)
	} else {
		// generate new nonce
		var nonce int64 = rand.Int63()
		addrNonceMap.Lock()
		addrNonceMap.m[addr.String()] = nonce
		addrNonceMap.Unlock()
		return sendUdpMessage(udpListener, addr, NonceMessage{Nonce: nonce})
	}
}

func checkHash(addr net.Addr, hashMessage *HashMessage, secret int64) *ErrMessage {
	addrNonceMap.RLock()
	nonce, ok := addrNonceMap.m[addr.String()]
	addrNonceMap.RUnlock()
	if !ok {
		return &ErrMessage{Error: "unknown remote client address"}
	}
	if computeNonceSecretHash(nonce, secret) != hashMessage.Hash {
		return &ErrMessage{Error: "unexpected hash value"}
	}
	return nil
}

func computeNonceSecretHash(nonce int64, secret int64) string {
	sum := nonce + secret
	buf := make([]byte, 512)
	n := binary.PutVarint(buf, sum)
	h := md5.New()
	h.Write(buf[:n])
	str := hex.EncodeToString(h.Sum(nil))
	return str
}

func sendUdpMessage(conn *net.UDPConn, addr net.Addr, message interface{}) error {
	tmpBuf, err := json.Marshal(message)
	if err != nil {
		return err
	}
	_, err = conn.WriteTo(tmpBuf, addr)
	return err
}

func errExit(err error, message string) {
	if err != nil {
		fmt.Println(err)
		fmt.Println(message)
		os.Exit(-1)
	}
}
