package main

import (
	"./util"
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"net/rpc"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// **********************************************
// Utility
// **********************************************
// channel that will return unavailable response once closed
type KvRequestQueue struct {
	Channel chan *KvRequest
	Closed  int32
}

func (c *KvRequestQueue) RequestAndClose(request *KvRequest) {
	atomic.StoreInt32(&c.Closed, 1)
	c.Channel <- request
}

func (c *KvRequestQueue) Request(request *KvRequest) bool {
	if atomic.CompareAndSwapInt32(&c.Closed, 0, 0) {
		c.Channel <- request
	} else {
		if request.Channel != nil {
			go func() {
				request.Channel <- &KvResponse{Unavailable: true}
			}()
		}
	}
	return true
}

func (c *KvRequestQueue) Receive() *KvRequest {
	return <-c.Channel
}

// debugging
func barfMetadata() {
	if false {
		return
	}
	metadata.RLock()
	for key, keym := range metadata.KeyMap {
		fmt.Println("****", key, keym)
	}
	for _, node := range metadata.NodeMap {
		fmt.Println("****", node)
	}
	metadata.RUnlock()
}

// **********************************************
// RPC
// **********************************************

const UNAVAILABLE = "unavailable"

type GetArgs struct {
	Key string
}

type PutArgs struct {
	Key string
	Val string
}

type TestSetArgs struct {
	Key     string
	TestVal string
	NewVal  string
}

type ValReply struct {
	Val string
}

type KeyValService int

func setResponse(key string, reply *ValReply, response *KvResponse) {
	if response.Unavailable {
		reply.Val = UNAVAILABLE
	} else {
		reply.Val = response.Values[key].Value
	}
}

func (kvs *KeyValService) Get(args *GetArgs, reply *ValReply) error {
	if args.Key[0:3] == "CMD" {
		tokens := strings.Fields(args.Key)
		// locking on metadata pretty much blocks most of the nodes at key points
		// basically we have a system to detect failure but there are delays
		// so we used a kill flag to prevent cmds from returning nodes we have already killed
		// tried another implementation that deadlocks so switched to this version
		if len(tokens) == 3 {
			if tokens[1] == "get-replicas-of" {
				var nodesList []string = make([]string, 0)
				metadata.Lock()
				key := tokens[2]
				keyMetadata, ok := metadata.KeyMap[key]
				if ok {
					nodesList = make([]string, len(keyMetadata.Nodes))
					i := 0
					for id, node := range keyMetadata.Nodes {
						if !node.Killed {
							nodesList[i] = id
							i++
						}
					}
					nodesList = nodesList[:i]
				}
				reply.Val = strings.Join(nodesList, " ")
				metadata.Unlock()
				return nil
			} else if tokens[1] == "kill-replica" {
				id := tokens[2]
				metadata.Lock()
				node, ok := metadata.NodeMap[id]
				if ok && !node.Killed {
					c := make(chan *KvResponse, 5)
					node.RequestQueue.RequestAndClose(&KvRequest{Type: KILL, Channel: c})
					<-c
					node.Killed = true
					reply.Val = "true"
				} else {
					reply.Val = "false"
				}
				metadata.Unlock()
				return nil
			}
		} else if len(tokens) == 4 {
			if tokens[1] == "kill-replicas-of" {
				key := tokens[2]
				nReplicas, err := strconv.Atoi(tokens[3])
				if err != nil {
					return err
				}
				if nReplicas < 0 {
					return errors.New("number of replicas killed should be positive")
				}
				var nodesList = make([]*NodeMetadata, 0)
				metadata.Lock()
				keyMetadata, ok := metadata.KeyMap[key]
				if ok {
					size := len(keyMetadata.Nodes)
					if size > nReplicas {
						size = nReplicas
					}
					nodesList = make([]*NodeMetadata, size)
					i := 0
					for _, node := range keyMetadata.Nodes {
						if !node.Killed {
							nodesList[i] = node
							i++
						}
						if i >= size {
							break
						}
					}
					nodesList = nodesList[:i]
					for _, node := range nodesList {
						c := make(chan *KvResponse, 5)
						node.RequestQueue.RequestAndClose(&KvRequest{Type: KILL, Channel: c})
						<-c
						node.Killed = true
					}
				}
				metadata.Unlock()
				reply.Val = strconv.Itoa(len(nodesList))
				return nil
			}
		}
	}

	c := make(chan *KvResponse, 5)
	kvReq := &KvRequest{
		Key:     args.Key,
		Keys:    make(map[string]*KeyArg),
		Channel: c,
	}
	kvReq.Keys[args.Key] = &KeyArg{Type: GET}
	requestChannel <- kvReq
	response := <-c
	setResponse(args.Key, reply, response)
	return nil
}

func (kvs *KeyValService) Put(args *PutArgs, reply *ValReply) error {
	c := make(chan *KvResponse, 5)
	kvReq := &KvRequest{
		Key:     args.Key,
		Keys:    make(map[string]*KeyArg),
		Channel: c,
	}
	kvReq.Keys[args.Key] = &KeyArg{Type: PUT, NewVal: args.Val}
	requestChannel <- kvReq
	response := <-c
	setResponse(args.Key, reply, response)
	return nil
}

func (kvs *KeyValService) TestSet(args *TestSetArgs, reply *ValReply) error {
	c := make(chan *KvResponse, 5)
	kvReq := &KvRequest{
		Key:     args.Key,
		Keys:    make(map[string]*KeyArg),
		Channel: c,
	}
	kvReq.Keys[args.Key] = &KeyArg{Type: TEST_PUT, TestVal: args.TestVal, NewVal: args.NewVal}
	requestChannel <- kvReq
	response := <-c
	setResponse(args.Key, reply, response)
	return nil
}

func startKvService(clientListener net.Listener) {
	kvservice := new(KeyValService)
	rpc.Register(kvservice)
	for {
		conn, _ := clientListener.Accept()
		go rpc.ServeConn(conn)
	}
}

// ************************************************************
// Keep track of KvNodes + Request Response Subsystem
// ************************************************************

// Listen to this to get any nodes disconnection/connection
var nodesBroadcast = make(chan *KvBroadcast, 1000)

const KV_CONNECT = "CONNECT"
const KV_DISCCONECT = "DISCONNECT"

type KvBroadcast struct {
	Type         string
	Id           string
	RequestQueue *KvRequestQueue
}

const GET = 0
const PUT = 1
const TEST_PUT = 2
const QUERY = 3
const DISCONNECTED = 4
const KILL = 5

type KeyArg struct {
	Type    int
	Version uint64
	TestVal string
	NewVal  string
}

type KvArgs struct {
	RequestId uint64
	Keys      map[string]*KeyArg
	Kill      bool
}

type KvRequest struct {
	Type    int
	Key     string
	Keys    map[string]*KeyArg
	Channel chan *KvResponse
}

type KvValue struct {
	Version uint64
	Value   string
}

type KvResponse struct {
	RequestId   uint64
	Values      map[string]*KvValue
	Unavailable bool
	Kill        bool
}

type RequestIdToChannel struct {
	sync.RWMutex
	Map map[uint64]chan<- *KvResponse
}

func startKvNodesListener(kvnodesListener net.Listener) {
	for {
		conn, _ := kvnodesListener.Accept()
		go handleKvNode(conn)
	}
}

// broadcast a node has just joined, then listen for any requests and write to the client
func handleKvNode(conn net.Conn) {
	// get id
	reader := bufio.NewReader(conn)
	id, err := util.KvReadLine(reader)
	if err != nil {
		fmt.Println(err)
		return
	}

	util.Debug(id + " connected")
	requestQueue := &KvRequestQueue{Channel: make(chan *KvRequest, 1000)}
	nodesBroadcast <- &KvBroadcast{Type: KV_CONNECT, Id: id, RequestQueue: requestQueue}
	requestIdToChan := RequestIdToChannel{Map: make(map[uint64]chan<- *KvResponse)}
	go listenKvNodeResponse(id, requestQueue, &requestIdToChan, reader)
	var requestId uint64 = 0
	writer := bufio.NewWriter(conn)
	for {
		request := requestQueue.Receive()
		if request.Type == DISCONNECTED {
			// disconnect, any channel that still need to be responded are unavailable
			util.Debug(id + " disconnected")
			nodesBroadcast <- &KvBroadcast{Type: KV_DISCCONECT, Id: id}
			for _, channel := range requestIdToChan.Map {
				channel <- &KvResponse{Unavailable: true}
			}
			return
		} else {
			// set up call back and send request
			requestIdToChan.Lock()
			requestIdToChan.Map[requestId] = request.Channel
			requestIdToChan.Unlock()
			args := KvArgs{Keys: request.Keys, RequestId: requestId}
			if request.Type == KILL {
				args = KvArgs{Keys: request.Keys, RequestId: requestId, Kill: true}
			}
			requestId++
			s, _ := json.Marshal(args)
			util.KvWriteLineSlice(writer, s)
		}
	}
}

func listenKvNodeResponse(id string, requestChan *KvRequestQueue, requestIdToChan *RequestIdToChannel,
	reader *bufio.Reader) {
	for {
		str, err := util.KvReadLineSlice(reader)
		if err != nil { // disconnected
			requestChan.RequestAndClose(&KvRequest{Type: DISCONNECTED})
			return
		}
		var response KvResponse
		json.Unmarshal(str, &response)
		util.Debug(id, response)
		requestIdToChan.Lock()
		channel := requestIdToChan.Map[response.RequestId]
		delete(requestIdToChan.Map, response.RequestId)
		requestIdToChan.Unlock()
		// well this is deadlocking, so lets just not block, there should be enough buffer space in the channels
		// note: used to deadlock* using blocking send is probably fine now
		if response.Kill {
			util.Debug(id + " killed")
		}
		select {
		case channel <- &response:
		default:
		}
	}
}

// ************************************************************
// Main system
// ************************************************************

type KeyMetadata struct {
	Version     uint64
	Unavailable bool
	Sending     int
	Nodes       map[string]*NodeMetadata
}

type NodeMetadata struct {
	Id           string
	Keys         map[string]*KeyMetadata
	RequestQueue *KvRequestQueue
	Killed       bool
}

var metadata = struct {
	sync.RWMutex
	KeyMap  map[string]*KeyMetadata
	NodeMap map[string]*NodeMetadata
}{KeyMap: make(map[string]*KeyMetadata), NodeMap: make(map[string]*NodeMetadata)}

// get nodes as list for use outside of the lock
func getNodesAsList(keyMetadata map[string]*NodeMetadata) []*NodeMetadata {
	nodesList := make([]*NodeMetadata, len(keyMetadata))
	i := 0
	for _, node := range keyMetadata {
		nodesList[i] = node
		i++
	}
	return nodesList
}

func randomizeNodes(nodesList []*NodeMetadata, upto int) {
	// fisher yates
	for i := 0; i < upto; i++ {
		k := rand.Intn(len(nodesList) - i)
		tmp := nodesList[i+k]
		nodesList[i+k] = nodesList[i]
		nodesList[i] = tmp
	}
}

func prepareNodesForFirstReplication(key string, replicationFactor int) ([]*NodeMetadata, *KeyMetadata) {
	// get random nodes that we want to replicate to
	nodesList := getNodesAsList(metadata.NodeMap)
	upto := len(nodesList)
	if upto > replicationFactor {
		upto = replicationFactor
	}
	randomizeNodes(nodesList, upto)
	nodesList = nodesList[:upto]

	// put them in metadata
	keyMetadata := &KeyMetadata{Nodes: make(map[string]*NodeMetadata), Version: 0}
	metadata.KeyMap[key] = keyMetadata
	for _, node := range nodesList {
		metadata.NodeMap[node.Id].Keys[key] = metadata.KeyMap[key]
		keyMetadata.Nodes[node.Id] = node
	}
	return nodesList, keyMetadata
}

// this takes the channel and find the first unavailable response, if they are all unavailable, respond unavailable to RPC
func filterChannelUnavailable(upto int, channel chan *KvResponse, keyMetadata *KeyMetadata) chan *KvResponse {
	newChannel := make(chan *KvResponse, upto+5)
	go func() {
		for i := 0; i < upto; i++ {
			response := <-newChannel
			if !response.Unavailable {
				metadata.RLock()
				unavailable := keyMetadata.Unavailable
				keyMetadata.Sending--
				metadata.RUnlock()
				if unavailable {
					channel <- &KvResponse{Unavailable: true}
				} else {
					channel <- response
				}
				return
			}
		}
		metadata.Lock()
		keyMetadata.Unavailable = true
		keyMetadata.Sending--
		metadata.Unlock()
		channel <- &KvResponse{Unavailable: true}
	}()
	return newChannel
}

var requestChannel = make(chan *KvRequest, 10000)

func mainRoutine(replicationFactor int) {
	for {
		request := <-requestChannel

		metadata.Lock()
		// key is not in map = new key, then initial replication
		var nodes []*NodeMetadata
		keyMetadata := metadata.KeyMap[request.Key]
		if keyMetadata == nil {
			nodes, keyMetadata = prepareNodesForFirstReplication(request.Key, replicationFactor)
		} else {
			nodes = getNodesAsList(keyMetadata.Nodes)
		}
		unavailable := keyMetadata.Unavailable
		keyMetadata.Sending++
		metadata.Unlock()

		// key is in map but there's no nodes with it = unavailable
		if unavailable || len(nodes) == 0 {
			if !unavailable {
				metadata.Lock()
				keyMetadata.Unavailable = true
				keyMetadata.Sending--
				metadata.Unlock()
			}
			go func() {
				request.Channel <- &KvResponse{Unavailable: true}
			}()
		} else {
			// otherwise send request to all nodes, update version if put/testset
			request.Channel = filterChannelUnavailable(len(nodes), request.Channel, keyMetadata)
			keyArg := request.Keys[request.Key]
			if keyArg.Type == PUT || keyArg.Type == TEST_PUT {
				metadata.Lock()
				keyMetadata.Version++
				keyArg.Version = keyMetadata.Version
				metadata.Unlock()
			}
			for _, requestQueue := range nodes {
				requestQueue.RequestQueue.Request(request)
			}
		}
	}
}

func nodesRoutine(replicationFactor int) {
	for {
		nodeInfo := <-nodesBroadcast
		if nodeInfo.Type == KV_CONNECT {
			// add new node
			metadata.Lock()
			nodeMetadata := &NodeMetadata{
				Id:           nodeInfo.Id,
				Keys:         make(map[string]*KeyMetadata),
				RequestQueue: nodeInfo.RequestQueue}
			metadata.NodeMap[nodeInfo.Id] = nodeMetadata
			numberOfNodes := len(metadata.NodeMap)
			metadata.Unlock()

			// replicate all the keys
			if numberOfNodes <= replicationFactor {
				metadata.RLock()
				keys := make(map[string]struct{})
				for key, keyMetadata := range metadata.KeyMap {
					if len(keyMetadata.Nodes) > 0 {
						keys[key] = struct{}{}
					}
				}
				metadata.RUnlock()
				go replicationRoutine(nodeMetadata, keys)
			}
		} else if nodeInfo.Type == KV_DISCCONECT {
			// remove nodes and associated node from the keys
			metadata.Lock()
			nodeMetadata := metadata.NodeMap[nodeInfo.Id]
			for key, _ := range nodeMetadata.Keys {
				delete(metadata.KeyMap[key].Nodes, nodeInfo.Id)
			}
			delete(metadata.NodeMap, nodeInfo.Id)
			metadata.Unlock()

			// get all keys, partition and replicate
			keys := make(map[string]struct{})
			for key, _ := range nodeMetadata.Keys {
				keys[key] = struct{}{}
			}
			partitionAndReplicate(keys)
		}
	}
}

// ************************************************************
// Replication
// ************************************************************

func partitionAndReplicate(keys map[string]struct{}) {
	metadata.RLock()
	nodesList := getNodesAsList(metadata.NodeMap)
	randomizeNodes(nodesList, len(nodesList))

	for _, node := range nodesList {
		subkeys := make(map[string]struct{})
		for key, _ := range keys {
			_, ok := node.Keys[key]
			if !ok {
				subkeys[key] = struct{}{}
				delete(keys, key)
			}
		}
		if len(subkeys) > 0 {
			go replicationRoutine(node, subkeys)
		}
		if len(keys) == 0 {
			break
		}
	}
	metadata.RUnlock()
}

func replicationRoutine(node *NodeMetadata, keys map[string]struct{}) {
	for len(keys) > 0 {
		// get all nodes which has the key we need to replicate
		nodesMap := make(map[string]*NodeMetadata)
		metadata.RLock()
		for key, _ := range keys {
			for id, _ := range metadata.KeyMap[key].Nodes {
				nodesMap[id] = metadata.NodeMap[id]
			}
		}
		metadata.RUnlock()

		// ask each node what they have
		responseChan := make(chan *KvResponse, len(nodesMap)+5)
		request := &KvRequest{Keys: make(map[string]*KeyArg), Channel: responseChan}
		for key, _ := range keys {
			request.Keys[key] = &KeyArg{Type: QUERY}
		}

		for _, queryNode := range nodesMap {
			queryNode.RequestQueue.Request(request)
		}

		// and just put the responses in the target node
		requestKeys := mergeReplicatingQuery(len(nodesMap), len(keys), responseChan)
		responseChan = make(chan *KvResponse, 5)
		newRequest := &KvRequest{Keys: requestKeys, Channel: responseChan}
		node.RequestQueue.Request(newRequest)
		response := <-responseChan

		// unavailable = node failed, then replicate rest of key in other nodes
		if response.Unavailable {
			partitionAndReplicate(keys)
			return
		}

		// let's see if we managed to replicate the correct values and update any metadata
		metadata.Lock()
		for key, _ := range keys {
			responseVal, ok := response.Values[key]
			keyMetadata := metadata.KeyMap[key]
			if len(keyMetadata.Nodes) == 0 {
				delete(keys, key)
			} else if ok && !keyMetadata.Unavailable && keyMetadata.Sending == 0 && responseVal.Version == keyMetadata.Version {
				// describe in design doc
				delete(keys, key)
				keyMetadata.Nodes[node.Id] = node
				node.Keys[key] = metadata.KeyMap[key]
			}
		}
		metadata.Unlock()
	}
}

// this get responses from each node and merging any keyvalue into a bunch of PUT key arguments
func mergeReplicatingQuery(nodesUpto int, keysUpto int, channel chan *KvResponse) map[string]*KeyArg {
	keyMap := make(map[string]*KeyArg)
	for i := 0; i < nodesUpto; i++ {
		request := <-channel
		if request.Unavailable {
			continue
		}
		for key, value := range request.Values {
			keyArg, ok := keyMap[key]
			if !ok {
				keyMap[key] = &KeyArg{Type: PUT, Version: value.Version, NewVal: value.Value}
			} else if keyArg.Version < value.Version {
				keyMap[key].Version = value.Version
				keyMap[key].NewVal = value.Value
			}
		}
		// try to replicate if we have all the values before waiting for other nodes to finish
		if len(keyMap) >= keysUpto {
			return keyMap
		}
	}
	return keyMap
}

func main() {
	if len(os.Args) < 4 {
		fmt.Fprintf(os.Stderr, "Usage: go run kv-front-end.go [client ip:port] [kv-node ip:port] [r]\n", os.Args[0])
		os.Exit(1)
	}

	rand.Seed(time.Now().UTC().UnixNano())

	clientListener, err := net.Listen("tcp", os.Args[1])
	util.CheckError(err)
	go startKvService(clientListener)

	replicationFactorU, err := strconv.ParseUint(os.Args[3], 10, 32)
	util.CheckError(err)
	replicationFactor := int(replicationFactorU)

	go mainRoutine(replicationFactor)
	go nodesRoutine(replicationFactor)

	kvnodesListener, err := net.Listen("tcp", os.Args[2])
	util.CheckError(err)
	startKvNodesListener(kvnodesListener)
}
