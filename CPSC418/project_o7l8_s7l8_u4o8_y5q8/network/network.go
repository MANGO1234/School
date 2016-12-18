// This acts as the network manager and manages the connection,
// broadcasting, and any passing of info
// among nodes

// Set-up:
// $ go get github.com/satori/go.uuid

// TODO: debug mode to log any errors

package network

import (
	"errors"
	"github.com/arcaneiceman/GoVector/govec"
	"net"
	"os"
	"strings"
)

type NetworkManager struct {
	id         string
	localAddr  string
	publicAddr string
	msgChan    chan Message
	nodePool   *nodePool
	session    *session
	// this is ugly and nt really good, maybe changed later once its working
	RemoteOpHandler      func([]byte)
	GetOpsReceiveVersion func() []byte
	VersionCheckHandler  func([]byte) ([]byte, bool)
	logger               *govec.GoLog
}

var (
	ErrAlreadyConnected    = errors.New("network: already connected")
	ErrAlreadyDisconnected = errors.New("network: already disconnected")
)

// NewNetworkManager initiate a new NetworkManager with listening
// address addr to handle network operations
func NewNetworkManager(localAddr, publicAddr string) (*NetworkManager, error) {
	os.Mkdir("govecLogTxt", os.ModeDir)
	logger := govec.Initialize(localAddr, "govecLogTxt/"+strings.Replace(localAddr, ":", "_", 100))
	manager := NetworkManager{
		localAddr:  localAddr,
		publicAddr: publicAddr,
		msgChan:    make(chan Message, 30),
		nodePool:   newNodePool(logger),
		logger:     logger,
	}
	err := startNewSessionOnNetworkManager(&manager)
	if err != nil {
		return nil, err
	}
	// stubs
	manager.SetRemoteOpHandler(func(msg []byte) {})
	manager.SetGetOpsReceiveVersion(func() []byte {
		return nil
	})
	manager.SetVersionCheckHandler(func(data []byte) ([]byte, bool) {
		return nil, false
	})
	return &manager, nil
}

func (nm *NetworkManager) GetCurrentId() string {
	return nm.id
}

// Notes on implementation of ConnectTo:
// To simplify the flow and design, we do not try to establish connection
// in the user command thread. Instead we put the load on the message handler
// which will in turn manage all connections

// ConnectTo registers the current node into the network of the node
// whose listening address is remoteAddr
func (nm *NetworkManager) ConnectTo(remoteAddr string) error {
	// TODO: maybe make the errors more friendly as it is user facing
	if nm.session == nil {
		err := startNewSessionOnNetworkManager(nm)
		if err != nil {
			return err
		}
	}
	conn, err := net.Dial("tcp", remoteAddr)
	if err != nil {
		return err
	}
	defer conn.Close()
	n := newConnWrapper(conn)
	n.logger = nm.logger
	err = n.writeMessage(dialingTypeRegister, "ConnectTo dialingTypeRegister")
	if err != nil {
		return err
	}
	incoming := new(NetMeta)
	err = n.readLog(incoming, "ConnectTo incomingNetMeta")
	if err != nil {
		return err
	}

	defer func() { nm.msgChan <- newNetMetaUpdateMsg(nm.id, *incoming) }()
	latestNetMeta := nm.nodePool.getLatestNetMeta()
	err = n.writeLog(latestNetMeta, "ConnectTo latestNetMeta")
	if err != nil {
		return errors.New("Partially connected: unable to send message to " +
			"requested node, but was able to receive information.")
	}
	nm.logger.LogLocalEvent("ConnectTo function done")
	return nil
}

// Disconnect disconnects from the rest of the network voluntarily
func (nm *NetworkManager) Disconnect() error {
	if nm.session == nil {
		return ErrAlreadyDisconnected
	}
	nm.session.end()
	nm.session = nil
	nm.logger.LogLocalEvent("Disconnected======")
	return nil
}

// Completely disconnect by throwing away all the NetMeta
func (nm *NetworkManager) CompleteDisconnect() {
	nm.Disconnect()
	nm.nodePool = newNodePool(nm.logger)
	startNewSessionOnNetworkManager(nm)
}

// Reconnect rejoins the network with new UUID.
func (nm *NetworkManager) Reconnect() error {
	nm.logger.LogLocalEvent("begin reconnect========")
	if nm.session != nil {
		return ErrAlreadyConnected
	}
	return startNewSessionOnNetworkManager(nm)
}

// Broadcast msg asynchronously return whether the msg
// if session has ended it will not broadcast
func (nm *NetworkManager) Broadcast(msg Message) {
	s := nm.session // this is necessary for thread safety and to avoid nil pointer dereference
	if s == nil || s.ended() {
		return
	}
	nm.logger.LogLocalEvent("begin broadcast========")
	s.nodePool.broadcast(msg)
}

// Send msg to a node with specified id
// if session has ended it will not send
func (nm *NetworkManager) SendMessageToNodeWithId(msg Message, id string) {
	s := nm.session // this is necessary for thread safety and to avoid nil pointer dereference
	if s == nil || s.ended() {
		return
	}
	s.nodePool.sendMessageToNodeWithId(msg, id)
}

func (nm *NetworkManager) GetNetworkMetadataString() string {
	return string(nm.nodePool.getLatestNetMetaJsonPrettyPrint())
}

func (nm *NetworkManager) GetNetworkMetadata() NetMeta {
	return nm.nodePool.getLatestNetMetaCopy()
}

func (nm *NetworkManager) SetRemoteOpHandler(fn func([]byte)) {
	nm.RemoteOpHandler = fn
}

func (nm *NetworkManager) SetVersionCheckHandler(fn func([]byte) ([]byte, bool)) {
	nm.VersionCheckHandler = fn
}

func (nm *NetworkManager) SetGetOpsReceiveVersion(fn func() []byte) {
	nm.GetOpsReceiveVersion = fn
}
