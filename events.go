package main

import (
	"reflect"
)

type CallbackID int

func (b *BanchoClient) addEvent(eventName string, callback interface{}) CallbackID {
	if reflect.TypeOf(callback).Kind() != reflect.Func {
		panic("callback is not a function!")
	}
	b.eventMutex.Lock()
	defer b.eventMutex.Unlock()
	_, ok := b.callbackPairs[eventName]
	if !ok {
		b.callbackPairs[eventName] = make([]CallbackID, 0)
	}
	b.callbacks[b.callbackID] = callback
	b.callbackPairs[eventName] = append(b.callbackPairs[eventName], b.callbackID)
	b.callbackID++
	return b.callbackID
}

//func (b *BanchoClient) removeEventListenerById(callbackId CallbackID) {
//	b.eventMutex.Lock()
//	defer b.eventMutex.Unlock()
//	for _, pairs := range b.callbackPairs {
//		for _, id := range pairs {
//			if callbackId == id {
//
//			}
//		}
//	}
//}

func (b *BanchoClient) removeAllEventListeners() {
	b.eventMutex.Lock()
	b.callbackPairs = make(map[string][]CallbackID)
	b.callbacks = make(map[CallbackID]interface{})
	b.eventMutex.Unlock()
}

func (b *BanchoClient) emitEvent(eventName string, values ...interface{}) {
	cbPairs, ok := b.callbackPairs[eventName]
	if !ok {
		return
	}

	var arguments []reflect.Value

	if len(values) > 0 {
		for _, v := range values {
			arguments = append(arguments, reflect.ValueOf(v))
		}
	}

	b.eventMutex.RLock()
	defer b.eventMutex.RUnlock()

	for _, v := range cbPairs {
		callback := reflect.ValueOf(b.callbacks[v])
		if callback.Type().NumIn() != len(arguments) {
			continue
		}
		go callback.Call(arguments)
	}
}
