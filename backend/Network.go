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
	TERM  Signal = 3
)

const InitialNodeCnt = 2

type Network interface {
	Init(eb *bus.EventBus)
	Emit(s Signal)
	GetConnections() [][]*Connection
	GetNodeCnt() int
	ConnectNodes(fromId, toId int)
	DisconnectNodes(fromId, toId int)
	SetData(json interface{}, toId int)
	GetData(toId int) interface{}
}

type network struct {
	Nodes   []Node
	Signals chan Signal
	NodeCnt int
}

func NewNetwork(eb *bus.EventBus) Network {
	var nodes []Node
	signals := make(chan Signal, 10)
	cnt := InitialNodeCnt
	return &network{nodes, signals, cnt}
}

func (n *network) setNodeCnt(cnt int) {
	n.NodeCnt = cnt
}

func (n *network) SetData(json interface{}, toId int) {
	n.Nodes[toId].SetData(json)
}

func (n *network) GetData(toId int) interface{} {
	return n.Nodes[toId].GetData()
}

func (n *network) ConnectNodes(fromId, toId int) {
	n.Nodes[fromId].ConnectTo(toId)
}

func (n *network) DisconnectNodes(fromId, toId int) {
	n.Nodes[fromId].DisconnectFrom(toId)
}

func (n *network) setAndRunNodes(eb *bus.EventBus) {
	cnt := n.NodeCnt
	n.Nodes = make([]Node, cnt)
	for i := 0; i < cnt; i++ {
		newNode := NewNode(i)
		newNode.Run(eb, n.Signals)
		n.Nodes[i] = newNode
	}
}

func (n *network) Init(eb *bus.EventBus) {
	n.setAndRunNodes(eb)

	// bind node handlers to the various relevant events
	// TODO : I think some, if not all of these should be outside the loop
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

	eb.Bind(bus.NodeDataChangeEvt, func(e bus.Event) {
		newData := e.Data.(bus.NodeDataChangeData)
		n.SetData(newData.Data, newData.TargetId)
	})

	eb.Bind(bus.GUINodeCntChangeEvt, func(e bus.Event) {
		// update node count
		newCnt := e.Data.(int)
		n.setNodeCnt(newCnt)

		// store connections
		// TODO : as mentioned on the nodes struct, connections are poorly designed,
		// when reworking try to create a design that avoids this reset :
		networkC := n.GetConnections()
		if newCnt < len(networkC) {
			networkC = networkC[:newCnt]
			for src, nodeC := range networkC {
				var newNodeC []*Connection
				for _, c := range nodeC {
					if c.Target < newCnt {
						newNodeC = append(newNodeC, c)
					}
				}
				networkC[src] = newNodeC
			}
		} else {
			for i := 0; i < newCnt-len(networkC); i++ {
				newSlice := make([]*Connection, 0)
				networkC = append(networkC, newSlice)
			}
		}
		connEvt := bus.Event{Type: bus.ConnectionChangeEvt, Data: networkC}
		eb.Publish(connEvt)

		// restart nodes
		n.Emit(TERM)
		n.setAndRunNodes(eb)

		// set connections
		for src, nodeC := range networkC {
			for _, c := range nodeC {
				n.ConnectNodes(c.Target, src)
			}
		}

		// send NetworkNodeCntChangeEvt
		sizeEvt := bus.Event{Type: bus.NetworkNodeCntChangeEvt, Data: newCnt}
		eb.Publish(sizeEvt)
	})

	// publish the initial node count to the ui
	evt := bus.Event{Type: bus.NetworkNodeCntChangeEvt, Data: n.NodeCnt}
	eb.Publish(evt)
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
