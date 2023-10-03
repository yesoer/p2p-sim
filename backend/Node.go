package backend

import (
	"sync"

	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
	"golang.org/x/net/context"
)

type Node interface {
	Send(targetId int, data any) int
	Await(cnt int) int
	ConnectTo(peerId int)
	DisconnectFrom(peerId int)
	GetConnections() []*Connection
	Run(signals chan Signal, code Code)
	SetData(json interface{})
}

type Connection struct {
	Chan   chan interface{}
	Target int
}

type node struct {
	connections []*Connection
	id          int         // TODO : still don't know about this, keep ids implicit or explicit everywhere
	data        interface{} // json data to expose to lua
}

func NewNode(id int) Node {
	var connections []*Connection
	return &node{connections, id, nil}
}

func (n *node) SetData(json interface{}) {
	n.data = json
}

// function to be used from lua to send a message (data is the first parameter)
// to a specific node
// TODO : another one to send to all
// TODO : another one to provide equation, send to all that resolve it e.g. for all even id's
func (n *node) Send(targetId int, data any) int {
	for _, c := range n.connections {
		if c.Target == targetId {
			c.Chan <- data
			return 0
		}
	}
	return 0
}

// function to be used from lua to wait for n messages from all connected peers
func (n *node) Await(cnt int) int {
	var wg sync.WaitGroup
	wg.Add(cnt)
	// channel to kill those channels where we don't expect a message ?
	kill := make(chan bool, 10)

	// listen on all channels until the specified number of messages is reached
	res := []interface{}{}
	for _, c := range n.connections {
		go func(c *Connection, wg *sync.WaitGroup) {
			for {
				select {
				case msg := <-c.Chan:
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

// a node will run continuously, the current state can be changed using signals
func (n *node) Run(signals chan Signal, code Code) {
	// code exec
	go func() {
		for {
			// wait for start signal
			// TODO : outsource to utils ?
			for sig := <-signals; sig != START; sig = <-signals {
				// await START
			}

			ctx, cancel := context.WithCancel(context.Background())

			// execute code
			go func() {
				i := interp.New(interp.Options{})

				i.Use(stdlib.Symbols)

				_, err := i.Eval(string(code))
				if err != nil {
					panic(err)
				}

				v, err := i.Eval("Run")
				if err != nil {
					panic(err)
				}

				// TODO : accept empty interface as return/do we even need returns ?
				userF := v.Interface().(func(context.Context, func(targetId int, data any) int, func(int) int) string)

				_ = userF(ctx, n.Send, n.Await)
			}()

			// wait for other signals
			for sig := range signals {
				if sig == STOP {
					// kill exec of userF and return to start of loop
					cancel()
					break
				}
			}
		}
	}()
}

// make a one way connection from  n to peer, meaning peer adds n's output as
// input
func (n *node) ConnectTo(peerId int) {
	c := make(chan interface{}, 10)
	newConnection := Connection{c, peerId}
	n.connections = append(n.connections, &newConnection)
}

func (n *node) DisconnectFrom(peerId int) {
	for connI, conn := range n.connections {
		if conn.Target == peerId {
			n.connections = append(n.connections[:connI], n.connections[connI+1:]...)
			return
		}
	}
}

func (n *node) GetConnections() []*Connection {
	return n.connections
}
