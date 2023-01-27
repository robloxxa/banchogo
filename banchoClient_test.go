package ircbanchogo

import (
	"fmt"
	"testing"
	"time"
)

func TestNewBanchoClient(t *testing.T) {
	b := NewBanchoClient("robloxxa", "b6529d77")

	b.OnConnect(func() {
		fmt.Println("Connected!")
	})

	b.Connect()
}

func TestBanchoClient_Connect(t *testing.T) {
	b := NewBanchoClient("robloxxa", "asd")
	b.Timeout = 10 * time.Second
	b.OnStateChanged(func(state ConnectState) {
		fmt.Println(state)
	})

	b.OnRawMessage(func(message string) {
		fmt.Println(message)
	})

	err := b.Connect()
	if err == nil {
		t.Fatal("error doesn't occur when it should")
	}
}
