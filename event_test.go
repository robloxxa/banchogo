package banchogo

import (
	"sync"
	"testing"
	"time"
)

func testPanic(t *testing.T, f func()) {
	defer func() {
		recover()
	}()

	f()

	t.Error("function doesn't panic when it should")
}

func TestEventEmitter_On(t *testing.T) {
	e := &EventEmitter{}

	testPanic(t, func() {
		e.On("test", func(error, error, error, error) {})
	})

	d := e.On("test", func() {})

	d()

	if len(e.handlers["test"]) > 0 {
		t.Error("delete function didn't delete handler")
	}
}

func TestEventEmitter_Emit(t *testing.T) {
	e := &EventEmitter{}
	wg := sync.WaitGroup{}
	done := make(chan struct{})

	wg.Add(3)
	e.On("test", func() {
		wg.Done()
	})
	e.On("test", func() {
		wg.Done()
	})
	e.On("test", func() {
		wg.Done()
	})
	go func() {
		wg.Wait()
		close(done)
	}()
	go e.Emit("test")
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Error("not all events was fired")
	}
}

func TestEventEmitter_Once(t *testing.T) {
	e := &EventEmitter{}
	wg := sync.WaitGroup{}

	wg.Add(1)
	e.Once("test", func() {
		wg.Done()
	})
	e.Emit("test")
	e.Emit("test")

}
