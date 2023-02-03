package banchogo

import (
	"strings"
	"sync"
)

type EventEmitter struct {
	AsyncEvents bool

	handlersMu   sync.RWMutex
	handlers     map[string][]*EventHandlerInstance
	handlersOnce map[string][]*EventHandlerInstance
}

type EventHandler interface {
	Call(...interface{})
	NumField() int
}

type EventHandlerInstance struct {
	eventHandler EventHandler
}

func (e *EventEmitter) addHandler(name string, eh EventHandler) func() {
	e.handlersMu.Lock()
	defer e.handlersMu.Unlock()

	if e.handlers == nil {
		e.handlers = map[string][]*EventHandlerInstance{}
	}

	ehi := &EventHandlerInstance{eh}
	e.handlers[name] = append(e.handlers[name], ehi)

	return func() {
		e.removeEventHandlerInstance(name, ehi)
	}
}

func (e *EventEmitter) addHandlerOnce(name string, eh EventHandler) func() {
	e.handlersMu.Lock()
	defer e.handlersMu.Unlock()

	if e.handlersOnce == nil {
		e.handlersOnce = map[string][]*EventHandlerInstance{}
	}
	ehi := &EventHandlerInstance{eh}
	e.handlersOnce[name] = append(e.handlersOnce[name], ehi)

	return func() {
		e.removeEventHandlerInstance(name, ehi)
	}
}

func (e *EventEmitter) removeEventHandlerInstance(name string, ehi *EventHandlerInstance) {
	e.handlersMu.Lock()
	defer e.handlersMu.Unlock()

	handlers := e.handlers[name]
	for i := range handlers {
		if handlers[i] == ehi {
			e.handlers[name] = append(handlers[:i], handlers[i+1:]...)
		}
	}

	handlersOnce := e.handlersOnce[name]
	for i := range handlersOnce {
		if handlersOnce[i] == ehi {
			e.handlersOnce[name] = append(handlersOnce[:i], handlersOnce[i+1:]...)
		}
	}
}

func (e *EventEmitter) handle(name string, params ...interface{}) {
	e.handlersMu.RLock()
	defer e.handlersMu.RUnlock()

	name = strings.ToLower(name)
	for _, eh := range e.handlers[name] {
		if e.AsyncEvents {
			go eh.eventHandler.Call(params...)
		} else {
			eh.eventHandler.Call(params...)
		}
	}

	for _, eh := range e.handlersOnce[name] {
		if e.AsyncEvents {
			go eh.eventHandler.Call(params...)
		} else {
			eh.eventHandler.Call(params...)
		}
		e.handlersOnce[name] = nil
	}
}

func (e *EventEmitter) AddHandler(name string, handler interface{}) func() {
	// Figure out how to panic if event name != handler

	eh := interfaceToEventHandler(handler)

	if eh == nil {
		// Maybe we should return error instead of panicing
		panic("trying to add a non existing handler")
	}

	return e.addHandler(strings.ToLower(name), eh)
}

func (e *EventEmitter) AddHandlerOnce(name string, handler interface{}) func() {
	// Figure out how to panic if event name != handler

	eh := interfaceToEventHandler(handler)

	if eh == nil {
		// Maybe we should return error instead of panicing
		panic("trying to add a non existing handler")
	}

	return e.addHandlerOnce(strings.ToLower(name), eh)
}
