package bus

import (
	"log"
	"sync"
)

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
	Data map[EventType]EventBusData
	Mu   sync.Mutex
}

func NewEventbus() *EventBus {
	return &EventBus{make(map[EventType]EventBusData), sync.Mutex{}}
}

type Callback func(e Event)

// TODO : add options to allow for only once or make that a different method ?
// TODO : double check that this is not called in forever loops or something like that
// Binds to an event type meaning whenever such an event rises the callback is
// executed. Be aware that if such an event has been published previous to the bind,
// the callback will be executed aswell !
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
	bus.Mu.Unlock()

	if current.Recent != nil {
		callback(*current.Recent)
	}

	log.Println("Bound func to event type : ", etype)
}

// TODO : could multiple Publishes at the dame time cause data races in variables
// used by cb ? if so maybe create a cb wrapper to secure with mutex ?
// I think that's what we experienced with the console output in gui/Console.go
func (bus *EventBus) Publish(e Event) {
	bus.Mu.Lock()
	if current, ok := bus.Data[e.Type]; !ok {
		bus.Data[e.Type] = EventBusData{nil, &e}
	} else {
		current.Recent = &e
		bus.Data[e.Type] = current
	}
	bus.Mu.Unlock()

	if current := bus.Data[e.Type]; current.Callbacks != nil {
		for _, cb := range current.Callbacks {
			cb(e)
		}
	}

	log.Println("Published : ", e)
}
