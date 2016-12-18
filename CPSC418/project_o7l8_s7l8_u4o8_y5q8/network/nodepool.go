package network

import (
	"../util"
	"github.com/arcaneiceman/GoVector/govec"
	"net"
	"sync"
	"time"
)

// node state
type NodeState int

const (
	nodeStateDisconnected NodeState = iota
	nodeStateConnected
	nodeStateLeft
	nodeStateSessionEnded
)

const chanBufferSize = 30

type node struct {
	stateMutex sync.Mutex
	state      NodeState
	id         string
	addr       string
	outChan    chan Message
	conn       net.Conn
	reader     *util.MessageReader
	writer     *util.MessageWriter
	interval   time.Duration // current interval to reconnect
	logger     *govec.GoLog
}

func (n *node) setState(state NodeState) bool {
	n.stateMutex.Lock()
	defer n.stateMutex.Unlock()
	switch n.state {
	case nodeStateLeft:
		return state == nodeStateLeft
	case nodeStateSessionEnded:
		if state == nodeStateLeft {
			n.state = nodeStateLeft
			return true
		}
		return state == nodeStateSessionEnded
	default:
		n.state = state
		return true
	}
}

type nodePool struct {
	netMetaMutex sync.RWMutex
	netMeta      NetMeta
	poolMutex    sync.RWMutex
	pool         map[string]*node
	logger       *govec.GoLog
}

func newNodePool(logger *govec.GoLog) *nodePool {
	var np nodePool
	np.netMeta = newNetMeta()
	np.pool = make(map[string]*node)
	np.logger = logger
	return &np
}

func (np *nodePool) sendMessageToNodeWithId(msg Message, id string) bool {
	if n, ok := np.getNodeWithId(id); ok {
		return putMsgOnSendingQueueOfNode(msg, n)
	}
	return false
}

func (np *nodePool) broadcast(msg Message) {
	if msg.Visited == nil {
		np.broadcastOnce(msg)
	} else {
		np.broadcastRecursive(msg)
	}
}

func (np *nodePool) broadcastOnce(msg Message) {
	connected := np.getConnectedNodes()
	for _, n := range connected {
		putMsgOnSendingQueueOfNode(msg, n)
	}
}

func (np *nodePool) broadcastRecursive(msg Message) {
	connected := np.getConnectedNodes()
	// copy visited nodes
	original := map[string]struct{}{}
	for id, _ := range msg.Visited {
		original[id] = struct{}{}
	}
	// add connected nodes to visited
	for _, n := range connected {
		msg.Visited[n.id] = struct{}{}
	}
	for _, n := range connected {
		if _, ok := original[n.id]; !ok {
			successful := putMsgOnSendingQueueOfNode(msg, n)
			if !successful {
				delete(msg.Visited, n.id)
			}
		}
	}
}

// returns whether putting on the queue is successful
func putMsgOnSendingQueueOfNode(msg Message, n *node) bool {
	// if the channel buffer is full we just drop it
	// there's no point in sending way more than the node can handle
	select {
	case n.outChan <- msg:
		return true
	default: // dumps the msg if buffer is full
		return false
	}
}

func (np *nodePool) has(id string) bool {
	_, ok := np.netMeta[id]
	return ok
}

func (np *nodePool) handleNewSession(s *session) {
	np.netMetaMutex.Lock()
	np.netMeta[s.id] = NodeMeta{s.manager.publicAddr, false}
	np.netMetaMutex.Unlock()
	np.poolMutex.RLock()
	for _, n := range np.pool {
		n.stateMutex.Lock()
		if n.state == nodeStateSessionEnded {
			n.state = nodeStateDisconnected
		}
		n.stateMutex.Unlock()
		s.initiateNewNode(n)
	}
	np.poolMutex.RUnlock()
}

func (np *nodePool) handleEndSession(s *session) {
	np.netMetaMutex.Lock()
	np.netMeta[s.id] = NodeMeta{s.manager.publicAddr, true}
	np.netMetaMutex.Unlock()
	np.poolMutex.RLock()
	for _, n := range np.pool {
		n.setState(nodeStateSessionEnded)
	}
	np.poolMutex.RUnlock()
}

func (np *nodePool) getConnectedNodes() []*node {
	np.poolMutex.RLock()
	defer np.poolMutex.RUnlock()
	nodes := make([]*node, 0)
	for _, n := range np.pool {
		if n.state == nodeStateConnected {
			nodes = append(nodes, n)
		}
	}
	return nodes
}

func (np *nodePool) getNodeWithId(id string) (*node, bool) {
	np.poolMutex.RLock()
	n, ok := np.pool[id]
	np.poolMutex.RUnlock()
	return n, ok
}

func (np *nodePool) getAllNodes() []*node {
	np.poolMutex.RLock()
	defer np.poolMutex.RUnlock()
	return getNodeListFromMap(np.pool)
}

func getNodeListFromMap(nodeMap map[string]*node) []*node {
	length := len(nodeMap)
	nodes := make([]*node, length, length)
	i := 0
	for _, n := range nodeMap {
		nodes[i] = n
		i++
	}
	return nodes
}

func (np *nodePool) removeNodeFromPool(id string) {
	np.poolMutex.Lock()
	if n, ok := np.pool[id]; ok {
		n.setState(nodeStateLeft)
		if n.conn != nil {
			n.close()
		}
		delete(np.pool, id)
	}
	np.poolMutex.Unlock()
}

func (np *nodePool) addOrGetNodeFromPool(id string, nodeMeta NodeMeta, logger *govec.GoLog) *node {
	np.poolMutex.Lock()
	n, ok := np.pool[id]
	if !ok {
		n = &node{
			id:      id,
			addr:    nodeMeta.Addr,
			outChan: make(chan Message, chanBufferSize),
			state:   nodeStateDisconnected,
			logger:  np.logger,
		}
		np.pool[id] = n
	}
	np.poolMutex.Unlock()
	return n
}

// return any new nodes to be handled and updates applied and whether any changes occurred
func (np *nodePool) applyReceivedUpdates(updates NetMeta) (nodeList []*node, delta NetMeta, changed bool) {
	nodeList = make([]*node, 0)
	np.netMetaMutex.Lock()
	delta, changed = np.netMeta.merge(updates)
	np.netMetaMutex.Unlock()
	for id, n := range delta {
		if n.Left {
			np.removeNodeFromPool(id)
		} else {
			nodeList = append(nodeList, np.addOrGetNodeFromPool(id, n, np.logger))
		}
	}
	return
}

func (np *nodePool) forceNodeQuit(n *node) {

}

//func (np *nodePool) sendMs

func (np *nodePool) getLatestNetMeta() NetMeta {
	np.netMetaMutex.RLock()
	defer np.netMetaMutex.RUnlock()
	return np.netMeta
}

func (np *nodePool) getLatestNetMetaJson() []byte {
	np.netMetaMutex.RLock()
	defer np.netMetaMutex.RUnlock()
	return np.netMeta.toJson()
}

func (np *nodePool) getLatestNetMetaJsonPrettyPrint() []byte {
	np.netMetaMutex.RLock()
	defer np.netMetaMutex.RUnlock()
	return np.netMeta.toJsonPrettyPrint()
}

func (np *nodePool) getLatestNetMetaCopy() NetMeta {
	np.netMetaMutex.RLock()
	defer np.netMetaMutex.RUnlock()
	return np.netMeta.copy()
}
