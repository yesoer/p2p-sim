package core

import (
	"bytes"
	"distributed-sys-emulator/bus"
	"distributed-sys-emulator/log"
	"fmt"
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
	ch chan interface{} // TODO : I think the channels are not intact sometimes (reproduce : run and add connections after and run again ?)
}

type node struct {
	connections []connection
	id          int
	data        interface{} // json data to expose to user code
}

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
		var userRes interface{}
		exec := func() {
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
				fmt.Println("Error ", err)
				return
			}

			userF := v.Interface().(func(ctx context.Context, fSend func(targetId int, data any) int, fAwait func(cnt int) []interface{}) interface{})

			// make node specific data accessible
			ctx = context.WithValue(ctx, "custom", n.data)
			ctx = context.WithValue(ctx, "connections", n.connections)
			ctx = context.WithValue(ctx, "id", n.id)

			// Execute the provided function
			userRes = userF(ctx, n.send, n.await)

			output = userFOut.String()
		}

		// wait for other signals
		running := false
		for sig := range signals {
			log.Debug("Node ", n.id, " received signal ", sig)
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
					time.Sleep(time.Second * 5)

					data := bus.NodeOutput{Log: output, Result: userRes, NodeId: n.id}
					e := bus.Event{Type: bus.NodeOutputEvt, Data: data}
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
func (n *node) await(cnt int) []interface{} {
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

	return res
}
