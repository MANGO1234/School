package network

import (
	"net"
)

func (s *session) handleNewConn(conn net.Conn) {
	n := newConnWrapper(conn)
	n.logger = s.manager.logger
	// distinguish purpose of this connection
	purpose, err := n.readMessage("handleNewConn purpose")
	n.logger.LogLocalEvent(purpose)
	if err != nil {
		conn.Close()
		return
	}
	switch purpose {
	case dialingTypeRegister:
		s.handleRegister(n)
	case dialingTypePoke:
		s.handlePoke(n)
	case dialingTypeConnect:
		s.handleConnect(n)
	default:
		// invalid purpose; ignore
	}
}

func (s *session) handleRegister(connWrapper *node) {
	defer connWrapper.close()
	latestNetMeta := s.nodePool.getLatestNetMeta()
	err := connWrapper.writeLog(latestNetMeta, "handleRegister lastestNetMeta")
	if err != nil {
		return
	}
	incoming := new(NetMeta)
	err = connWrapper.readLog(incoming, "handleRegister lastestNetMeta")
	if err != nil {
		return
	}
	//incoming, err := newNetMetaFromJson(incomingNetMeta)

	if err != nil {
		return
	}
	s.manager.msgChan <- newNetMetaUpdateMsg(s.id, *incoming)
}

func (s *session) handlePoke(connWrapper *node) {
	defer connWrapper.close()
	expectedId, err := connWrapper.readMessage("handlePoke expectedId")
	if err != nil {
		return
	}
	if s.id != expectedId {
		connWrapper.writeMessage("false", "handlePoke false")
		return
	}
	err = connWrapper.writeMessage("true", "handlePoke true")
	if err != nil {
		return
	}
	id, err := connWrapper.readMessage("handlePoke id")
	if err != nil {
		return
	}
	addr, err := connWrapper.readMessage("handlePoke addr")
	if err != nil {
		return
	}
	s.manager.msgChan <- newNetMetaUpdateMsg(s.id, newJoinNetMeta(id, addr))
	// TODO: not sure if the following is necessary when using tcp
	// but it gives more guarantees
	connWrapper.writeMessage("done", "handlePoke done") // best we can do
}

func (s *session) handleConnect(connWrapper *node) {
	expectedId, err := connWrapper.readMessage("handleConnect expectedId")
	defer func(err error, connWrapper *node) {
		if err != nil {
			connWrapper.close()
		}
	}(err, connWrapper)
	if err != nil {
		return
	}
	if s.id != expectedId {
		connWrapper.writeMessage("false", "handleConnect false")
		return
	}
	err = connWrapper.writeMessage("true", "handleConnect true")
	if err != nil {
		return
	}
	id, err := connWrapper.readMessage("handleConnect id")
	if err != nil {
		return
	}
	addr, err := connWrapper.readMessage("handleConnect addr")
	if err != nil {
		return
	}
	hasMsg, msg := s.getLatestVersionCheckMsg()
	if hasMsg {
		err = connWrapper.sendMessage(msg, "handleConnect versionCheckMsg")
	}
	if err != nil {
		return
	}
	n := s.nodePool.addOrGetNodeFromPool(id, NodeMeta{addr, false}, s.manager.logger)
	n.conn = connWrapper.conn
	n.reader = connWrapper.reader
	n.writer = connWrapper.writer
	if n.setState(nodeStateConnected) {
		go s.sendThread(getSendWrapperFromNode(n))
		go s.receiveThread(n)
	}
}
