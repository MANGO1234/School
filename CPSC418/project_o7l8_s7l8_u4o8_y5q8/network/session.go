package network

import (
	"github.com/satori/go.uuid"
	"net"
	"time"
)

// how often to check if version on two nodes match
const versionCheckInterval = 3 * time.Second

type session struct {
	id       string
	listener *net.TCPListener
	manager  *NetworkManager
	done     chan struct{}
	nodePool *nodePool
}

func startNewSessionOnNetworkManager(nm *NetworkManager) error {
	lAddr, err := net.ResolveTCPAddr("tcp", nm.localAddr)
	if err != nil {
		return err
	}
	listener, err := net.ListenTCP("tcp", lAddr)
	if err != nil {
		return err
	}
	newSession := session{
		id:       uuid.NewV1().String(),
		listener: listener,
		manager:  nm,
		done:     make(chan struct{}),
		nodePool: nm.nodePool,
	}
	newSession.nodePool.handleNewSession(&newSession)
	go newSession.listenForNewConn()
	go newSession.periodicallyCheckVersion()
	go newSession.serveIncomingMessages()
	nm.id = newSession.id
	nm.session = &newSession
	return nil
}

func (s *session) end() {
	close(s.done)
	s.listener.Close()
	s.nodePool.handleEndSession(s)
}

func (s *session) ended() bool {
	select {
	case <-s.done:
		return true
	default:
		return false
	}
}

func (s *session) serveIncomingMessages() {
	for done := false; !done || len(s.manager.msgChan) > 0; {
		select {
		case msg := <-s.manager.msgChan:
			switch msg.Type {
			case MSG_TYPE_NET_META_UPDATE:
				s.handleIncomingNetMeta(msg)
			case MSG_TYPE_REMOTE_OP:
				s.handleIncomingRemoteOp(msg)
			case MSG_TYPE_VERSION_CHECK:
				s.handleIncomingVersionCheck(msg)
			default:
				// ignore and do nothing
			}
		case <-s.done:
			done = true
		}
	}
}

func (s *session) handleIncomingNetMeta(msg Message) {
	updates, err := newNetMetaFromJson(msg.Msg)
	if err != nil {
		return
	}
	newNodes, deltaNetMeta, changed := s.nodePool.applyReceivedUpdates(updates)
	if changed {
		s.handleNewNodes(newNodes)
		msg.Msg = deltaNetMeta.toJson()
		s.nodePool.broadcast(msg)
	}
}

func (s *session) handleIncomingRemoteOp(msg Message) {
	if s.manager.RemoteOpHandler != nil {
		go s.manager.RemoteOpHandler(msg.Msg)
	}
	if msg.Visited != nil { // we should only recursively broadcast in this case
		s.nodePool.broadcast(msg)
	}
}

func (s *session) handleIncomingVersionCheck(msg Message) {
	content, err := newVersionCheckMsgContentFromJson(msg.Msg)
	if err != nil {
		return
	}
	s.handleIncomingNetMeta(newNetMetaUpdateMsg(s.id, content.NetworkMeta))
	syncInfo, shouldReply := s.manager.VersionCheckHandler(content.VersionVector)
	if shouldReply {
		toSend := NewReplyMessage(MSG_TYPE_REMOTE_OP, syncInfo)
		go func() {
			s.nodePool.sendMessageToNodeWithId(toSend, content.Source)
		}()
	}
}

// These functions launches major network threads
func (s *session) listenForNewConn() {
	for {
		conn, err := s.listener.Accept()
		if s.ended() {
			return
		}
		if err != nil {
			// TODO: What else can we do in this case
			continue
		}
		go s.handleNewConn(conn)
	}
}

func (s *session) handleNewNodes(nodes []*node) {
	for _, n := range nodes {
		s.initiateNewNode(n)
	}
}

func (s *session) periodicallyCheckVersion() {
	for {
		time.Sleep(versionCheckInterval)
		hasMsg, msg := s.getLatestVersionCheckMsg()
		if s.ended() {
			return
		}
		if hasMsg {
			s.nodePool.broadcast(msg)
		}
	}
}

func (s *session) getLatestVersionCheckMsg() (bool, Message) {
	slice := s.manager.GetOpsReceiveVersion()
	if slice != nil {
		latestMeta := s.nodePool.getLatestNetMetaCopy()
		versionCheckMsgContent := VersionCheckMsgContent{
			s.id,
			latestMeta,
			slice,
		}
		content := versionCheckMsgContent.toJson()
		return true, newSyncOrCheckMessage(MSG_TYPE_VERSION_CHECK, content)
	}
	return false, Message{}
}
