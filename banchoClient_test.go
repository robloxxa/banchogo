package banchogo

import (
	"fmt"
	"testing"
)

func TestNewBanchoClient(t *testing.T) {
	b := NewBanchoClient("robloxxa", "b6529d77")

	b.OnConnect(func() {
		fmt.Println("Connected!!!")
	})

	b.Connect()
}
