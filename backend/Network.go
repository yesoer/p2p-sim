package backend

import (
	"log"
)

type Signal int

type Code string

const (
	START Signal = 1 // start code execution from the beginning
	STOP  Signal = 2 // stop code execution
)

type Network interface {
	Init(cnt int)
	Emit(s Signal)
	GetConnections() [][]*Connection
	GetNodeCnt() int
	ConnectNodes(fromId, toId int)
	DisconnectNodes(fromId, toId int)
	SetCode(code Code)
	GetCode() Code
	SetData(json interface{}, toId int)
}

func (n *network) GetCode() Code {
	return *n.Code
}

type network struct {
	Nodes   []Node
	Signals chan Signal
	Code    *Code
}

func NewNetwork() Network {
	var nodes []Node
	var code Code
	signals := make(chan Signal, 10)
	return &network{nodes, signals, &code}
}

func (n *network) SetData(json interface{}, toId int) {
	n.Nodes[toId].SetData(json)
}

func (n *network) SetCode(code Code) {
	*n.Code = code
}

func (n *network) ConnectNodes(fromId, toId int) {
	n.Nodes[fromId].ConnectTo(toId)
}

func (n *network) DisconnectNodes(fromId, toId int) {
	n.Nodes[fromId].DisconnectFrom(toId)
}

func (n *network) Init(cnt int) {
	for i := 0; i < cnt; i++ {
		newNode := NewNode(i)
		newNode.Run(n.Signals, n.Code)
		n.Nodes = append(n.Nodes, newNode)
	}
}

func (n *network) Emit(s Signal) {
	log.Printf("Emit signal %d to nodes", s)
	for range n.Nodes {
		n.Signals <- s
	}
}

// returns exactly one connections slice for each node
func (n *network) GetConnections() [][]*Connection {
	res := make([][]*Connection, len(n.Nodes))

	for nodei, node := range n.Nodes {
		connections := node.GetConnections()
		res[nodei] = connections
	}

	return res
}

func (n *network) GetNodeCnt() int {
	return len(n.Nodes)
}
