package ircbanchogo

import (
	"sync"
	"testing"
	"time"
)

func TestEventEmitter_Emit(t *testing.T) {
	wg := sync.WaitGroup{}
	waitCh := make(chan struct{})
	e := &EventEmitter{}

	wg.Add(3)

	e.On("test", func(message string) {
		defer wg.Done()
		t.Log(message)
	})

	e.On("test", func(message string) {
		defer wg.Done()
		t.Log(message)
	})

	e.On("test3", func() {
		defer wg.Done()
		t.Log("This is a test message from callback without arguments")
	})

	// Since irc bancho go doesn't use callbacks with variadic functions it is not critical to fix it right now

	//e.On("test2", func(dd int, something ...interface{}) {
	//	defer wg.Done()
	//	t.Logf("This is a test message with types %T and %T", dd, something)
	//})

	e.Listen()
	defer e.Close()

	go func() {
		wg.Wait()
		close(waitCh)
	}()

	e.Emit("test", "this is a test message inside test callback")
	e.Emit("test3")

	select {
	case <-waitCh:
	case <-time.After(time.Second):
		t.Fatal("wait group timed out")
	}
}

func TestEventEmitter_Close(t *testing.T) {
	e := EventEmitter{}

	e.On("test", func() {})
	e.Listen()
	e.Emit("test")
	e.Emit("test")
	e.Close()
	select {
	case <-e.End:
	case <-time.After(2 * time.Second):
		t.Fatal("channel did not close")
	}
}

func TestEventEmitter_Add(t *testing.T) {
	nonValidParameters := []interface{}{
		"",
		0,
		struct{}{},
	}
	e := &EventEmitter{}

	for _, p := range nonValidParameters {
		_, err := e.On("test", p)
		if err == nil {
			t.Errorf("error does not returned when passed %s as callback", p)
		}
	}

	_, err := e.On("test", func(message string) {
	})
	_, err = e.On("test", func(dd int, something ...interface{}) {
	})

	if err != nil {
		t.Fatal("error fired when it is not supposed to")
	}

	if e.CallbackID == 0 {
		t.Error("callback id does not increment")
	}

}
