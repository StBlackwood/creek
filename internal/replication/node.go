package replication

import (
	"fmt"
	"net"
)

type Node struct {
	Id       string
	Address  string
	conn     net.Conn
	IsSelf   bool
	IsLeader bool
}

func (n *Node) String() string {
	return n.Id
}

func (n *Node) writeData(msg string) error {
	if msg == "" {
		return nil
	}
	_, err := n.conn.Write([]byte(msg))
	return err
}

func (n *Node) IsConnected() bool {
	return n.conn != nil
}

func (n *Node) Close() error {
	return n.conn.Close()
}

func (n *Node) SendRepCmd(cmd *RepCmd) error {
	if !n.IsConnected() {
		return fmt.Errorf("node is not connected")
	}

	return n.writeData(cmd.String())
}
