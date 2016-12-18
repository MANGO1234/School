package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"net/rpc"
	"os"
	"sync"
	"time"
)

type ErrMessage struct {
	Error string
}

type FortuneReqMessage struct {
	FortuneNonce int64
}

type FortuneMessage struct {
	Fortune string
}

type FortuneServerRPC struct{}

type FortuneInfoMessage struct {
	FortuneServer string
	FortuneNonce  int64
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	udpIp := os.Args[2]
	rpcIP := os.Args[1]

	// start listening on rpc
	rpcListener, err := net.Listen("tcp", rpcIP)
	errExit(err, "Cannot listen at "+rpcIP)
	rpc.Register(new(FortuneServerRPC))
	go func() {
		rpc.Accept(rpcListener)
	}()

	udpAddr, err := net.ResolveUDPAddr("udp", udpIp)
	errExit(err, "Cannot resolve "+udpIp)
	udpListener, err := net.ListenUDP("udp", udpAddr)
	errExit(err, "Cannot listen at "+udpIp)

	for {
		var buf [1024]byte
		nRead, addr, err := udpListener.ReadFrom(buf[:])
		go handleFortuneRequest(udpListener, buf[0:nRead], addr, err)
	}
}

var addrNonceMap = struct {
	sync.RWMutex
	m map[string]int64
}{m: make(map[string]int64)}

func (this *FortuneServerRPC) GetFortuneInfo(clientAddr string, fInfoMsg *FortuneInfoMessage) error {
	fInfoMsg.FortuneServer = os.Args[2]
	fInfoMsg.FortuneNonce = rand.Int63()
	addrNonceMap.Lock()
	addrNonceMap.m[clientAddr] = fInfoMsg.FortuneNonce
	addrNonceMap.Unlock()
	return nil
}

func handleFortuneRequest(udpListener *net.UDPConn, request []byte, addr net.Addr, err error) {
	if err == nil {
		fmt.Println("Request from " + addr.String())
		err = handleFortuneRequestHelper(udpListener, request, addr)
	}
	if err != nil {
		fmt.Print("Error for " + addr.String() + ": ")
		fmt.Println(err)
	}
}

func handleFortuneRequestHelper(udpListener *net.UDPConn, request []byte, addr net.Addr) error {
	// read fortune request
	var fortuneReq *FortuneReqMessage
	err := json.Unmarshal(request, &fortuneReq)
	if err != nil {
		return sendUdpMessage(udpListener, addr, ErrMessage{"could not interpret message"})
	}

	// check nonce
	addrNonceMap.RLock()
	nonce, ok := addrNonceMap.m[addr.String()]
	addrNonceMap.RUnlock()
	if !ok {
		return sendUdpMessage(udpListener, addr, ErrMessage{"unknown remote client address"})
	}
	if nonce != fortuneReq.FortuneNonce {
		return sendUdpMessage(udpListener, addr, ErrMessage{"incorrect fortune nonce"})
	}

	// send fortune
	fortune := os.Args[3]
	return sendUdpMessage(udpListener, addr, FortuneMessage{Fortune: fortune})
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
