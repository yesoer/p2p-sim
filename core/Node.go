package core

import (
	"bytes"
	"distributed-sys-emulator/bus"
	"distributed-sys-emulator/log"
	"fmt"
	"sync"

	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
	"golang.org/x/net/context"
)

type Node interface {
	ConnectTo(peerId int)
	DisconnectFrom(peerId int)
	GetConnections() bus.Connections
	SetData(json any)
	Run(eb bus.EventBus, signals <-chan Signal)
}

type connection struct {
	to int
	ch chan any
}

type node struct {
	connections []connection
	id          int
	data        any // json data to expose to user code
}

func NewNode(id int) Node {
	var connections []connection
	return &node{connections, id, nil}
}

// make a one way connection from  n to peer, meaning peer adds n's output as
// input
func (n *node) ConnectTo(peerId int) {
	c := make(chan any, 10)
	newConnection := connection{peerId, c}
	n.connections = append(n.connections, newConnection)
}

func (n *node) DisconnectFrom(peerId int) {
	for connI, conn := range n.connections {
		if conn.to == peerId {
			n.connections = append(n.connections[:connI], n.connections[connI+1:]...)
			return
		}
	}
}

func (n *node) GetConnections() bus.Connections {
	res := make(bus.Connections, len(n.connections))
	for i, c := range n.connections {
		res[i] = bus.Connection{From: n.id, To: c.to}
	}
	return res
}

func (n *node) SetData(json any) {
	n.data = json
}

// a node will run continuously, the current state can be changed using signals
func (n *node) Run(eb bus.EventBus, signals <-chan Signal) {
	// TODO : we need to undo this bind when the node stops
	code := Code("")
	eb.Bind(bus.CodeChangeEvt, func(newCode Code) {
		code = newCode
	})

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
				go n.codeExec(codeCancel, code, resChan)
				running = true
			}
		case STOP:
			if running {
				// kill exec of userF and return to start of loop
				close(codeCancel)
				fmt.Print("waiting ")
				data := <-resChan
				fmt.Print("publish ", data)
				e := bus.Event{Type: bus.NodeOutputEvt, Data: data}
				eb.Publish(e)
				running = false
			}
		case TERM:
			if running {
				close(codeCancel)
				close(resChan)
			}
			return
		}
	}
}

func (n *node) codeExec(codeCancel chan any, code Code, resChan chan bus.NodeOutput) {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-codeCancel
		cancel()
	}()

	// TODO : stream buffer changes (detected through hashes?) to UI, and should both
	var userFOut bytes.Buffer
	i := interp.New(interp.Options{Stdout: &userFOut, Stderr: &userFOut})

	if err := i.Use(stdlib.Symbols); err != nil {
		panic(err)
	}

	_, err := i.Eval(string(code))
	if err != nil {
		panic(err)
	}

	v, err := i.Eval("Run")
	if err != nil {
		log.Error(err)
		data := bus.NodeOutput{Log: err.Error(), Result: nil, NodeId: n.id}
		resChan <- data
	}

	userF := v.Interface().(func(ctx context.Context, fSend func(targetId int, data any) int, fAwait func(cnt int) []any) any)

	// make node specific data accessible
	neighborsIds := make([]int, len(n.connections))
	for i, c := range n.connections {
		neighborsIds[i] = c.to
	}
	ctx = context.WithValue(ctx, "custom", n.data)
	ctx = context.WithValue(ctx, "neighbors", neighborsIds)
	ctx = context.WithValue(ctx, "id", n.id)

	// Execute the provided function
	userRes := userF(ctx, n.send, n.await)
	output := userFOut.String()

	data := bus.NodeOutput{Log: output, Result: userRes, NodeId: n.id}
	fmt.Print("going to send ", resChan)
	resChan <- data
	fmt.Print("res sent ", resChan)
}

/*
* USER CODE UTILS
* The following are functions which should be exposed to the user code e.g.
* for communication between the nodes.
 */

// function to be used from user code to send a message (data is the first )
// parameter to a specific node
// TODO : another one to send to all
// TODO : another one to provide equation, send to all that resolve it e.g. for all even id's
func (n *node) send(targetId int, data any) int {
	for _, c := range n.connections {
		if c.to == targetId {
			c.ch <- data
			return 0
		}
	}
	return 0
}

// function to be used from user code to wait for n messages from all connected
// peers
func (n *node) await(cnt int) []any {
	var wg sync.WaitGroup
	wg.Add(cnt)
	// channel to kill those channels where we don't expect a message ?
	kill := make(chan bool, 10)

	// listen on all channels until the specified number of messages is reached
	res := []any{}
	for _, c := range n.connections {
		go func(c connection, wg *sync.WaitGroup) {
			for {
				select {
				case msg := <-c.ch:
					res = append(res, msg)
					wg.Done()
				case <-kill:
					return
				}
			}
		}(c, &wg)
	}

	wg.Wait()
	for i := 0; i <= len(n.connections)-cnt; i++ {
		kill <- true
	}

	return res
}
