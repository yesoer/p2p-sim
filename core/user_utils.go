package core

import (
	"github.com/yesoer/p2p-sim/bus"
	"github.com/yesoer/p2p-sim/log"

	"golang.org/x/net/context"
)

/*
* USER CODE UTILS
* The following are functions which should be exposed to the user code e.g.
* for communication between the nodes.
 */

// function to be used from user code to send a message (data is the first )
// parameter to a specific node
// TODO : feat : send to all/many
// TODO : feat : provide equation, send to all that resolve it e.g. for all even id's
func (n *node) getSender(ctx context.Context, eb bus.EventBus, debug bool) func(targetId int, data any) int {
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

			eb.AwaitEvent(ctx, bus.ContinueNodesEvt)
		}

		return reachedNodesCnt
	}
}

// Function to be used from user code to wait for n messages from all connected
// peers. It extends receivAll with debug possibilities.
// TODO : unsure about the context here, if it makes sense/would ever be used,
//
//	shouldn't it also be handed down to receiveAll ?
func (n *node) getAwaiter(ctx context.Context, eb bus.EventBus, debug bool) func(cnt int) any {
	return func(cnt int) any {
		if debug {
			awaitStart := bus.Event{Type: bus.AwaitStartEvt, Data: bus.NodeId(n.id)}
			eb.Publish(awaitStart)
		}

		log.Debug("Await ", cnt, " from ", len(n.ins), " connections")
		res := n.receiveAll(cnt)

		if debug {
			awaitEnd := bus.Event{Type: bus.AwaitEndEvt, Data: res}
			eb.Publish(awaitEnd)

			eb.AwaitEvent(ctx, bus.ContinueNodesEvt)
		}

		return res
	}
}

// Function to wait for cnt many messages from all input channels of this node.
func (n *node) receiveAll(cnt int) []bus.SendTask {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	resChan := make(chan bus.SendTask, cnt)
	defer close(resChan)
	for _, c := range n.ins {
		go n.receive(ctx, c, resChan)
	}

	// accumulate results
	res := make([]bus.SendTask, 0, cnt)
	for i := 0; i < cnt; i++ {
		response := <-resChan
		res = append(res, response)
	}

	return res
}

// Wait for messages from the given connection and report any results back to
// the caller via the res channel. Since this is running indefinitely, the caller
// may kill it through the provided context
func (n *node) receive(ctx context.Context, c connection, res chan bus.SendTask) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-c.ch:
			transmittedData := bus.SendTask{From: c.peer, To: n.id, Data: msg}
			select {
			case res <- transmittedData:
			default: /* channel possibly has been closed */
			}
		}
	}
}
