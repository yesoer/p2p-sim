package bus

import (
	"distributed-sys-emulator/log"
	"errors"
	"reflect"
	"sync"
)

// This eventbus is supposed to serve as a connection between the core and gui.
// This includes :
// - immediately exec the callback on bind, if an event has been published before
// - no constantly running process
// - callbacks may publish
// - implicit eventtype/eventdata matching and checking

type EventType string

// callbacks are stored as any so we can store varying signatures
// type safety is ensured through the reflect package
type callback any

type Event struct {
	Type EventType `json:"type"`
	Data any       `json:"data"`
}

type eventBusData struct {
	callbacks []callback
	recent    *Event
	cbType    reflect.Type
}

type EventBus interface {
	Bind(etype EventType, cb callback)
	AwaitBind(etype EventType, cb callback) bool
	Publish(e Event)
	AwaitPublish(e Event) bool
}

type eventBus struct {
	data map[EventType]eventBusData
	mu   sync.Mutex
}

func NewEventbus() EventBus {
	data := make(map[EventType]eventBusData)
	return &eventBus{data, sync.Mutex{}}
}

func (bus *eventBus) Bind(etype EventType, cb callback) {
	bus.bind(etype, cb, false)
}

// returns whether the bind was successfull
func (bus *eventBus) AwaitBind(etype EventType, cb callback) bool {
	return bus.bind(etype, cb, true)
}

// TODO : add options to allow for only once or make that a different method ?
// TODO : we might need to be able to unbind
func (bus *eventBus) bind(etype EventType, cb callback, await bool) bool {
	if await {
		return bus.bindLogic(etype, cb)
	} else {
		go bus.bindLogic(etype, cb)
	}
	return true
}

func (bus *eventBus) bindLogic(etype EventType, cb callback) bool {
	// TODO : use sync map ?
	bus.mu.Lock()
	defer bus.mu.Unlock()

	current, ok := bus.data[etype]
	if !ok {
		// first bind/publish to this eventtype
		callbacks := []callback{cb}
		cbType := reflect.TypeOf(cb)
		bus.data[etype] = eventBusData{callbacks, nil, cbType}
	} else {
		// a previous bind/publish has implicitly defined which data type is
		// expected by callbacks, check if this callbacks signature matches
		match := current.cbType == reflect.TypeOf(cb)
		if !match {
			err := errors.New("the provided callback does not match the expected arg type")
			log.Error(err)
			return false
		}
		newCallbacks := append(current.callbacks, cb)
		bus.data[etype] = eventBusData{newCallbacks, current.recent, current.cbType}
	}
	log.Debug("Bound func to event type : ", etype)

	if current.recent != nil {
		cbv := reflect.ValueOf(cb)
		in := []reflect.Value{}
		if current.recent.Data != nil {
			arg := reflect.ValueOf(current.recent.Data)
			in = append(in, arg)
		}
		cbv.Call(in)
	}

	return true
}

func (bus *eventBus) Publish(e Event) {
	bus.publish(e, false)
}

// returns whether the publish was successfull
func (bus *eventBus) AwaitPublish(e Event) bool {
	return bus.publish(e, true)
}

func (bus *eventBus) publish(e Event, await bool) bool {
	if await {
		return bus.publishLogic(e)
	} else {
		go bus.publishLogic(e)
	}

	return true
}

func (bus *eventBus) publishLogic(e Event) bool {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	if current, ok := bus.data[e.Type]; !ok {
		// no previous bind/publish
		cbSig := getFSignature(e.Data)
		bus.data[e.Type] = eventBusData{nil, &e, cbSig}
	} else {
		current.recent = &e
		bus.data[e.Type] = current
	}

	// execute all callbacks for this event
	log.Debug("Publish Event", e)
	if current := bus.data[e.Type]; current.callbacks != nil {
		for _, cb := range current.callbacks {
			// check if event data matches expected callback arg type
			argType := getFSignature(e.Data)
			if argType != current.cbType {
				err := errors.New("event data type does not match callback arg type")
				log.Error(err)
				return false
			}

			cbv := reflect.ValueOf(cb)
			in := []reflect.Value{}
			if e.Data != nil {
				arg := reflect.ValueOf(e.Data)
				in = append(in, arg)
			}
			cbv.Call(in)
		}
	}

	return true
}

func getFSignature(arg any) reflect.Type {
	cbArgType := reflect.TypeOf(arg)
	in := []reflect.Type{}
	if cbArgType != nil {
		in = append(in, cbArgType)
	}
	out := []reflect.Type{}
	fSig := reflect.FuncOf(in, out, false)
	return fSig
}
