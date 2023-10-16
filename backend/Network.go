package backend

import (
	"distributed-sys-emulator/bus"
	"log"
)

type Signal int

type Code string

const (
	START Signal = 1
	STOP  Signal = 2
)

type Network interface {
	Init(cnt int, eb *bus.EventBus)
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

func NewNetwork(eb *bus.EventBus) Network {
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

func (n *network) Init(cnt int, eb *bus.EventBus) {
	for i := 0; i < cnt; i++ {
		newNode := NewNode(i)

		newNode.Run(n.Signals, n.Code)
		n.Nodes = append(n.Nodes, newNode)

		// bind node handlers to the various relevant events
		eb.Bind(bus.StartEvt, func(e bus.Event) { n.Emit(START) })

		eb.Bind(bus.StopEvt, func(e bus.Event) { n.Emit(STOP) })

		eb.Bind(bus.ConnectNodesEvt, func(e bus.Event) {
			connData := e.Data.(bus.CheckboxPos)
			n.ConnectNodes(connData.Ccol, connData.Crow)

			// publish event back to gui
			connections := n.GetConnections()
			newEvent := bus.Event{Type: bus.ConnectionChangeEvt, Data: connections}
			eb.Publish(newEvent)
		})

		eb.Bind(bus.DisconnectNodesEvt, func(e bus.Event) {
			connData := e.Data.(bus.CheckboxPos)
			n.DisconnectNodes(connData.Ccol, connData.Crow)

			// publish event back to gui
			connections := n.GetConnections()
			newEvent := bus.Event{Type: bus.ConnectionChangeEvt, Data: connections}
			eb.Publish(newEvent)
		})

		eb.Bind(bus.CodeChangedEvt, func(e bus.Event) {
			code := e.Data.(Code)
			n.SetCode(code)
		})

		eb.Bind(bus.NodeDataChangeEvt, func(e bus.Event) {
			newData := e.Data.(bus.NodeDataChangeData)
			n.SetData(newData.Data, newData.TargetId)
		})
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
