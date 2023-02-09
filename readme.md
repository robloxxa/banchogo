# banchogo

This module is currently WIP

Inspired by [banchojs](https://github.com/ThePooN/bancho.js)

## TODO
- [ ] Multiplayer support
- [ ] Better Reconnect mechanism
- [ ] Tests

## Getting Started
Install module with `go get github.com/robloxxa/banchogo`

```go
package main

import (
	"fmt"
	"github.com/robloxxa/banchogo"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	client := banchogo.NewBanchoClient(banchogo.ClientOptions{
		Username: "robloxxa",
		Password: "irc_password",
	})

	client.OnceConnect(func() {
		fmt.Println("Connected!")
	})

	client.OnMessage(func(msg *banchogo.PrivateMessage) {
		fmt.Println(msg.User.Name() + ": " + msg.Content())
	})

	err := client.Connect()
	if err != nil {
		panic(err)
	}
    
	// Banchogo launch their own goroutines and doesn't block main one.
	// We can block main goroutine by waiting for a CTRL-C input
	// OR
	// We can use <-client.Done channel which will be closed on Disconnect
	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
    // Gracefully shutting down, stopping all goroutines and properly disconnect from IRC
	client.Disconnect()
}
```

## Event Problem
Since this module aimed to be easy and similar to banchojs, a banchogo Event system is synchronous like Node.js.

This can be problematic if you use methods like `user.Where(), user.Whois(), user.Stats(), etc.`, which block reading goroutine until it hits timeout while you wait response from channel.
```go
client.OnMessage(func(m *PrivateMessage) {
	data := <-m.User.Where()
	// ...it will hang here until timeout
	// this is because response is ahead and can't be obtained in same goroutine where event emitted
}) 
```
But it is nothing to worry about, just run these methods in separate goroutine
```go
client.OnMessage(func(m *PrivateMessage) {
	go func() {
        data := <-m.User.Where()
		fmt.Println("yay data: ", data)
	}()
}) 
```
Basically if you see that method returns a go channel, you should consider calling it in separate from emitted event goroutine
