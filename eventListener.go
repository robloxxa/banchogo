package ircbanchogo

import (
	"errors"
	"reflect"
	"sync"
)

type EventName string

type CallbackID int

type Event struct {
	Name   string
	Values []interface{}
}

type EventEmitter struct {
	EventChan chan Event

	mu            sync.RWMutex
	CallbackID    CallbackID
	CallbackPairs map[string][]CallbackID
	callbacks     map[CallbackID]interface{}

	End chan bool
}

func (e *EventEmitter) Listen() {
	e.EventChan = make(chan Event)
	e.End = make(chan bool)
	e.handleEvents()
}

func (e *EventEmitter) handleEvents() {
	for {
		select {
		case event := <-e.EventChan:
			e.mu.RLock()

			pairs, ok := e.CallbackPairs[event.Name]
			if !ok {
				e.mu.RUnlock()
				break
			}

			var arguments []reflect.Value

			if len(event.Values) > 0 {
				for _, v := range event.Values {
					arguments = append(arguments, reflect.ValueOf(v))
				}
			}

			for _, v := range pairs {
				callback := reflect.ValueOf(e.callbacks[v])
				// TODO: Make callbacks work with variadic functions (e.g. func(smh, ...something))
				if callback.Type().NumIn() != len(arguments) {
					continue
				}
				callback.Call(arguments)
			}
			e.mu.RUnlock()
		case <-e.End:
			return
		}
	}
}

func (e *EventEmitter) Close() {
	close(e.End)
	e.EventChan = nil
}

func (e *EventEmitter) on(eventName string, callback interface{}) CallbackID {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.CallbackPairs == nil {
		e.CallbackPairs = make(map[string][]CallbackID)
	}
	if e.callbacks == nil {
		e.callbacks = make(map[CallbackID]interface{})
	}

	_, ok := e.CallbackPairs[eventName]
	if !ok {
		e.CallbackPairs[eventName] = make([]CallbackID, 0)
	}

	e.callbacks[e.CallbackID] = callback
	e.CallbackPairs[eventName] = append(e.CallbackPairs[eventName], e.CallbackID)
	e.CallbackID++

	return e.CallbackID
}

func (e *EventEmitter) On(eventName string, callback interface{}) (CallbackID, error) {
	if reflect.TypeOf(callback).Kind() != reflect.Func {
		return -1, errors.New("callback isn't a function")
	}

	return e.on(eventName, callback), nil
}

//func (e *EventEmitter) removeEventListenerById(callbackId CallbackID) {
//	e.mu.Lock()
//	defer e.mu.Unlock()
//	for _, pairs := range e.CallbackPairs {
//		for _, id := range pairs {
//			if callbackId == id {
//
//			}
//		}
//	}
//}

func (e *EventEmitter) removeAllEventListeners() {
	e.mu.Lock()
	e.CallbackPairs = make(map[string][]CallbackID)
	e.callbacks = make(map[CallbackID]interface{})
	e.CallbackID = 0
	e.mu.Unlock()
}

func (e *EventEmitter) Emit(eventName string, values ...interface{}) {
	if e.EventChan == nil {
		return
	}
	e.EventChan <- Event{eventName, values}
}
