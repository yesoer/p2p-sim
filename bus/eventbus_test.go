package bus

import (
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"
)

var bus *EventBus

func TestNewEventbus_Basic(t *testing.T) {
	bus = NewEventbus()
	if bus == nil {
		t.Error("Expected a non-nil EventBus, got nil")
	}
	if len(bus.Data) != 0 {
		t.Errorf("Expected an empty Data map, but it has %d elements", len(bus.Data))
	}

	basicEvt := EventType("basic")
	data := "data"

	testCallback := func(e Event) {
		d := e.Data.(string)
		if d != data {
			t.Errorf("Transmitted event data doesn't match")
		}
	}

	bus.Bind(basicEvt, testCallback)

	e := Event{basicEvt, data}
	bus.Publish(e)
}

func TestEventBus_Publish_Stress(t *testing.T) {
	countEvt := EventType("count")
	counter := 0
	testCnt := 1000
	mu := sync.Mutex{}

	bus.Bind(countEvt, func(e Event) {
		mu.Lock()
		counter++
		fmt.Println("counter ", counter)
		mu.Unlock()
	})

	for i := 0; i < testCnt; i++ {
		e := Event{countEvt, i}
		bus.Publish(e)
	}

	// sleep since publish executes as go rountine
	time.Sleep(time.Second * 1)
	mu.Lock()
	fmt.Println("counter ", counter)
	if counter != testCnt {
		t.Errorf("Data race on event publish detected, was %d but should've been %d", counter, testCnt)
	}
	mu.Unlock()
}

func TestEventBus_Publish_Nested(t *testing.T) {
	// test publish with a callback that publishes
	nestedEvt := EventType("nested")
	bus.Bind(nestedEvt, func(e Event) {
	})

	wrapperEvt := EventType("wrapper")
	bus.Bind(wrapperEvt, func(e Event) {
		event := Event{nestedEvt, nil}
		bus.Publish(event)
	})

	e := Event{wrapperEvt, nil}
	bus.Publish(e)
}

func TestEventBus_Publish_Order(t *testing.T) {
	// check whether order is preserved (across events, not for a
	//	singular event)
	evtTypeCnt := 10
	var types []EventType
	for i := 0; i < evtTypeCnt; i++ {
		s := strconv.Itoa(i)
		typ := EventType(s)
		types = append(types, typ)
	}

	cnt := 0
	res := struct {
		Fail bool
		Exp  int
		Rec  int
	}{false, -1, -1}
	cntMu := sync.Mutex{}
	for _, typ := range types {
		bus.Bind(typ, func(e Event) {
			cntMu.Lock()
			if cnt != e.Data.(int) {
				res.Fail = true
				res.Exp = cnt
				res.Rec = e.Data.(int)
			}
			cnt++
			cntMu.Unlock()
		})
	}

	for i, typ := range types {
		e := Event{typ, i}
		// time.Sleep(time.Millisecond * 100)
		bus.Publish(e)
	}

	time.Sleep(time.Second)
	cntMu.Lock()
	if res.Fail {
		t.Errorf("Expected %d got %d", res.Exp, res.Rec)
	}
	cntMu.Unlock()
}
