package core

import (
	"bytes"
	"distributed-sys-emulator/bus"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
	"golang.org/x/net/context"
)

type Node interface {
	ConnectTo(peerId int)
	DisconnectFrom(peerId int)
	GetConnections() bus.Connections
	SetData(json interface{})
	Run(eb bus.EventBus, signals <-chan Signal)
}

type connection struct {
	to int
	ch chan interface{}
}

type node struct {
	connections []connection
	id          int
	data        interface{} // json data to expose to user code
}

type userFunc func(context.Context, func(targetId int, data any) int, func(int) int) string

func NewNode(id int) Node {
	var connections []connection
	return &node{connections, id, nil}
}

// make a one way connection from  n to peer, meaning peer adds n's output as
// input
func (n *node) ConnectTo(peerId int) {
	c := make(chan interface{}, 10)
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

func (n *node) SetData(json interface{}) {
	n.data = json
}

// a node will run continuously, the current state can be changed using signals
func (n *node) Run(eb bus.EventBus, signals <-chan Signal) {
	code := Code("")

	// TODO : we need to undo this bind when the node stops
	eb.Bind(bus.CodeChangeEvt, func(newCode Code) {
		code = newCode
	})

	// code exec
	go func() {
		var cancel context.CancelFunc
		ctx, cancel := context.WithCancel(context.Background())
		var output string
		exec := func() {
			var stdout, stderr bytes.Buffer
			i := interp.New(interp.Options{Stdout: &stdout, Stderr: &stderr})

			if err := i.Use(stdlib.Symbols); err != nil {
				panic(err)
			}

			_, err := i.Eval(string(code))
			if err != nil {
				panic(err)
			}

			v, err := i.Eval("Run")
			if err != nil {
				fmt.Println("Error ", err)
				return
			}

			// TODO : accept any return (should be shown on the corresponding node ?
			userF := v.Interface().(func(context.Context, func(targetId int, data any) int, func(int) int) string)

			// make node specific data accessible
			ctx = context.WithValue(ctx, "custom", n.data)
			ctx = context.WithValue(ctx, "connections", n.connections)
			ctx = context.WithValue(ctx, "id", n.id)

			// Execute the provided function
			userF(ctx, n.send, n.await)

			output = stdout.String()
		}

		// wait for other signals
		running := false
		for sig := range signals {
			log.Println("Node ", n.id, " received signal ", sig)
			switch sig {
			case START:
				if !running {
					go exec()
					running = true
				}
			case STOP:
				if running {
					// kill exec of userF and return to start of loop
					cancel()
					running = false
					// TODO : this waiting for userF to cancel sucks
					time.Sleep(time.Second * 10)
					data := bus.NodeLog{Str: output, NodeId: n.id}
					e := bus.Event{Type: bus.NodeOutputLogEvt, Data: data}
					eb.Publish(e)
				}
			case TERM:
				if running {
					cancel()
				}
				return
			}
		}
	}()
}

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
func (n *node) await(cnt int) int {
	var wg sync.WaitGroup
	wg.Add(cnt)
	// channel to kill those channels where we don't expect a message ?
	kill := make(chan bool, 10)

	// listen on all channels until the specified number of messages is reached
	res := []interface{}{}
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

	// TODO : return res
	return 1
}
