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

type EventBus struct {
	Callbacks map[EventType][]Callback
	Mu        sync.Mutex
}

func NewEventbus() *EventBus {
	return &EventBus{make(map[EventType][]Callback), sync.Mutex{}}
}

type Callback func(e Event)

// TODO : add options to allow for only once or make that a different method ?
func (bus *EventBus) Bind(etype EventType, callback Callback) {
	bus.Mu.Lock()
	bus.Callbacks[etype] = append(bus.Callbacks[etype], callback)
	bus.Mu.Unlock()
	log.Println("Bound func to event type : ", etype)
}

func (bus *EventBus) Publish(e Event) {
	for _, cb := range bus.Callbacks[e.Type] {
		cb(e)
	}
	log.Println("Published : ", e)
}
