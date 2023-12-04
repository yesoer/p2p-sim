package bus

import (
	"context"
	"distributed-sys-emulator/log"
	"distributed-sys-emulator/smap"
	"errors"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
)

/* This eventbus is supposed to serve as a connection between the core and gui.
* This includes :
* - immediately exec the callback on bind, if an event has been published before
* - no constantly running process
* - callbacks may publish
* - implicit eventtype/eventdata matching and checking at runtime
 */

type EventType string

type Event struct {
	Type EventType `json:"type"`
	Data any       `json:"data"`
}

type eventBusData struct {
	callbacks []reflect.Value // called on event occurence and possibly once when added with the most recent event
	waitlist  []chan bool     // can be used to await the next occurence of an event
	recent    *Event
	cbType    reflect.Type // defines the expected callback signature and publish arg type
}

type EventBus interface {
	Bind(etype EventType, cb any)
	AwaitBind(etype EventType, cb any) bool
	Publish(e Event)
	AwaitPublish(e Event) bool
	Unbind(etype EventType, cb any)
	AwaitUnbind(etype EventType, cb any) bool
	AwaitEvent(ctx context.Context, etype EventType)
}

type eventBus struct {
	data smap.SMap[EventType, eventBusData]
}

func NewEventbus() EventBus {
	data := smap.NewSMap[EventType, eventBusData]()
	return &eventBus{data}
}

func (bus *eventBus) AwaitEvent(ctx context.Context, etype EventType) {
	_, ok := bus.data.Load(etype)
	if !ok {
		callbacks := []reflect.Value{}
		waitlist := []chan bool{}
		var cbType reflect.Type
		bus.data.Store(etype, eventBusData{callbacks, waitlist, nil, cbType})
	}

	wait := make(chan bool)
	modifier := func(value eventBusData) eventBusData {
		value.waitlist = append(value.waitlist, wait)
		return value
	}
	bus.data.Update(etype, modifier)

	trace := trace()
	log.Debug("Await Event ", etype, trace)
	select {
	case <-ctx.Done():
		log.Debug("Cancel awaiting Event ", etype, trace)
	case <-wait:
		log.Debug("Continue after Event ", etype, trace)
	}
	return
}

func (bus *eventBus) Unbind(etype EventType, cb any) {
	bus.unbind(etype, cb, false)
}

// returns whether the bind was successfull
func (bus *eventBus) AwaitUnbind(etype EventType, cb any) bool {
	return bus.unbind(etype, cb, true)
}

func (bus *eventBus) unbind(etype EventType, cb any, await bool) bool {
	trace := trace()
	if await {
		return bus.unbindLogic(etype, cb, trace)
	} else {
		go bus.unbindLogic(etype, cb, trace)
	}
	return true
}

func (bus *eventBus) unbindLogic(etype EventType, cb any, trace string) bool {
	current, ok := bus.data.Load(etype)
	if !ok {
		return true
	}

	cbv := reflect.ValueOf(cb)
	for i, registeredCB := range current.callbacks {
		if registeredCB.Pointer() == cbv.Pointer() {
			modifier := func(value eventBusData) eventBusData {
				value.callbacks = append(value.callbacks[:i], value.callbacks[i+1:]...)
				return value
			}
			bus.data.Update(etype, modifier)
			log.Debug("Unbound callback from event: ", etype, trace)
			return true
		}
	}
	log.Debug("Failed unbind callback from event: ", etype, trace)
	return false
}

func (bus *eventBus) Bind(etype EventType, cb any) {
	bus.bind(etype, cb, false)
}

// returns whether the bind was successfull
func (bus *eventBus) AwaitBind(etype EventType, cb any) bool {
	return bus.bind(etype, cb, true)
}

// TODO : add options to allow for only once or make that a different method ?
func (bus *eventBus) bind(etype EventType, cb any, await bool) bool {
	trace := trace()
	if await {
		return bus.bindLogic(etype, cb, trace)
	} else {
		go bus.bindLogic(etype, cb, trace)
	}
	return true
}

func (bus *eventBus) bindLogic(etype EventType, cb any, trace string) bool {

	current, ok := bus.data.Load(etype)
	cbv := reflect.ValueOf(cb)
	if !ok {
		// first bind/publish to this eventtype
		callbacks := []reflect.Value{cbv}
		waitlist := []chan bool{}
		cbType := reflect.TypeOf(cb)
		bus.data.Store(etype, eventBusData{callbacks, waitlist, nil, cbType})
	} else {
		// a previous bind/publish has implicitly defined which data type is
		// expected by callbacks, check if this callbacks signature matches
		match := current.cbType == reflect.TypeOf(cb)
		if !match {
			err := errors.New("the provided callback does not match the expected arg type")
			log.Error(err)
			return false
		}

		modifier := func(value eventBusData) eventBusData {
			value.callbacks = append(value.callbacks, cbv)
			return value
		}
		bus.data.Update(etype, modifier)
	}

	log.Debug("Bound func to event type : ", etype, trace)

	if current.recent != nil {
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
	trace := trace()
	if await {
		return bus.publishLogic(e, trace)
	} else {
		go bus.publishLogic(e, trace)
	}

	return true
}

func (bus *eventBus) publishLogic(e Event, trace string) bool {

	current, ok := bus.data.Load(e.Type)
	if !ok {
		// no previous bind/publish
		callbacks := []reflect.Value{}
		cbSig := getFSignature(e.Data)
		waitlist := []chan bool{}
		bus.data.Store(e.Type, eventBusData{callbacks, waitlist, &e, cbSig})
	} else {
		modifier := func(value eventBusData) eventBusData {
			value.recent = &e
			return value
		}
		bus.data.Update(e.Type, modifier)
	}

	// execute all callbacks for this event
	log.Debug("Publish Event", e, " to ", len(current.callbacks), " callbacks ", trace)
	if current.callbacks != nil {
		for _, cbv := range current.callbacks {
			// check if event data matches expected callback arg type
			argType := getFSignature(e.Data)
			if argType != current.cbType {
				err := errors.New("event data type does not match callback arg type")
				log.Error(err)
				return false
			}

			in := []reflect.Value{}
			if e.Data != nil {
				arg := reflect.ValueOf(e.Data)
				in = append(in, arg)
			}
			cbv.Call(in)
		}
	}

	// notify all waiting processes
	if current.waitlist != nil {
		for _, w := range current.waitlist {
			w <- true
			close(w)
		}
		modifier := func(value eventBusData) eventBusData {
			value.waitlist = value.waitlist[:0]
			return value
		}
		bus.data.Update(e.Type, modifier)
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

// get the last caller file that is a user file (to avoid assembly source files
// and such) and not eventbus.go
func trace() string {
	if *log.LogLvlFlag < int(log.DebugLevel) {
		return ""
	}

	pc := make([]uintptr, 100) // TODO : this fixed value of 100 is not optimal
	n := runtime.Callers(0, pc)
	frames := runtime.CallersFrames(pc[:n])

	f, more := frames.Next()
	for more {
		f, more = frames.Next()
		fileName := filepath.Base(f.File)
		isEB := strings.HasSuffix(fileName, "eventbus.go")
		isUserCode := strings.HasPrefix(f.File, "/Users")
		if isUserCode && !isEB {
			trace := "\nCalled from  : " + f.File + ":" + strconv.Itoa(f.Line)
			return trace
		}
	}

	return ""
}
