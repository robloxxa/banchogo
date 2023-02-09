package banchogo

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/puzpuzpuz/xsync/v2"
	"github.com/thehowl/go-osuapi"
	"go.uber.org/ratelimit"
	"io"
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

	ErrUserOffline  = errors.New("user offline")
	ErrUserNotFound = errors.New("user not found")

	ErrChannelNotFound = errors.New("no such channel")
)

type ClientOptions struct {
	Username string
	Password string
	Host     string
	Port     string

	BotAccount bool
	Reconnect  *bool

	ApiKey string

	RateLimiter ratelimit.Limiter
}

type Client struct {
	// TODO: Make documentation
	ev EventEmitter

	Api osuapi.Client

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

	currentWhois *WhoisResponse

	messageQueue    chan *OutgoingMessage
	reconnectSignal chan struct{}
	connectSignal   chan error

	Done chan struct{}
}

func NewBanchoClient(opt ClientOptions) (b *Client) {
	b = &Client{
		Username:   opt.Username,
		Password:   opt.Password,
		BotAccount: opt.BotAccount,
	}

	if opt.RateLimiter == nil {
		var (
			amount   int
			duration time.Duration
		)

		if b.BotAccount {
			amount = 298
			duration = 62500 * time.Millisecond
		} else {
			amount = 9
			duration = 12500 * time.Millisecond
		}
		b.RateLimiter = ratelimit.New(amount, ratelimit.Per(duration))
		b.RateLimiter.Take() // Init Ratelimiter
	}

	if opt.ApiKey != "" {
		b.Api = *osuapi.NewClient(opt.ApiKey)
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
	if b.reconnectSignal == nil {
		b.reconnectSignal = make(chan struct{})
	}

	b.conn, err = net.Dial("tcp", b.Host+":"+b.Port)
	if err != nil {
		return err
	}

	b.Done = make(chan struct{})
	b.connectSignal = make(chan error)
	b.messageQueue = make(chan *OutgoingMessage)

	go b.readIrcMessages(b.conn, b.Done)
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

	select {
	case err = <-b.connectSignal:
	case <-b.Done:
		err = errors.New("client disconnected")
	case <-time.After(b.Timeout):
		err = errors.New("server timed out")
	}
	return
}

func (b *Client) stop() {
	select {
	case <-b.Done:
	default:
		close(b.Done)
	}

	select {
	case <-b.reconnectSignal:
	default:
		close(b.reconnectSignal)
	}

	if b.conn != nil {
		b.conn.Close()
	}
}

func (b *Client) reconnect() {
	// TODO: redesign reconnect mechanism
	if b.IsReconnecting() {
		return
	}

	select {
	case <-b.Done:
		return
	default:
	}

	b.Disconnect()
	for {
		err := b.Connect()

		if err == nil {
			return
		}

		b.ev.Emit("error", err)

		timer := time.NewTimer(5 * time.Second)
		b.setConnectState(Reconnecting)

		select {
		case <-timer.C:
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

	b.Channels.Range(func(_ string, c *Channel) bool {
		if c.Joined {
			emitPart(b, b.GetSelf(), c)
		}
		return true
	})

	b.Send("QUIT")
}

// Close method call Disconnect and stops all listening goroutines
func (b *Client) Close() {
	b.Disconnect()

}

func (b *Client) Send(format string, a ...any) error {
	if b.IsDisconnected() || b.IsReconnecting() {
		return ErrConnectionClosed
	}
	m := fmt.Sprintf(format, a...)
	_, err := fmt.Fprintf(b.conn, "%s\r\n", m)
	if err != nil {
		return err
	}
	b.ev.Emit("OnRawMessage", strings.Split(m, " "))
	return nil
}

func (b *Client) readIrcMessages(conn net.Conn, done <-chan struct{}) {
	tp := textproto.NewReader(bufio.NewReader(conn))
	for {
		content, err := tp.ReadLine()
		if err != nil {
			if err == io.EOF {
				b.ev.Emit("Error", ErrConnectionClosed)
			} else {
				b.ev.Emit("Error", err)
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
			b.ev.Emit("RawMessage", splits)
			if splits[0] == "PING" {
				b.ev.Emit("PING")
				b.Send("PONG " + strings.Join(append(splits[:0], splits[1:]...), " "))
			}

			// When connection was closed unexpectedly from our side there can be an unfinished message
			// TODO: find a way to fix that issue
			if len(splits) < 1 {
				break
			}

			ircHandler, ok := IrcHandlers[splits[1]]
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

			if b.BotAccount && msg.Type() == "channel" {
				msg.C <- errors.New("bot accounts aren't allowed to send messages in channels")
			}
			err := b.Send(fmt.Sprintf("PRIVMSG %s :%s", name, content))
			if err != nil {
				msg.C <- err
				break
			}

			switch s := msg.MessageSender.(type) {
			case *User:
				b.ev.Emit("PrivateMessage", newPrivateMessage(b, b.GetSelf(), s, true, content))
			case *Channel:
				b.ev.Emit("ChannelMessage", newChannelMessage(b, b.GetSelf(), s, true, content))
			}

			msg.C <- nil
		case <-b.Done:
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
		channel = NewChannel(b, channelName)
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
		b.ev.Emit("Connect", state)
	}

	if state == Disconnected {
		b.ev.Emit("Disconnect", nil)
	}

	b.ev.Emit("StateChanged", state)
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
