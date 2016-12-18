package network

import (
	"../documentmanager"
	"../util"
	"bufio"
	"fmt"
	"net"
)

// These are methods on node which acts like a connection wrapper
// Structural wise, this is probably a misuse
// TODO fix the structure if possible

func getSendWrapperFromNode(n *node) *node {
	return &node{
		conn:    n.conn,
		reader:  n.reader,
		writer:  n.writer,
		outChan: n.outChan,
		logger:  n.logger,
	}
}

func newConnWrapper(conn net.Conn) *node {
	var n node
	n.setConn(conn)
	return &n
}

func (n *node) setConn(conn net.Conn) {
	n.conn = conn
	n.reader = &util.MessageReader{bufio.NewReader(conn)}
	n.writer = &util.MessageWriter{bufio.NewWriter(conn)}
}

func (n *node) close() {
	n.conn.Close()
}

func (n *node) writeLog(buf interface{}, msgNote string) error {
	d := fmt.Sprint(buf)
	msgByte := n.logger.PrepareSend("send byte "+msgNote+" "+d, buf)
	err := n.writer.WriteMessageSlice(msgByte)
	return err
}

func (n *node) readLog(unpack interface{}, msgNote string) error {
	msg, err := n.reader.ReadMessageSlice()
	n.logger.UnpackReceive("receive byte "+msgNote, msg, unpack)
	return err
}

/*
func (n *node) writeMessageSlice(msg []byte) error {
	return n.writer.WriteMessageSlice(msg)
}
func (n *node) readMessageSlice() ([]byte, error) {
	return n.reader.ReadMessageSlice()
}
*/

func (n *node) writeMessage(msg string, msgNote string) error {
	msgByte := n.logger.PrepareSend("send string "+msgNote+" "+msg, msg)
	err := n.writer.WriteMessageSlice(msgByte)
	return err
}

func (n *node) readMessage(msgNote string) (string, error) {
	msg, err := n.reader.ReadMessageSlice()
	var unpack string
	n.logger.UnpackReceive("receive string "+msgNote, msg, &unpack)
	return unpack, err
}

func parseMessageHelper(msg Message) string {
	msgPrint := msg.Type + fmt.Sprintln(msg.Visited)
	switch msg.Type {
	case MSG_TYPE_NET_META_UPDATE:
		nm, _ := newNetMetaFromJson(msg.Msg)
		msgPrint = msgPrint + "Content: " + fmt.Sprint(nm)
		break
	case MSG_TYPE_VERSION_CHECK:
		vcm, _ := newVersionCheckMsgContentFromJson(msg.Msg)
		msgPrint = msgPrint + "Content: " + fmt.Sprint(vcm)
		break
	default: //remote op
		msgPrint = msgPrint + "Content: " + fmt.Sprint(documentmanager.RemoteOperationsFromSlice(msg.Msg))
	}
	return msgPrint
}

func (n *node) sendMessage(msg Message, msgNote string) error {
	msgPrint := parseMessageHelper(msg)
	msgByte := n.logger.PrepareSend("send byte "+msgNote+" "+msgPrint, msg)
	err := n.writer.WriteMessageSlice(msgByte)
	return err
}

func (n *node) receiveMessage(msgNote string) (Message, error, bool) {
	msgJson, err := n.reader.ReadMessageSlice()
	if err != nil {
		return Message{}, err, false
	}
	msg := new(Message)
	n.logger.UnpackReceive("receive byte "+msgNote, msgJson, &msg)
	return *msg, err, true
}
