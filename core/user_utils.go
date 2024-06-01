package core

import (
	"distributed-sys-emulator/bus"
	"distributed-sys-emulator/log"

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

// function to be used from user code to wait for n messages from all connected
// peers
func (n *node) getAwaiter(ctx context.Context, eb bus.EventBus, debug bool) func(cnt int) []any {
	return func(cnt int) []any {
		if debug {
			awaitStart := bus.Event{Type: bus.AwaitStartEvt, Data: bus.NodeId(n.id)}
			eb.Publish(awaitStart)
		}

		log.Debug("Await ", cnt, " from ", len(n.ins), " connections")
		res, userRes := n.receiveAll(cnt)

		if debug {
			awaitEnd := bus.Event{Type: bus.AwaitEndEvt, Data: res}
			eb.Publish(awaitEnd)

			eb.AwaitEvent(ctx, bus.ContinueNodesEvt)
		}

		return userRes
	}
}

func (n *node) receiveAll(cnt int) ([]bus.SendTask, []any) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	resChan := make(chan bus.SendTask, cnt)
	defer close(resChan)

	for _, c := range n.ins {
		go n.receive(ctx, c, resChan)
	}

	// accumulate results
	// TODO : I feel like two slices shouldn't be necessary
	res := make([]bus.SendTask, 0, cnt)
	userRes := make([]any, 0, cnt)
	for i := 0; i < cnt; i++ {
		response := <-resChan
		res = append(res, response)
		userRes = append(userRes, response)
	}

	return res, userRes
}

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
