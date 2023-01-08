package main

import (
	"errors"
	"fmt"
	"net"
	"sync"
)

type ConnectState int8

const (
	Disconnected ConnectState = iota
	Connecting
	Reconnecting
	Connected
)

type BanchoClientOptions struct {
	Host     string
	Port     string
	Username string
	Password string
}

type BanchoClient struct {
	options          BanchoClientOptions
	client           *net.Conn
	mu               sync.RWMutex
	connectState     ConnectState
	reconnectTimeout int
}

func (b *BanchoClient) send(message string) error {
	if b.IsConnected() || b.IsConnecting() {
		_, err := fmt.Fprintf(*b.client, "%s\r\n", message)
		if err != nil {
			return err
		}
		return nil
	}
	return errors.New("you can't send messages while being disconnected")

}

func (b *BanchoClient) setConnectState(state ConnectState) {
	b.mu.Lock()
	b.connectState = state
	b.mu.Unlock()
}

func (b *BanchoClient) GetConnectState() ConnectState {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.connectState
}

func (b *BanchoClient) IsConnected() bool {
	return b.GetConnectState() == Connected
}

func (b *BanchoClient) IsConnecting() bool {
	return b.GetConnectState() == Connecting
}

func (b *BanchoClient) IsDisconnected() bool {
	return b.GetConnectState() == Disconnected
}

func (b *BanchoClient) IsReconnecting() bool {
	return b.GetConnectState() == Reconnecting
}

func (b *BanchoClient) Connect() error {
	conn, err := net.Dial("tcp", b.options.Host+":"+b.options.Port)
	if err != nil {
		return err
	}
	if b.Username == "" || b.Password == "" {
		return errors.New("you should provide credentials, dumb dumb")
	}

	b.send("USER " + b.options.Username)
	b.send("PASS " + b.options.Password)

	b.client = &conn

	return nil
}

func (b *BanchoClient) Disconnect() error {
	if b.client == nil {
		return errors.New("you aren't connected")
	}

	if b.IsConnected() {
		_ = b.send("QUIT")
	} else if b.IsConnecting() {
		b.client = nil
	}

	return nil
}

func NewBanchoClient(username, password string) *BanchoClient {

	return &BanchoClient{
		Host:     "irc.ppy.sh",
		Port:     "6667",
		Username: username,
		Password: password,
	}
}
