package banchogo

import (
	"github.com/puzpuzpuz/xsync/v2"
	"strings"
)

type event struct {
	Name string
	Args []interface{}
}

type EventEmitter struct {
	handlers     *xsync.MapOf[string, []*EventHandlerInstance]
	handlersOnce *xsync.MapOf[string, []*EventHandlerInstance]

	eventQueue chan event
	done       chan struct{}
}

type EventHandler interface {
	Call(...interface{})
	NumField() int
}

type EventHandlerInstance struct {
	eventHandler EventHandler
}

func NewEmitter() *EventEmitter {
	return &EventEmitter{
		handlers:     xsync.NewMapOf[[]*EventHandlerInstance](),
		handlersOnce: xsync.NewMapOf[[]*EventHandlerInstance](),
	}
}

func (e *EventEmitter) addHandler(name string, eh EventHandler) func() {
	if e.handlers == nil {
		e.handlers = xsync.NewMapOf[[]*EventHandlerInstance]()
	}

	ehi := &EventHandlerInstance{eh}
	_, ok := e.handlers.Compute(name, func(oldValue []*EventHandlerInstance, loaded bool) (newValue []*EventHandlerInstance, delete bool) {
		newValue = append(oldValue, ehi)
		delete = false
		return
	})

	if !ok {
		panic("sync map failed to store a value in handlers")
	}

	return func() {
		e.removeEventHandlerInstance(name, ehi)
	}
}

func (e *EventEmitter) addHandlerOnce(name string, eh EventHandler) func() {
	//if e.handlersOnce == nil {
	//	e.handlersOnce =
	//}

	ehi := &EventHandlerInstance{eh}

	_, ok := e.handlersOnce.Compute(name, func(oldValue []*EventHandlerInstance, loaded bool) (newValue []*EventHandlerInstance, delete bool) {
		newValue = append(oldValue, ehi)
		delete = false
		return
	})

	if !ok {
		panic("sync map failed to store a value in handlers")
	}

	return func() {
		e.removeEventHandlerInstance(name, ehi)
	}
}

func (e *EventEmitter) removeEventHandlerInstance(name string, ehi *EventHandlerInstance) {
	e.handlers.Compute(name, func(oldValue []*EventHandlerInstance, loaded bool) (newValue []*EventHandlerInstance, delete bool) {
		for i := range oldValue {
			if oldValue[i] == ehi {
				newValue = append(oldValue[:i], oldValue[i+1:]...)
				return
			}
		}
		return
	})

	e.handlersOnce.Compute(name, func(oldValue []*EventHandlerInstance, loaded bool) (newValue []*EventHandlerInstance, delete bool) {
		for i := range oldValue {
			if oldValue[i] == ehi {
				newValue = append(oldValue[:i], oldValue[i+1:]...)
				return
			}
		}
		return
	})
}

func (e *EventEmitter) handle(name string, params ...interface{}) {
	name = strings.ToLower(name)
	if handlers, ok := e.handlers.Load(name); ok {
		for _, eh := range handlers {
			eh.eventHandler.Call(params...)
		}
	}

	if handlersOnce, ok := e.handlersOnce.Load(name); ok {
		for _, eh := range handlersOnce {
			eh.eventHandler.Call(params...)
			e.handlersOnce.Delete(name)
		}
	}
}

func (e *EventEmitter) Handle(name string, params ...interface{}) {
	if e.done == nil {
		e.done = make(chan struct{})
		e.eventQueue = make(chan event) // TODO: maybe make it buffered, so sender won't block
		go e.Listen(e.eventQueue)
	}
	e.eventQueue <- event{name, params}
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

func (e *EventEmitter) Listen(eventQueue chan event) {
	for {
		select {
		case ev := <-eventQueue:
			e.handle(ev.Name, ev.Args...)
		case <-e.done:
			return
		}
	}
}

// Close method stops a Listen goroutine
func (e *EventEmitter) Close() {
	if e.done != nil {
		close(e.done)
		e.done = nil
	}
}
