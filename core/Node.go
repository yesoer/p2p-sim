package core

import (
	"bytes"
	"distributed-sys-emulator/bus"
	"distributed-sys-emulator/log"
	"sync"

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
	ins  []connection // stores connections TO other nodes
	outs []connection // stores connections FROM other nodes
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
	code := Code("")
	updateCode := func(newCode Code) {
		code = newCode
		log.Debug("node ", n.id, " received code")
	}
	eb.Bind(bus.CodeChangeEvt, updateCode)

	var codeCancel chan any
	resChan := make(chan bus.NodeOutput)

	// wait for other signals
	running := false
	for sig := range signals {
		log.Debug("Node ", n.id, " received signal ", sig)
		switch sig {
		case START:
			if !running {
				codeCancel = make(chan any, 1)
				go n.codeExec(eb, codeCancel, code, resChan, false)
				running = true
			}
		case DEBUG:
			if !running {
				codeCancel = make(chan any, 1)
				go n.codeExec(eb, codeCancel, code, resChan, true)
				running = true
			}
		case STOP:
			if running {
				// kill exec of userF and return to start of loop
				close(codeCancel)
				data := <-resChan
				e := bus.Event{Type: bus.NodeOutputEvt, Data: data}
				eb.Publish(e)
				running = false
			}
		case TERM:
			if running {
				close(codeCancel)
			}
			close(resChan)
			eb.Unbind(bus.CodeChangeEvt, updateCode)
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

	userF := v.Interface().(func(ctx context.Context, fSend func(targetId int, data any) int, fAwait func(cnt int) []any) any)

	// make node specific data accessible
	neighborsIds := make([]int, len(n.outs))
	for i, c := range n.outs {
		neighborsIds[i] = c.peer
	}
	ctx = context.WithValue(ctx, "custom", n.data)
	ctx = context.WithValue(ctx, "outgoing-connections", neighborsIds)
	ctx = context.WithValue(ctx, "id", n.id)

	// Execute the provided function
	userRes := userF(ctx, n.getSender(eb, debug), n.getAwaiter(eb, debug)) // TODO : pass ctx to getAwaiter and getSender aswell for canceling (e.g. while blocking)
	output := userFOut.String()

	data := bus.NodeOutput{Log: output, Result: userRes, NodeId: n.id}
	resChan <- data
}

/*
* USER CODE UTILS
* The following are functions which should be exposed to the user code e.g.
* for communication between the nodes.
 */

// function to be used from user code to send a message (data is the first )
// parameter to a specific node
// TODO : feat : send to all/many
// TODO : feat : provide equation, send to all that resolve it e.g. for all even id's
func (n *node) getSender(eb bus.EventBus, debug bool) func(targetId int, data any) int {
	return func(targetId int, data any) int {
		reachedNodesCnt := 0
		for _, c := range n.outs {
			if c.peer == targetId {
				c.ch <- data
				reachedNodesCnt++
				break
			}
		}

		if debug {
			sendEvtData := bus.SendTask{From: n.id, To: targetId, Data: data}
			sendEvt := bus.Event{Type: bus.SentToEvt, Data: sendEvtData}
			eb.Publish(sendEvt)

			// TODO : this solution is ugly, need a builtin for this use case
			// block until continue nodes is published
			var wg sync.WaitGroup
			wg.Add(2)
			eb.Bind(bus.ContinueNodesEvt, func() {
				wg.Done()
			})
			wg.Wait()
		}

		return reachedNodesCnt
	}
}

// function to be used from user code to wait for n messages from all connected
// peers
func (n *node) getAwaiter(eb bus.EventBus, debug bool) func(cnt int) []any {
	return func(cnt int) []any {
		if debug {
			awaitStart := bus.Event{Type: bus.AwaitStartEvt, Data: bus.NodeId(n.id)}
			eb.Publish(awaitStart)
		}

		var wg sync.WaitGroup
		wg.Add(cnt)

		kill := make(chan bool, 10) // channel to send kill signals

		// listen on all channels until the specified number of messages is reached
		var res []bus.SendTask
		var userRes []any
		log.Debug("Await ", cnt, " from ", len(n.ins), " connections")
		for _, c := range n.ins {
			go func(c connection, wg *sync.WaitGroup) {
				for {
					select {
					case msg := <-c.ch:
						transmittedData := bus.SendTask{From: n.id, To: c.peer, Data: msg}
						res = append(res, transmittedData)
						userRes = append(userRes, transmittedData)
						wg.Done()
					case <-kill:
						return
					}
				}
			}(c, &wg)
		}

		wg.Wait()

		for i := 0; i <= len(n.ins)-cnt; i++ {
			kill <- true
		}

		if debug {
			awaitEnd := bus.Event{Type: bus.AwaitEndEvt, Data: res}
			eb.Publish(awaitEnd)

			// TODO : this solution is ugly, need a builtin for this use case
			// block until continue nodes is published
			var wg sync.WaitGroup
			wg.Add(2)
			eb.Bind(bus.ContinueNodesEvt, func() {
				wg.Done()
			})
			wg.Wait()
		}

		return userRes
	}
}
