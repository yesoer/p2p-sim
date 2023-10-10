package backend

import (
	"log"
)

type Signal int

type Code string

const (
	START Signal = 1
	STOP  Signal = 2
)

type Network interface {
	Init(cnt int)
	Emit(s Signal)
	GetConnections() [][]*Connection
	GetNodeCnt() int
	ConnectNodes(fromId, toId int)
	DisconnectNodes(fromId, toId int)
	SetCode(code Code)
	SetData(json interface{}, toId int)
	GetData(toId int) interface{}
}

type network struct {
	Nodes   []Node
	Signals chan Signal
	Code    chan Code
}

func NewNetwork() Network {
	var nodes []Node
	signals := make(chan Signal, 10)
	code := make(chan Code, 10)
	return &network{nodes, signals, code}
}

func (n *network) SetData(json interface{}, toId int) {
	n.Nodes[toId].SetData(json)
}

func (n *network) GetData(toId int) interface{} {
	return n.Nodes[toId].GetData()
}

func (n *network) SetCode(code Code) {
	for range n.Nodes {
		n.Code <- code
	}
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
