package banchogo

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/puzpuzpuz/xsync/v2"
	"go.uber.org/ratelimit"
	"net"
	"net/textproto"
	"strings"
	"sync"
	"time"
)

type ConnectState int

const (
	Disconnected ConnectState = iota
	Reconnecting
	Connecting
	Connected
)

const (
	BANCHOHOST = "irc.ppy.sh"
	BANCHOPORT = "6667"
)

var (
	ErrMissingCredentials = errors.New("username or password field is empty string, please provide credentials")
	ErrBadAuthentication  = errors.New("bancho authentication failed")

	ErrMessageTimeout   = errors.New("message timeout")
	ErrConnectionClosed = errors.New("connection closed")

	ErrUserOffline = errors.New("user offline")
)

type MessageSender interface {
	Name() string
	SendMessage(string) <-chan error
	SendAction(string) <-chan error
}

type ClientOptions struct {
	Username string
	Password string
	Host     string
	Port     string

	BotAccount bool
	Reconnect  *bool

	RateLimiter ratelimit.Limiter

	// TODO: Check if we even need logger
	//Log *log.Logger
}

type Client struct {
	// TODO: Make documentation
	ev *EventEmitter

	Username string
	Password string
	Host     string
	Port     string

	// BotAccount set it to "true" if you have bot account https://osu.ppy.sh/wiki/en/Bot_account.
	// Used for initialising default values for RateLimiter and prevent sending messages to a public channel like #osu.
	// False by default
	BotAccount bool
	// Reconnect set to "false" if you don't want to reconnect after error
	Reconnect bool

	Timeout time.Duration

	// RateLimiter by default banchogo will use github.com/uber-go/ratelimit
	// Default ratelimiter use values from https://github.com/ThePooN/bancho.js/blob/dac8a2bd3e8ffca01fac6753759e68de651a9f5b/lib/BanchoClient.js#L88
	// You can initialize limiter with non-default values or use your own limiter that implements Limiter interface
	RateLimiter ratelimit.Limiter

	// TODO: check for data race when editing user/channel objects
	Users    *xsync.MapOf[string, *User]
	Channels *xsync.MapOf[string, *Channel]

	conn net.Conn

	stateMutex   sync.RWMutex
	connectState ConnectState

	wg              sync.WaitGroup
	messageQueue    chan *OutgoingMessage
	reconnectSignal chan struct{}
	connectSignal   chan error

	done chan struct{}
}

func NewBanchoClient(opt ClientOptions) (b *Client) {
	b = &Client{
		ev:         NewEmitter(),
		Username:   opt.Username,
		Password:   opt.Password,
		BotAccount: opt.BotAccount,
	}

	if opt.RateLimiter == nil {
		amount := 9
		duration := 12500 * time.Millisecond
		if b.BotAccount {
			amount = 298
			duration = 62500 * time.Millisecond
		}
		b.RateLimiter = ratelimit.New(amount, ratelimit.Per(duration))
		b.RateLimiter.Take() // Init Ratelimiter
	}

	return
}

func (b *Client) Connect() (err error) {
	if b.Username == "" || b.Password == "" {
		return ErrMissingCredentials
	}

	if b.IsConnected() || b.IsConnecting() {
		return errors.New("already connected/connecting")
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
	if b.Users == nil {
		b.Users = xsync.NewMapOf[*User]()
	}
	if b.Channels == nil {
		b.Channels = xsync.NewMapOf[*Channel]()
	}

	b.conn, err = net.Dial("tcp", b.Host+":"+b.Port)
	if err != nil {
		return err
	}

	b.done = make(chan struct{})
	b.connectSignal = make(chan error)
	b.messageQueue = make(chan *OutgoingMessage)

	b.wg.Add(2)
	go b.readIrcMessages(b.conn, b.done)
	go b.processMessages(b.messageQueue)

	defer func() {
		if err != nil {
			b.stop()
			b.setConnectState(Disconnected)
		}
	}()

	b.setConnectState(Connecting)

	b.Send("PASS " + b.Password)
	b.Send("USER " + b.Username + " 0 * :" + b.Username)
	b.Send("NICK " + b.Username)

	timeout := time.NewTimer(b.Timeout)
	select {
	case err = <-b.connectSignal:
	case <-b.done:
		err = errors.New("client disconnected")
	case <-timeout.C:
		err = errors.New("server timed out")
	}
	return
}

func (b *Client) stop() {
	if b.done != nil {
		close(b.done)
		b.done = nil
	}

	if b.reconnectSignal != nil {
		close(b.reconnectSignal)
		b.reconnectSignal = nil
	}

	if b.conn != nil {
		b.conn.Close()
		b.conn = nil
	}

	b.wg.Wait()

	b.Channels.Range(func(k string, u *Channel) bool {
		u.Joined = false
		return true
	})
}

func (b *Client) reconnect() {
	if b.IsReconnecting() {
		return
	}

	b.stop()

	for {
		err := b.Connect()

		if err == nil {
			return
		}

		b.ev.Handle("error", err)

		b.reconnectSignal = make(chan struct{})
		timer := time.NewTimer(5 * time.Second)

		b.setConnectState(Reconnecting)

		select {
		case <-timer.C:
			b.reconnectSignal = nil
		case <-b.reconnectSignal:
			return
		}
	}
}

// Disconnect method properly disconnects from irc.
// It Sends PART to all channels and QUIT on exit.
// Use a Close method if you want to stop all goroutines e.g. when graceful shutdown
func (b *Client) Disconnect() {
	if b.IsDisconnected() {
		return
	}
	defer func() {
		b.stop()
		b.setConnectState(Disconnected)
	}()
	// If we're currently reconnecting and waiting for a timeout,
	// we send a signal to shut down that goroutine for reconnecting

	// TODO: Send PART to all channels for clean disconnect

	b.Send("QUIT")
}

// Close method call Disconnect and stops all listening goroutines
func (b *Client) Close() {
	b.Disconnect()

	b.ev.Close()

	b.Channels.Range(func(_ string, ch *Channel) bool {
		//ch.ev.Close()
		return true
	})
	b.Users.Range(func(_ string, u *User) bool {
		u.ev.Close()
		return false
	})
}

func (b *Client) Send(message string) error {
	if b.IsDisconnected() || b.IsReconnecting() {
		return ErrConnectionClosed
	}
	_, err := fmt.Fprintf(b.conn, "%s\r\n", message)
	if err != nil {
		return err
	}
	b.ev.Handle("OnRawMessage", strings.Split(message, " "))
	return nil
}

func (b *Client) readIrcMessages(conn net.Conn, done <-chan struct{}) {
	tp := textproto.NewReader(bufio.NewReader(conn))
	for {
		content, err := tp.ReadLine()
		if err != nil {
			if err.Error() == "EOF" {
				b.ev.Handle("Error", ErrConnectionClosed)
			} else {
				b.ev.Handle("Error", err)
			}

			if b.conn == conn {
				go b.reconnect()
			}
			return
		}

		select {
		case <-done:
			return
		default:
			splits := strings.Split(content, " ")

			b.ev.Handle("RawMessage", splits)
			if splits[0] == "PING" {
				b.ev.Handle("PING")
				b.Send("PONG " + strings.Join(append(splits[:0], splits[1:]...), " "))
			}

			ircHandler, ok := ircHandlers[splits[1]]
			if ok {
				ircHandler(b, splits)
			}
		}

	}
}

func (b *Client) processMessages(messageQueue <-chan *OutgoingMessage) {
	defer func() {
		close(b.messageQueue)
		for msg := range b.messageQueue {
			msg.C <- ErrConnectionClosed
		}
		b.messageQueue = nil
	}()
	for {
		select {
		case msg := <-messageQueue:
			if !b.IsConnected() {
				msg.C <- errors.New("currently disconnected")
				break
			}

			if b.RateLimiter != nil {
				b.RateLimiter.Take()
			}

			name := TruncateString(strings.Split(msg.Name(), "\n")[0], 28)
			content := strings.Split(msg.Content, "\n")[0]

			err := b.Send(fmt.Sprintf("PRIVMSG %s :%s", name, content))
			if err != nil {
				msg.C <- err
				break
			}

			switch s := msg.MessageSender.(type) {
			case *User:
				b.ev.Handle("PrivateMessage", newPrivateMessage(b, b.GetSelf(), s, true, content))
			case *Channel:
				b.ev.Handle("ChannelMessage", newChannelMessage(b, b.GetSelf(), s, true, content))
			}

			msg.C <- nil
		case <-b.done:
			return
		}
	}
}

func (b *Client) GetChannel(channelName string) (channel *Channel, err error) {
	// TODO: MultiplayerChannels support
	if strings.Index(channelName, "#") != 0 || len(channelName) < 0 {
		return nil, errors.New("invalid channel name")
	}
	channel, ok := b.Channels.Load(channelName)
	if !ok {
		channel = NewBanchoChannel(b, channelName)
		b.Channels.Store(channelName, channel)
	}
	return
}

func (b *Client) GetUser(username string) (user *User) {
	username = strings.Replace(username, " ", "_", -1)
	user, ok := b.Users.Load(strings.ToLower(username))
	if !ok {
		user = newBanchoUser(b, username)
		b.Users.Store(strings.ToLower(username), user)
	}
	return user
}

func (b *Client) GetSelf() (user *User) {
	return b.GetUser(b.Username)
}

func (b *Client) getConnectState() ConnectState {
	b.stateMutex.RLock()
	defer b.stateMutex.RUnlock()
	return b.connectState
}

func (b *Client) setConnectState(state ConnectState) {
	b.stateMutex.Lock()
	b.connectState = state
	b.stateMutex.Unlock()

	if state == Connected {
		b.ev.Handle("Connect", state)
	}

	if state == Disconnected {
		b.ev.Handle("Disconnect")
	}

	b.ev.Handle("StateChanged", state)
}

func (b *Client) IsDisconnected() bool {
	return b.getConnectState() == Disconnected
}

func (b *Client) IsReconnecting() bool {
	return b.getConnectState() == Reconnecting
}

func (b *Client) IsConnecting() bool {
	return b.getConnectState() == Connecting
}

func (b *Client) IsConnected() bool {
	return b.getConnectState() == Connected
}
