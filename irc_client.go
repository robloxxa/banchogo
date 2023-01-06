package main

import (
	"errors"
	"fmt"
	"net"
)

type ConnectState int8

const (
	Disconnected ConnectState = iota
	Connecting
	Reconnecting
	Connected
)

type BanchoClient struct {
	Host         string
	Port         string
	Username     string
	Password     string
	connectState ConnectState
	reconnectTimeout
	client *net.Conn
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
	b.connectState = state
}

func (b *BanchoClient) GetConnectState() ConnectState {
	return b.connectState
}

func (b *BanchoClient) IsConnected() bool {
	return b.connectState == Connected
}

func (b *BanchoClient) IsConnecting() bool {
	return b.connectState == Connecting
}

func (b *BanchoClient) IsDisconnected() bool {
	return b.connectState == Disconnected
}

func (b *BanchoClient) IsReconnecting() bool {
	return b.connectState == Reconnecting
}

func (b *BanchoClient) Connect() error {
	conn, err := net.Dial("tcp", b.Host+":"+b.Port)
	if err != nil {
		return err
	}
	if b.Username == "" || b.Password == "" {
		return errors.New("you should provide credentials, dumb dumb")
	}

	b.send("USER " + b.Username)
	b.send("PASS " + b.Password)

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
