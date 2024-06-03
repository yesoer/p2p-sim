package core

import (
	"bytes"
	"github.com/yesoer/p2p-sim/bus"
	"github.com/yesoer/p2p-sim/log"

	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
	"golang.org/x/net/context"
)

type Node interface {
	AddOutputTo(peerId int, c chan any)
	DelOutputTo(peerId int)
	AddInputFrom(peerId int, c chan any)
	DelInputFrom(peerId int)
	GetOutConnections() bus.Connections
	SetData(json any)
	Run(eb bus.EventBus, signals <-chan Signal)
}

// stores a connection between this node and another peer
// whether its in- or outgoing depends on the context
type connection struct {
	peer int
	ch   chan any
}

type node struct {
	ins  []connection // stores connections FROM other nodes
	outs []connection // stores connections TO other nodes
	id   int
	data any // json data to expose to user code
}

func NewNode(id int) Node {
	var ins []connection
	var outs []connection
	return &node{ins, outs, id, nil}
}

func (n *node) AddOutputTo(peerId int, c chan any) {
	newConnection := connection{peerId, c}
	n.outs = append(n.outs, newConnection)
}

func (n *node) DelOutputTo(peerId int) {
	for connI, conn := range n.outs {
		if conn.peer == peerId {
			n.outs = append(n.outs[:connI], n.outs[connI+1:]...)
			return
		}
	}
}

func (n *node) AddInputFrom(peerId int, c chan any) {
	newConnection := connection{peerId, c}
	n.ins = append(n.ins, newConnection)
}

func (n *node) DelInputFrom(peerId int) {
	for connI, conn := range n.ins {
		if conn.peer == peerId {
			n.ins = append(n.ins[:connI], n.ins[connI+1:]...)
			return
		}
	}
}

func (n *node) GetOutConnections() bus.Connections {
	res := make(bus.Connections, len(n.outs))
	for i, c := range n.outs {
		res[i] = bus.Connection{From: n.id, To: c.peer}
	}
	return res
}

func (n *node) SetData(json any) {
	n.data = json
}

// a node will run continuously, the current state can be changed using signals
func (n *node) Run(eb bus.EventBus, signals <-chan Signal) {

	// a channel to communicate all code execution results here, to be published
	// to the event bus
	resChan := make(chan bus.NodeOutput)
	defer close(resChan)
	go func() {
		data := <-resChan
		e := bus.Event{Type: bus.NodeOutputEvt, Data: data}
		eb.Publish(e)
	}()

	// keep the project source up to date
	var project bus.Source
	updateProject := func(source bus.Source) {
		if source.Type != bus.Directory {
			return
		}
		project = source
	}
	eb.Bind(bus.OpenEvt, updateProject)

	// a channel to cancel the code execution
	var codeCancel chan any

	// wait for signals which change the state of the node
	running := false
	for sig := range signals {
		switch sig {
		case START:
			code, err := bundle(project)
			if err != nil {
				data := bus.NodeOutput{Log: err.Error(), Result: nil, NodeId: n.id}
				resChan <- data
				break
			}

			if !running {
				codeCancel = make(chan any, 1)
				go n.codeExec(eb, codeCancel, code, resChan, false)
				running = true
			}
		case DEBUG:
			code, err := bundle(project)
			if err != nil {
				data := bus.NodeOutput{Log: err.Error(), Result: nil, NodeId: n.id}
				resChan <- data
				break
			}

			if !running {
				codeCancel = make(chan any, 1)
				go n.codeExec(eb, codeCancel, code, resChan, true)
				running = true
			}
		case STOP:
			if running {
				// kill exec of userF and return to start of loop
				close(codeCancel)
				running = false
			}
		case TERM:
			if running {
				close(codeCancel)
			}
			eb.Unbind(bus.OpenEvt, updateProject)
			return
		}
	}
}

// TODO : since we included eb we might not need the other channels anymore ?
func (n *node) codeExec(eb bus.EventBus, codeCancel chan any, code Code, resChan chan bus.NodeOutput, debug bool) {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-codeCancel
		cancel()
	}()

	// TODO : stream buffer changes (detected through hashes?) to UI, and should both
	var userFOut bytes.Buffer
	i := interp.New(interp.Options{Stdout: &userFOut, Stderr: &userFOut})

	if err := i.Use(stdlib.Symbols); err != nil {
		log.Error(err)
		return
	}

	_, err := i.Eval(string(code))
	if err != nil {
		log.Error(err)
		data := bus.NodeOutput{Log: err.Error(), Result: nil, NodeId: n.id}
		resChan <- data
		return
	}

	v, err := i.Eval("Run")
	if err != nil {
		log.Error(err)
		data := bus.NodeOutput{Log: err.Error(), Result: nil, NodeId: n.id}
		resChan <- data
		return
	}

	userF := v.Interface().(func(ctx context.Context, fSend func(targetId int, data any) int, fAwait func(cnt int) any) any)

	// make node specific data accessible
	outNeighborsIds := make([]int, len(n.outs))
	for i, c := range n.outs {
		outNeighborsIds[i] = c.peer
	}
	inNeighborsIds := make([]int, len(n.ins))
	for i, c := range n.ins {
		inNeighborsIds[i] = c.peer
	}
	ctx = context.WithValue(ctx, "custom", n.data)
	ctx = context.WithValue(ctx, "out-neighbors", outNeighborsIds)
	ctx = context.WithValue(ctx, "in-neighbors", inNeighborsIds)
	ctx = context.WithValue(ctx, "id", n.id)

	// Execute the provided function
	userRes := userF(ctx, n.getSender(ctx, eb, debug), n.getAwaiter(ctx, eb, debug))
	output := userFOut.String()

	data := bus.NodeOutput{Log: output, Result: userRes, NodeId: n.id}
	resChan <- data
}
