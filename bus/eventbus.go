package bus

import (
	"log"
	"sync"
)

// This eventbus is supposed to serve as a connection between backend and gui.
// This includes :
// - immediately exec the callback on bind, if an event has been published before
// - no constantly running process
// - callbacks may publish
// - preserve publish order

type EventType string

type Event struct {
	Type EventType   `json:"type"`
	Data interface{} `json:"data"`
}

type EventBusData struct {
	Callbacks []Callback
	Recent    *Event
}

type EventBus struct {
	Data       map[EventType]EventBusData
	WaitList   map[int]chan bool
	WaitListMu sync.Mutex
	NextId     int
	Mu         sync.Mutex
}

func NewEventbus() *EventBus {
	data := make(map[EventType]EventBusData)
	waitlist := make(map[int]chan bool)
	return &EventBus{data, waitlist, sync.Mutex{}, 0, sync.Mutex{}}
}

type Callback func(e Event)

// TODO : add options to allow for only once or make that a different method ?
// TODO : double check that this is not called in forever loops or something like that
func (bus *EventBus) Bind(etype EventType, callback Callback) {
	bus.Mu.Lock()
	current, ok := bus.Data[etype]
	if !ok {
		// first bind init EventBusData
		callbacks := []Callback{callback}
		bus.Data[etype] = EventBusData{callbacks, nil}
	} else {
		newCallbacks := append(current.Callbacks, callback)
		bus.Data[etype] = EventBusData{newCallbacks, current.Recent}
	}

	if current.Recent != nil {
		callback(*current.Recent)
	}
	log.Println("Bound func to event type : ", etype)
	bus.Mu.Unlock()
}

func (bus *EventBus) Publish(e Event) {
	// wait for previous publish to finish
	bus.WaitListMu.Lock()
	id := bus.NextId
	ch, ok := bus.WaitList[id]
	if !ok {
		ch = make(chan bool, 10)
		bus.WaitList[id] = ch
	}
	bus.NextId++
	bus.WaitListMu.Unlock()

	go func() {
		if id > 0 {
			<-ch
		}

		// execute callbacks
		bus.Mu.Lock()
		if current, ok := bus.Data[e.Type]; !ok {
			bus.Data[e.Type] = EventBusData{nil, &e}
		} else {
			current.Recent = &e
			bus.Data[e.Type] = current
		}

		if current := bus.Data[e.Type]; current.Callbacks != nil {
			for _, cb := range current.Callbacks {
				cb(e)
			}
		}
		log.Println("Published event : ", e)
		bus.Mu.Unlock()

		// continue with the next publish
		bus.WaitListMu.Lock()
		next := id + 1
		nextch, ok := bus.WaitList[next]
		if !ok {
			nextch = make(chan bool, 10)
			bus.WaitList[next] = nextch
		}
		nextch <- true
		bus.WaitListMu.Unlock()
	}()
}
