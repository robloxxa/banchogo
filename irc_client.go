package main

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"net/textproto"
	"sync"
	"time"
)

var (
	BANCHOHOST = "irc.ppy.sh"
	BANCHOPORT = "6667"
)

type BanchoClient struct {
	Username string
	Password string
	Host     string
	Port     string

	Timeout          time.Duration
	ReconnectTimeout time.Duration

	client *net.Conn

	OnMessage func(message string)

	stateMutex   sync.RWMutex
	quit         bool
	connectState ConnectState

	// Channels
	msgChan     chan string
	errChan     chan error
	welcomeChan chan bool
	done        chan bool
}

func (b *BanchoClient) send(message string) error {
	if !b.IsConnected() && !b.IsConnecting() {
		return errors.New("you can't send messages while being disconnected")
	}
	_, err := fmt.Fprintf(*b.client, "%s\r\n", message)
	if err != nil {
		return err
	}
	return nil

}

func (b *BanchoClient) setConnectState(state ConnectState) {
	b.stateMutex.Lock()
	b.connectState = state
	b.stateMutex.Unlock()
}

func (b *BanchoClient) GetConnectState() ConnectState {
	b.stateMutex.RLock()
	defer b.stateMutex.RUnlock()
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
	if !b.IsDisconnected() {
		return errors.New("client already running")
	}
	if b.Username == "" || b.Password == "" {
		return errors.New("you should provide credentials, dumb dumb")
	}

	if b.Host == "" {
		b.Host = BANCHOHOST
	}

	if b.Port == "" {
		b.Port = BANCHOPORT
	}

	if b.Timeout == 0 {
		b.Timeout = 1 * time.Minute
	}

	conn, err := net.DialTimeout("tcp", b.Host+":"+b.Port, b.Timeout)
	if err != nil {
		return err
	}

	b.client = &conn
	b.setConnectState(Connecting)

	b.welcomeChan = make(chan bool, 1)
	b.done = make(chan bool)

	err = b.send("PASS " + b.Password)
	err = b.send("USER " + b.Username + " 0 * :" + b.Username)
	err = b.send("NICK " + b.Username)

	if err != nil {
		return err
	}
	go b.handleIrcMessage()
	for {
		select {
		case <-b.welcomeChan:
			b.setConnectState(Connected)
		case <-b.done:
			return nil
		}
	}
}

func (b *BanchoClient) handleIrcMessage() {
	msgChan := make(chan string)
	errChan := make(chan error)

	go readFromIrc(b.client, msgChan, errChan, b.done)
	for {
		select {
		case msg := <-msgChan:

		case <-b.done:
			return
		}
	}
}

func readFromIrc(conn *net.Conn, msgChan chan<- string, errChan chan<- error, done <-chan bool) {
	tp := textproto.NewReader(bufio.NewReader(*conn))
	for {
		message, err := tp.ReadLine()
		if err == nil {
			select {
			case msgChan <- message:
			case <-done:
				return
			}
		} else {
			select {
			case errChan <- err:
			case <-done:
			}

			return
		}
	}
}

func (b *BanchoClient) Disconnect() error {
	if b.client == nil || b.IsDisconnected() {
		return errors.New("you aren't connected")
	}

	if b.IsConnected() {
		_ = b.send("QUIT")
	} else if b.IsConnecting() {
		b.client = nil
	}
	b.setConnectState(Disconnected)
	close(b.done)

	return nil
}

func NewBanchoClient(username, password string) *BanchoClient {

	return &BanchoClient{
		Username: username,
		Password: password,
	}
}
