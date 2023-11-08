package bus

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

var bus EventBus

func TestNewEventbus_Basic(t *testing.T) {
	bus = NewEventbus()
	if bus == nil {
		t.Error("Expected a non-nil EventBus, got nil")
	}

	basicEvt := EventType("basic")
	data := "data"

	testCallback := func(d string) {
		if d != data {
			t.Errorf("Transmitted event data doesn't match")
		}
	}

	bus.AwaitBind(basicEvt, testCallback)

	e := Event{basicEvt, data}
	bus.Publish(e)
}

func TestEventBus_Wrong_Publish_Arg_Type(t *testing.T) {
	wrongArgEvt := EventType("wrong-pub-arg")

	bus.AwaitBind(wrongArgEvt, func(d int) {})

	res := bus.AwaitPublish(Event{wrongArgEvt, "test"})
	if res != false {
		t.Errorf("Mismatching data types should have failed but not with a panic")
	}
}

func TestEventBus_Wrong_Bind_Arg_Type(t *testing.T) {
	wrongArgEvt := EventType("wrong-bind-arg")

	bus.AwaitPublish(Event{wrongArgEvt, "test"})

	res := bus.AwaitBind(wrongArgEvt, func(d int) {})
	if res != false {
		t.Errorf("Mismatching data types should have failed but not with a panic")
	}
}

func TestEventBus_Publish_Stress(t *testing.T) {
	countEvt := EventType("stress-test")
	counter := 0
	testCnt := 10
	mu := sync.Mutex{}

	bus.AwaitBind(countEvt, func(d int) {
		mu.Lock()
		counter++
		fmt.Println("counter ", counter)
		mu.Unlock()
	})

	for i := 0; i < testCnt; i++ {
		e := Event{countEvt, i}
		bus.Publish(e)
	}

	// await publishes, don't use AwaitPublish so we can inspect data races !
	time.Sleep(time.Second)

	mu.Lock()
	fmt.Println("counter ", counter)
	if counter != testCnt {
		t.Errorf("Data race on event publish detected, was %d but should've been %d", counter, testCnt)
	}
	mu.Unlock()
}

func TestEventBus_Publish_Nested(t *testing.T) {
	// test publish with a callback that publishes
	nestedEvt := EventType("nested-publish")
	bus.AwaitBind(nestedEvt, func() {
	})

	wrapperEvt := EventType("wrapper-nested-publish")
	bus.AwaitBind(wrapperEvt, func() {
		event := Event{nestedEvt, nil}
		bus.Publish(event)
	})

	e := Event{wrapperEvt, nil}
	bus.Publish(e)
}

func TestEventBus_Bind_Nested(t *testing.T) {
	nestedBindEvt := EventType("nested-bind")
	wrapperBindEvt := EventType("wrapper-nested-bind")

	nestedBindComplete := make(chan bool)
	bus.AwaitBind(wrapperBindEvt, func() {
		bus.Bind(nestedBindEvt, func() {
		})
		nestedBindComplete <- true
	})

	e := Event{wrapperBindEvt, nil}

	go bus.Publish(e)

	// wait for the nested Bind to complete or for a timeout to occur
	select {
	case <-time.After(time.Second):
		t.Error("Deadlock between Publish and nested Bind")
	case <-nestedBindComplete:
		return
	}
}

