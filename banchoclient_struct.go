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

type BanchoMessageType int8

const (
	PM BanchoMessageType = iota
	CM
)

type BanchoUser struct {
	Username string
	// TODO

}

type BanchoChannel struct {
	Id      string
	Members map[string]*BanchoChannelMember
}

type BanchoPMMessage struct {
	// TODO
	User    BanchoUser
	Channel BanchoChannel
	Type    BanchoMessageType
	Message string

	Raw string
}

type BanchoClient struct {
	Username string
	Password string
	Host     string
	Port     string

	Timeout          time.Duration
	ReconnectTimeout time.Duration

	Users    map[string]BanchoUser
	Channels map[string]BanchoChannel

	onRawMessage func(string)
	//onMessage       func(BanchoMessage)
	onDirectMessage func()
	onError         func(error)

	conn *net.Conn

	stateMutex   sync.RWMutex
	connectState ConnectState

	eventMutex    sync.RWMutex
	callbackID    CallbackID
	callbackPairs []CallbackPair
	callbacks     map[CallbackID]interface{}

	welcomeChan chan bool
	done        chan bool
}

func NewBanchoClient(username, password string) *BanchoClient {
	return &BanchoClient{
		Username: username,
		Password: password,
	}
}

func (b *BanchoClient) Connect() error {
	if b.Username == "" || b.Password == "" {
		return errors.New("you should give username and password")
	}

	b.conn = nil
	b.done = nil
	b.welcomeChan = nil

	{
		if b.Host == "" {
			b.Host = BANCHOHOST
		}
		if b.Port == "" {
			b.Port = BANCHOPORT
		}
		if b.Timeout == 0 {
			b.Timeout = 1 * time.Minute
		}
		if b.ReconnectTimeout == 0 {
			b.ReconnectTimeout = 5 * time.Minute
		}

	}

	conn, err := net.DialTimeout("tcp", b.Host+":"+b.Port, b.Timeout)

	if err != nil {
		return err
	}

	b.conn = &conn
	b.welcomeChan = make(chan bool, 1)
	b.done = make(chan bool)

	go b.handleIrcMessages()

	err = b.send("PASS " + b.Password)
	err = b.send("USER " + b.Username + " 0 * :" + b.Username)
	err = b.send("NICK " + b.Username)

	if err != nil {
		return err
	}
	b.setConnectState(Connecting)
	for {
		select {
		case <-b.welcomeChan: // TODO: понять где менять стейт на коннектед
			b.setConnectState(Connected)
		case <-b.done:
			return nil
		}
	}
}

func (b *BanchoClient) Disconnect() {
	if b.IsDisconnected() {

	} else {
		b.send("QUIT")
		close(b.done)
		//if b.onDisconnect() != nil {
		//	b.onDisconnect()
		//}
	}
}

func (b *BanchoClient) send(message string) error {
	if b.IsDisconnected() || b.IsReconnecting() {
		//if b.onError != nil {
		//	b.onError(errors.New("can't send messages while disconnected or reconnecting"))
		//}
		return nil
	}
	_, err := fmt.Fprintf(*b.conn, "%s\r\n", message)
	if err != nil {
		return err
	}
	return nil
}

func (b *BanchoClient) handleIrcMessages() {
	msgChan := make(chan string)
	errChan := make(chan error)

	go readFromConn(b.conn, msgChan, errChan, b.done)

	for {
		select {
		case message := <-msgChan:
			//TODO: Написать парсер сообщения
			fmt.Println(message)
		case <-b.done:
			return
		}
	}
}

func readFromConn(conn *net.Conn, msgChan chan<- string, errChan chan<- error, done <-chan bool) {
	tp := textproto.NewReader(bufio.NewReader(*conn))
	for {
		message, err := tp.ReadLine()
		if err != nil {
			select {
			case msgChan <- message:
			case <-done:
				return
			}
		} else {
			// TODO: написать хендлер получше (или узнать оставить ли его таким)
			select {
			case errChan <- err:
			case <-done:
			}
			return
		}

	}
}

func (b *BanchoClient) getConnectState() ConnectState {
	b.stateMutex.RLock()
	defer b.stateMutex.RUnlock()
	return b.connectState
}

func (b *BanchoClient) setConnectState(state ConnectState) {
	b.stateMutex.Lock()
	b.connectState = state
	b.stateMutex.Unlock()
}

func (b *BanchoClient) IsDisconnected() bool {
	return b.getConnectState() == Disconnected
}

func (b *BanchoClient) IsReconnecting() bool {
	return b.getConnectState() == Reconnecting
}

func (b *BanchoClient) IsConnecting() bool {
	return b.getConnectState() == Connecting
}

func (b *BanchoClient) IsConnected() bool {
	return b.getConnectState() == Connected
}
