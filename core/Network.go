package core

import (
	"distributed-sys-emulator/bus"
	"distributed-sys-emulator/log"
)

type Code string

type Signal int

const (
	START Signal = 1
	STOP  Signal = 2
	TERM  Signal = 3
)

const initialNodeCnt = 2

type Network interface {
	Init(eb bus.EventBus)
}

type network struct {
	nodes   []Node
	signals chan Signal
	nodeCnt int
}

func NewNetwork(eb bus.EventBus) Network {
	var nodes []Node
	signals := make(chan Signal, 10) // TODO : should the buffersize depend on nodecnt ?
	cnt := initialNodeCnt
	return network{nodes, signals, cnt}
}

func (n network) Init(eb bus.EventBus) {
	n.setAndRunNodes(eb)

	// bind node handlers to the various relevant events
	eb.Bind(bus.StartNodesEvt, func() { n.emit(START) })

	eb.Bind(bus.StopNodesEvt, func() { n.emit(STOP) })

	eb.Bind(bus.ConnectNodesEvt, func(connData bus.Connection) {
		n.connectNodes(connData.From, connData.To)

		// publish event back to gui
		connections := n.getConnections()
		newEvent := bus.Event{Type: bus.NetworkConnectionsEvt, Data: connections}
		eb.Publish(newEvent)
	})

	eb.Bind(bus.DisconnectNodesEvt, func(connData bus.Connection) {
		n.disconnectNodes(connData.From, connData.To)

		// publish event back to gui
		connections := n.getConnections()
		newEvent := bus.Event{Type: bus.NetworkConnectionsEvt, Data: connections}
		eb.Publish(newEvent)
	})

	eb.Bind(bus.NodeDataChangeEvt, func(newData bus.NodeData) {
		n.setData(newData, newData.TargetId)
	})

	eb.Bind(bus.NodeCntChangeEvt, func(newCnt int) {
		n.resize(eb, newCnt)
	})

	// publish the initial node count to the ui
	resizeData := bus.NetworkResize{Connections: nil, Cnt: n.nodeCnt}
	evt := bus.Event{Type: bus.NetworkResizeEvt, Data: resizeData}
	eb.Publish(evt)
}

func (n *network) resize(eb bus.EventBus, newCnt int) {
	// update node count
	n.setNodeCnt(newCnt)

	// store connections
	oldNetworkC := n.getConnections()

	// restart nodes
	n.emit(TERM)
	n.setAndRunNodes(eb)

	// set connections
	var newNetworkC bus.Connections
	for _, nodeC := range oldNetworkC {
		if nodeC.From < newCnt && nodeC.To < newCnt {
			newNetworkC = append(newNetworkC, nodeC)
			n.connectNodes(nodeC.From, nodeC.To)
		}
	}

	// send NetworkNodeCntChangeEvt
	resizeData := bus.NetworkResize{Connections: newNetworkC, Cnt: newCnt}
	sizeEvt := bus.Event{Type: bus.NetworkResizeEvt, Data: resizeData}
	eb.Publish(sizeEvt)
}

func (n *network) setNodeCnt(cnt int) {
	n.nodeCnt = cnt
}

func (n *network) setData(json any, toId int) {
	n.nodes[toId].SetData(json)
}

func (n network) connectNodes(fromId, toId int) {
	c := make(chan any, 10)
	n.nodes[fromId].AddOutputTo(toId, c)
	n.nodes[toId].AddInputFrom(fromId, c)
}

func (n network) disconnectNodes(fromId, toId int) {
	n.nodes[fromId].DelOutputTo(toId)
	n.nodes[toId].DelInputFrom(fromId)
}

func (n *network) setAndRunNodes(eb bus.EventBus) {
	cnt := n.nodeCnt
	n.nodes = make([]Node, cnt)
	for i := 0; i < cnt; i++ {
		newNode := NewNode(i)
		go newNode.Run(eb, n.signals)
		n.nodes[i] = newNode
	}
}

func (n *network) emit(s Signal) {
	log.Debug("Emit signal %d to nodes", s)
	for range n.nodes {
		n.signals <- s
	}
}

// returns exactly one connections slice for each node
func (n *network) getConnections() bus.Connections {
	var res bus.Connections
	for _, node := range n.nodes {
		connections := node.GetOutConnections()
		res = append(res, connections...)
	}

	return res
}
