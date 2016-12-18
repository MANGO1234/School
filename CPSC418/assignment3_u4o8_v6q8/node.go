package main

import (
	"fmt"
	"net/rpc"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Utility methods for key value service

const UNAVAILABLE string = "unavailable"

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

var kvService *rpc.Client

func initKeyValueService(addr string) error {
	var err error
	kvService, err = rpc.Dial("tcp", os.Args[1])
	return err
}

func getVal(key string) (string, error) {
	var kvVal ValReply
	args := GetArgs{Key: key}
	err := kvService.Call("KeyValService.Get", args, &kvVal)
	return kvVal.Val, err
}

func putVal(key string, value string) (string, error) {
	var kvVal ValReply
	args := PutArgs{Key: key, Val: value}
	err := kvService.Call("KeyValService.Put", args, &kvVal)
	return kvVal.Val, err
}

func testPutVal(key string, test string, value string) (string, error) {
	var kvVal ValReply
	args := TestSetArgs{Key: key, TestVal: test, NewVal: value}
	err := kvService.Call("KeyValService.TestSet", args, &kvVal)
	return kvVal.Val, err
}

// this may be overkill for correctness
type NonceStruct struct {
	Nonce uint64
	Mutex sync.RWMutex
}

var nonceMap = make(map[string]*NonceStruct)
var nonceMapMutex = &sync.Mutex{}

func getNonceAndLock(key string) uint64 {
	nonceMapMutex.Lock()
	nonce, ok := nonceMap[key]
	if !ok {
		nonce = &NonceStruct{Nonce: 0, Mutex: sync.RWMutex{}}
		nonceMap[key] = nonce
	}
	nonceMapMutex.Unlock()
	nonce.Mutex.Lock()
	v := nonce.Nonce
	return v
}

func setNonceAndUnlock(key string, value uint64) {
	nonceMapMutex.Lock()
	nonce := nonceMap[key]
	nonceMapMutex.Unlock()
	nonce.Nonce = value
	nonce.Mutex.Unlock()
}

// These versions deals with unavailability by incrementing a nonce
func getValU(key string) (string, error) {
	key = key + " "
	currentNonce := getNonceAndLock(key)
	result, err := getVal(key + uInt64ToStr(currentNonce))
	for result == UNAVAILABLE && err == nil {
		currentNonce++
		result, err = getVal(key + uInt64ToStr(currentNonce))
	}
	setNonceAndUnlock(key, currentNonce)
	return result, err
}

func putValU(key string, val string) (string, error) {
	key = key + " "
	currentNonce := getNonceAndLock(key)
	result, err := putVal(key+uInt64ToStr(currentNonce), val)
	for result == UNAVAILABLE && err == nil {
		currentNonce++
		result, err = putVal(key+uInt64ToStr(currentNonce), val)
	}
	setNonceAndUnlock(key, currentNonce)
	return result, err
}

func testPutValU(key string, test string, val string) (string, error) {
	key = key + " "
	currentNonce := getNonceAndLock(key)
	result, err := testPutVal(key+uInt64ToStr(currentNonce), test, val)
	for result == UNAVAILABLE && err == nil {
		currentNonce++
		result, err = testPutVal(key+uInt64ToStr(currentNonce), test, val)
	}
	setNonceAndUnlock(key, currentNonce)
	return result, err
}

// **********************************************
// Utility
// **********************************************

var debugOn = false

func debug(v interface{}) {
	if debugOn {
		fmt.Println(v)
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error ", err.Error())
		os.Exit(1)
	}
}

func uInt64ToStr(n uint64) string {
	return strconv.FormatUint(n, 10)
}

func removeId(ids []string, id string) []string {
	for i, v := range ids {
		if v == id {
			return append(ids[:i], ids[i+1:]...)
		}
	}
	return ids
}

func hasId(ids []string, id string) bool {
	for _, v := range ids {
		if v == id {
			return true
		}
	}
	return false
}

var ID = os.Args[2]

// **********************************************
// Leader
// **********************************************

const LEADER = "leader key"
const LEADER_KEY = "leader key "

type LeaderVal struct {
	Era     uint64
	Nonce   uint64
	Version uint64
	Id      string
	Nodes   []string
}

func LeaderValToString(leader LeaderVal) string {
	return uInt64ToStr(leader.Era) + " " + uInt64ToStr(leader.Nonce) + " " + uInt64ToStr(leader.Version) + " " + leader.Id + " " + strings.Join(leader.Nodes, " ")
}

func StringToLeaderVal(str string) LeaderVal {
	result := strings.Split(str, " ")
	era, _ := strconv.ParseUint(result[0], 10, 64)
	nonce, _ := strconv.ParseUint(result[1], 10, 64)
	pong, _ := strconv.ParseUint(result[2], 10, 64)
	if result[4] == "" {
		return LeaderVal{Era: era, Nonce: nonce, Version: pong, Id: result[3], Nodes: make([]string, 0)}
	} else {
		return LeaderVal{Era: era, Nonce: nonce, Version: pong, Id: result[3], Nodes: result[4:]}
	}
}

func printActive(leader LeaderVal) {
	fmt.Print(leader.Id)
	fmt.Print(" ")
	fmt.Println(strings.Join(leader.Nodes, " "))
}

// this can be used to see if leader needs to be updated a > b means a is newer, a = b means they are equal
func compareLeaders(a LeaderVal, b LeaderVal) int {
	if a.Era < b.Era {
		return -1
	} else if a.Era > b.Era {
		return 1
	} else {
		if a.Nonce < b.Nonce {
			return 1
		} else if a.Nonce > b.Nonce {
			return -1
		} else {
			if a.Version < b.Version {
				return -1
			} else if a.Version > b.Version {
				return 1
			} else {
				return 0
			}
		}
	}
}

var currentLeader LeaderVal
var currentLeaderMutex = &sync.RWMutex{}

// at the beginning, try to make myself the leader, or if there's already a leader, record it
func initLeader() {
	currentLeaderNonce := getNonceAndLock(LEADER_KEY)
	leaderRequest := LeaderVal{Era: 0, Nonce: currentLeaderNonce, Version: 0, Id: ID, Nodes: make([]string, 0)}
	result, err := testPutVal(LEADER_KEY+uInt64ToStr(currentLeaderNonce), "", LeaderValToString(leaderRequest))
	checkError(err)
	for result == UNAVAILABLE {
		currentLeaderNonce++
		leaderRequest.Nonce = currentLeaderNonce
		result, err = testPutVal(LEADER_KEY+uInt64ToStr(currentLeaderNonce), "", LeaderValToString(leaderRequest))
		checkError(err)
	}
	setNonceAndUnlock(LEADER_KEY, currentLeaderNonce)
	currentLeader = StringToLeaderVal(result)

	// leader can actually now come back fom previous session if needed
	if currentLeader.Id == ID {
		for _, id := range currentLeader.Nodes {
			idToChannel[id] = make(chan string)
			go pongRoutine(id, leaderRequest.Id, idToChannel[id])
		}
	}
}

// try to make myself the leader, or if there's already a leader, record it
func becomeOrUpdateLeader(lastLeader LeaderVal) {
	currentLeaderNonce := getNonceAndLock(LEADER_KEY)
	newNodes := removeId(lastLeader.Nodes, ID)
	leaderRequest := LeaderVal{Era: lastLeader.Era + 1, Nonce: currentLeaderNonce, Version: 0, Id: ID, Nodes: newNodes}

	var result string = ""
	var err error
	for {
		request := LeaderValToString(leaderRequest)
		result, err = testPutVal(LEADER_KEY+uInt64ToStr(currentLeaderNonce), result, request)
		checkError(err)

		if result == request {
			// pong any active nodes to help speed up
			for _, id := range newNodes {
				idToChannel[id] = make(chan string)
				go pongRoutine(id, leaderRequest.Id, idToChannel[id])
			}
			break
		} else if result == UNAVAILABLE {
			currentLeaderNonce++
			leaderRequest.Nonce = currentLeaderNonce
		} else if result != "" {
			leader := StringToLeaderVal(result)
			if compareLeaders(leader, lastLeader) > 0 {
				break
			}
		}
	}

	setNonceAndUnlock(LEADER_KEY, currentLeaderNonce)
	currentLeader = StringToLeaderVal(result)
}

// try to update leader (only used when I'm follower)
func updateLeader() {
	leaderVal, err := getValU(LEADER)
	checkError(err)
	if leaderVal != "" {
		leader := StringToLeaderVal(leaderVal)
		if compareLeaders(leader, currentLeader) >= 0 {
			currentLeader = leader
		}
	}
}

// try to update leader (only used when I'm leader), will follow a new leader if detected
func updateLeaderValue(leader LeaderVal) LeaderVal {
	result := LeaderValToString(currentLeader)
	target := LeaderValToString(leader)
	var err error
	for {
		result, err = testPutValU(LEADER, result, target)
		checkError(err)
		if result == target {
			return leader
		} else if result != "" {
			resultLeader := StringToLeaderVal(result)
			if compareLeaders(resultLeader, currentLeader) > 0 {
				return resultLeader
			}
		}
	}
}

var idToChannel = make(map[string]chan string)

func joinRoutine(currentId string, msg chan<- string) {
	for {
		currentLeaderMutex.RLock()
		leader := currentLeader
		currentLeaderMutex.RUnlock()

		if leader.Id != currentId {
			return
		}

		joinValue := ""
		for {
			result, err := testPutValU(JOIN, joinValue, "")
			checkError(err)
			if result == "" {
				break
			}
			joinValue = result
		}
		debug("Join: " + joinValue)
		if joinValue != "" {
			msg <- joinValue
		} else {
			time.Sleep(JOIN_INTERVAL)
		}
	}
}

func leaderRoutine(currentId string, joinChan <-chan string) {
	for {
		currentLeaderMutex.RLock()
		leader := currentLeader
		currentLeaderMutex.RUnlock()

		if leader.Id != currentId {
			return
		}

		// remove any nodes that disconnected
		for id, msgs := range idToChannel {
			select {
			case _ = <-msgs:
				leader.Nodes = removeId(leader.Nodes, id)
				delete(idToChannel, id)
			default:
			}
		}

		// check for request to join nodes
		select {
		case joinAddr := <-joinChan:
			// added any nodes to active list and pong it
			if joinAddr != "" && joinAddr != UNAVAILABLE {
				joinNodes := strings.Split(joinAddr, " ")
				for _, id := range joinNodes {
					// when key becomes unavailable this is needed to prevent duplicate
					if !hasId(leader.Nodes, id) {
						idToChannel[id] = make(chan string)
						go pongRoutine(id, leader.Id, idToChannel[id])
						leader.Nodes = append(leader.Nodes, id)
					}
				}
			}
		default:
		}

		leader.Version++
		updatedLeader := updateLeaderValue(leader)

		currentLeaderMutex.Lock()
		currentLeader = updatedLeader
		currentLeaderMutex.Unlock()
		debug(currentLeader)
		printActive(currentLeader)

		time.Sleep(LEADER_UPDATE_INTERVAL)
	}
}

const JOIN = "join key"
const PING = "ping"
const PONG = "pong"
const PING_INTERVAL = 400 * time.Millisecond
const PONG_INTERVAL = 400 * time.Millisecond
const LEADER_UPDATE_INTERVAL = 800 * time.Millisecond
const FOLLOWER_UPDATE_INTERVAL = 800 * time.Millisecond
const JOIN_INTERVAL = 500 * time.Millisecond

// ping and pong just ping and pong each other to see if each other are alive
// server use this to detect any failed node
func pongRoutine(id string, currentId string, msg chan<- string) {
	for {
		currentLeaderMutex.RLock()
		leader := currentLeader
		currentLeaderMutex.RUnlock()

		if leader.Id != currentId {
			return
		}

		retry := 0
		for {
			debug("pong " + id)
			str, err := getValU(id)
			checkError(err)
			if str != PING && retry >= 5 {
				msg <- "DISCONNECTED"
				return
			} else if str == PING || str == "" {
				_, err := putValU(id, PONG)
				checkError(err)
				break
			} else {
				retry++
				time.Sleep(PONG_INTERVAL)
			}
		}
		time.Sleep(PONG_INTERVAL)
	}
}

func pingRoutine(msg chan<- string) {
	for {
		currentLeaderMutex.RLock()
		leader := currentLeader
		currentLeaderMutex.RUnlock()

		if leader.Id == ID {
			return
		}

		retry := 0
		for {
			debug("ping " + ID)
			str, err := getValU(ID)
			checkError(err)
			if str != PONG && retry >= 5 {
				msg <- "DISCONNECTED"
				break
			} else if str == PONG || str == "" {
				_, err := putValU(ID, PING)
				checkError(err)
				break
			} else {
				retry++
				time.Sleep(PING_INTERVAL)
			}
			time.Sleep(PING_INTERVAL)
		}
	}
}

// basically ping -> update leader -> ping -> update leader ... , if server doesn't pong try to update leader or become new leader
func followerRoutine(msg <-chan string) {
	for {
		currentLeaderMutex.RLock()
		leader := currentLeader
		currentLeaderMutex.RUnlock()

		if leader.Id == ID {
			return
		}

		select {
		case _ = <-msg:
			// if pinging stops, assume leader has failed and try to become new leader
			becomeOrUpdateLeader(currentLeader)
		default:
			updateLeader()
			if !hasId(currentLeader.Nodes, ID) {
				newJoinAddr := ID
				result, err := testPutValU(JOIN, "", newJoinAddr)
				checkError(err)
				for result != newJoinAddr {
					if result == "" {
						newJoinAddr = ID
					} else {
						newJoinAddr = result + " " + ID
					}
					result, err = testPutValU(JOIN, result, newJoinAddr)
					checkError(err)
				}
			}
			debug(currentLeader)
			printActive(currentLeader)
		}

		time.Sleep(FOLLOWER_UPDATE_INTERVAL)
	}
}

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: go run node.go [ip:port] [id]\n", os.Args[0])
		os.Exit(1)
	}

	initKeyValueService(os.Args[1])
	initLeader()

	for {
		if currentLeader.Id == ID {
			var joinToLeader = make(chan string)
			go joinRoutine(currentLeader.Id, joinToLeader)
			leaderRoutine(currentLeader.Id, joinToLeader)
		} else {
			var pingToFollower = make(chan string)
			go pingRoutine(pingToFollower)
			followerRoutine(pingToFollower)
		}
	}
}
