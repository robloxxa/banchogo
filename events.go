package main

import "reflect"

type CallbackID int

type CallbackPair struct {
	Id   CallbackID
	Name string
}

func (b *BanchoClient) addEvent(eventName string, callback interface{}) CallbackID {
	if reflect.TypeOf(callback).Kind() != reflect.Func {
		panic("callback is not a function!")
	}
	b.eventMutex.Lock()
	defer b.eventMutex.Unlock()
	b.callbackPairs = append(b.callbackPairs, CallbackPair{b.callbackID, eventName})
	b.callbacks[b.callbackID] = callback
	b.callbackID++
	return b.callbackID
}

func (b *BanchoClient) runEvent(callbackName string, values ...interface{}) {
	var arguments []reflect.Value
	if len(values) > 0 {
		for _, v := range values {
			arguments = append(arguments, reflect.ValueOf(v))
		}
	}
	for _, v := range b.callbackPairs {
		if v.Name != callbackName {
			continue
		}
		reflect.ValueOf(b.callbacks[v.Id]).Call(arguments)
	}
}
