package main

import (
	"./util"
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
)

const GET = 0
const PUT = 1
const TEST_PUT = 2
const QUERY = 3

type KeyArg struct {
	Type    int
	Version uint64
	TestVal string
	NewVal  string
}

type KvArgs struct {
	RequestId uint64
	Keys      map[string]KeyArg
	Kill      bool
}

type KvValue struct {
	Version uint64
	Value   string
}

type KvResponse struct {
	RequestId uint64
	Values    map[string]KvValue
	Kill      bool
}

func main() {
	if len(os.Args) < 4 {
		fmt.Fprintf(os.Stderr, "Usage: go run kv-node.go [local ip] [front-end ip:port] [id]\n", os.Args[0])
		os.Exit(1)
	}

	localAddr, err := net.ResolveTCPAddr("tcp", os.Args[1]+":0")
	util.CheckError(err)
	frontendAddr, err := net.ResolveTCPAddr("tcp", os.Args[2])
	util.CheckError(err)
	id := os.Args[3]

	frontendConn, err := net.DialTCP("tcp", localAddr, frontendAddr)
	util.CheckError(err)

	writer := bufio.NewWriter(frontendConn)
	err = util.KvWriteLine(writer, id)
	util.CheckError(err)

	writeChannel := make(chan *KvResponse, 1000)
	go writeRoutine(writer, writeChannel)

	kvmap := make(map[string]KvValue)
	reader := bufio.NewReader(frontendConn)
	for {
		request, err := util.KvReadLineSlice(reader)
		util.CheckError(err)
		handleRequest(kvmap, request, writeChannel)
	}
}

func handleRequest(kvmap map[string]KvValue, request []byte, writeChannel chan<- *KvResponse) {
	var arg KvArgs
	err := json.Unmarshal(request, &arg)
	util.CheckError(err) // should never happen

	if arg.Kill {
		writeChannel <- &KvResponse{RequestId: arg.RequestId, Kill: true}
		return
	}

	responseMap := make(map[string]KvValue)
	for key, keyarg := range arg.Keys {
		if keyarg.Type == GET {
			value, ok := kvmap[key]
			responseMap[key] = value
			if !ok {
				kvmap[key] = KvValue{Value: "", Version: 0}
			}
		} else if keyarg.Type == QUERY {
			responseMap[key] = kvmap[key]
		} else if keyarg.Type == PUT {
			if kvmap[key].Version < keyarg.Version {
				kvmap[key] = KvValue{Value: keyarg.NewVal, Version: keyarg.Version}
			}
			value, ok := kvmap[key]
			responseMap[key] = KvValue{Value: "", Version: value.Version}
			if !ok {
				kvmap[key] = KvValue{Value: "", Version: 0}
			}
		} else if keyarg.Type == TEST_PUT {
			if kvmap[key].Version < keyarg.Version {
				if kvmap[key].Value == keyarg.TestVal {
					kvmap[key] = KvValue{Value: keyarg.NewVal, Version: keyarg.Version}
				} else {
					kvmap[key] = KvValue{Value: kvmap[key].Value, Version: keyarg.Version}
				}
			}
			value, ok := kvmap[key]
			responseMap[key] = value
			if !ok {
				kvmap[key] = KvValue{Value: "", Version: 0}
			}
		}
	}
	util.Debug(arg)
	util.Debug(kvmap)

	writeChannel <- &KvResponse{RequestId: arg.RequestId, Values: responseMap}
}

func writeRoutine(writer *bufio.Writer, writeChannel <-chan *KvResponse) {
	for {
		response := <-writeChannel
		str, _ := json.Marshal(response)
		util.KvWriteLineSlice(writer, str)
		if response.Kill {
			os.Exit(-1)
		}
	}
}
