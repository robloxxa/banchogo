package banchogo

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"net/textproto"
	"strings"
	"sync"
	"time"
)

type Sender interface {
	SendMessage(string)
	SendAction(string)
}

type ConnectState uint8

const (
	Disconnected ConnectState = iota
	Reconnecting
	Connecting
	Connected
	Quiting
)

const (
	BANCHOHOST = "irc.ppy.sh"
	BANCHOPORT = "6667"
)

var (
	MissingCredentials = errors.New("username or password fields are empty strings, please provide credentials")
)

type BanchoClient struct {
	Username string
	Password string
	Host     string
	Port     string

	Timeout          time.Duration
	ReconnectTimeout time.Duration
	KeepAlive        time.Duration

	Users    map[string]BanchoUser
	Channels map[string]BanchoChannel

	Event *EventEmitter

	conn net.Conn

	stateMutex   sync.RWMutex
	connectState ConnectState

	welcomeChan chan bool
	done        chan struct{}
}

func NewBanchoClient(username, password string) *BanchoClient {
	return &BanchoClient{
		Username: username,
		Password: password,
	}
}

func (b *BanchoClient) Connect() (err error) {

	if b.Username == "" || b.Password == "" {
		return MissingCredentials
	}

	b.conn = nil
	b.done = nil

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

	conn, err := net.Dial("tcp", b.Host+":"+b.Port)
	if err != nil {
		return err
	}

	b.conn = conn
	b.done = make(chan struct{})
	b.welcomeChan = make(chan bool, 1)

	go b.handleIrcMessages()

	b.Event.Listen()

	defer func() {
		if err != nil {
			close(b.done)
			b.setConnectState(Disconnected)
			b.Event.Emit("error", err)
		}
	}()

	b.Send("PASS " + b.Password)
	b.Send("USER " + b.Username + " 0 * :" + b.Username)
	b.Send("NICK " + b.Username)

	b.setConnectState(Connecting)

	// TODO: Make a Done or Loop function that will handle exit and reconnect
	timeout := time.NewTimer(b.Timeout)
	select {
	case <-b.welcomeChan:
	case <-b.done:
		err = errors.New("client disconnected")
	case <-timeout.C:
		err = errors.New("server timed out")
	}
	return
}

func (b *BanchoClient) close() {
	close(b.done)
	if b.conn != nil {
		_ = b.conn.Close()
	}
}

func (b *BanchoClient) Loop() {
	defer func() {
		b.setConnectState(Disconnected)
		b.Event.Close()
	}()
	for {
		<-b.done
		if b.IsQuiting() {
			return
		}

	}
}

func (b *BanchoClient) Disconnect() {

}

func (b *BanchoClient) Send(message string) {
	if b.IsDisconnected() || b.IsReconnecting() {
		b.Event.Emit("error", errors.New("sending message while disconnected or reconnecting"))
		return
	}
	_, err := fmt.Fprintf(b.conn, "%s\r\n", message)
	if err != nil {
		b.Event.Emit("error", err)
		return
	}
	return
}

func (b *BanchoClient) handleIrcMessages() {
	msgChan := make(chan string)
	errChan := make(chan error)

	go readFromConn(b.conn, msgChan, errChan, b.done)

	for {
		select {
		case message := <-msgChan:
			b.Event.Emit("rawMessage", message)
			splits := strings.Split(message, " ")

			if splits[0] == "PING" {
				b.Event.Emit("PING")
				b.Send("PONG " + strings.Join(append(splits[:0], splits[1:]...), " "))
			}

			ircHandler, ok := ircHandlers[splits[1]]
			if ok {
				ircHandler(b, splits)
			}
		case err := <-errChan:
			b.Event.Emit("error", err) // TODO: not sure if i emit it here
			b.close()
			return
		case <-b.done:
			return
		}
	}
}

func readFromConn(conn net.Conn, msgChan chan<- string, errChan chan<- error, done <-chan struct{}) {
	tp := textproto.NewReader(bufio.NewReader(conn))
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

func (b *BanchoClient) getConnectState() ConnectState {
	b.stateMutex.RLock()
	defer b.stateMutex.RUnlock()
	return b.connectState
}

func (b *BanchoClient) setConnectState(state ConnectState) {
	b.stateMutex.Lock()
	b.Event.Emit("changeState", state)
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

func (b *BanchoClient) IsQuiting() bool {
	return b.getConnectState() == Quiting
}
