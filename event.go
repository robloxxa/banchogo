package banchogo

import (
	"strings"
	"sync"
)

type EventEmitter struct {
	handlersMu sync.Mutex
	handlers   map[string][]*EventHandlerInstance
}

type EventHandler interface {
	Call(...interface{})
	NumField() int
}

type EventHandlerInstance struct {
	once         *sync.Once
	eventHandler EventHandler
}

func (e *EventEmitter) on(name string, once bool, eh EventHandler) func() {
	e.handlersMu.Lock()
	defer e.handlersMu.Unlock()

	if e.handlers == nil {
		e.handlers = map[string][]*EventHandlerInstance{}
	}

	ehi := &EventHandlerInstance{eventHandler: eh}
	if once {
		ehi.once = &sync.Once{}
	}

	e.handlers[name] = append(e.handlers[name], ehi)

	return func() {
		e.off(name, ehi)
	}
}

func (e *EventEmitter) off(name string, ehi *EventHandlerInstance) {
	e.handlersMu.Lock()
	defer e.handlersMu.Unlock()

	handlers := e.handlers[name]
	for i := range handlers {
		if handlers[i] == ehi {
			e.handlers[name] = append(handlers[:i], handlers[i+1:]...)
		}
	}
}

func (e *EventEmitter) emit(name string, params ...interface{}) {
	e.handlersMu.Lock()
	if e.handlers == nil {
		e.handlers = make(map[string][]*EventHandlerInstance)
	}
	handlers, ok := e.handlers[name]
	e.handlersMu.Unlock()

	if ok {
		for _, eh := range handlers {
			if eh.once != nil {
				eh.once.Do(func() {
					eh.eventHandler.Call(params...)
					e.off(name, eh)
				})
			} else {
				eh.eventHandler.Call(params...)
			}
		}
	}

}

func (e *EventEmitter) Emit(name string, params ...interface{}) {
	e.emit(strings.ToLower(name), params...)
}

func (e *EventEmitter) On(name string, handler interface{}) func() {
	// Figure out how to panic if event name != handler

	eh := interfaceToEventHandler(handler)

	if eh == nil {
		// Maybe we should return error instead of panicing
		panic("trying to add a non existing handler")
	}

	return e.on(strings.ToLower(name), false, eh)
}

func (e *EventEmitter) Once(name string, handler interface{}) func() {
	// Figure out how to panic if event name != handler

	eh := interfaceToEventHandler(handler)

	if eh == nil {
		// Maybe we should return error instead of panicing
		panic("trying to add a non existing handler")
	}

	return e.on(strings.ToLower(name), true, eh)
}

func (e *EventEmitter) RemoveAllListeners(name string) {
	e.handlersMu.Lock()
	e.handlers[name] = nil
	e.handlersMu.Unlock()
}
