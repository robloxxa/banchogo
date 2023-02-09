package banchogo

import (
	"github.com/puzpuzpuz/xsync/v2"
	"runtime"
	"time"
)

type Channel struct {
	ev     *EventEmitter
	client *Client

	ChannelName string
	Topic       string
	Joined      bool
	Members     *xsync.MapOf[string, *ChannelMember]

	handlerRemovers [3]func()

	joinSignal chan error
	partSignal chan error
}

func NewChannel(b *Client, name string) *Channel {
	return &Channel{
		client:      b,
		ChannelName: name,
		Topic:       "",
		Members:     xsync.NewMapOf[*ChannelMember](),
		joinSignal:  make(chan error),
		partSignal:  make(chan error),
	}
}

func (c *Channel) Name() string {
	return c.ChannelName
}

func (c *Channel) SendMessage(message string) error {
	return newOutgoingBanchoMessage(c.client, c, message).Send()
}

func (c *Channel) SendAction(message string) error {
	return newOutgoingBanchoMessage(c.client, c, "\x01ACTION "+message+"\x01").Send()
}

func (c *Channel) Type() string {
	return "channel"
}

func (c *Channel) on(name string, handler interface{}, once bool) func() {

	if c.ev == nil {

		c.ev = &EventEmitter{}

		c.handlerRemovers = [3]func(){
			c.client.OnChannelMessage(func(m *ChannelMessage) {
				if m.Channel != c {
					return
				}
				c.ev.Emit("Message", m)
			}),

			c.client.OnJoin(func(m *ChannelMember) {
				if m.Channel != c {
					return
				}
				c.ev.Emit("Join")
			}),

			c.client.OnPart(func(m *ChannelMember) {
				if m.Channel != c {
					return
				}
				c.ev.Emit("Part")
			})}

		// TODO: Figure out the proper way to clear events when object is gced
		// Note: SetFinalizer prevent object to be freed when gc tries to free it for the first time
		runtime.SetFinalizer(c, func() {
			for _, f := range c.handlerRemovers {
				f()
			}
		})
	}
	if once {
		return c.ev.Once(name, handler)
	} else {
		return c.ev.On(name, handler)
	}
}

func (c *Channel) Join() <-chan error {
	return c.joinOrPart("JOIN", c.joinSignal)
}

func (c *Channel) Leave() <-chan error {
	return c.joinOrPart("PART", c.partSignal)
}

func (c *Channel) joinOrPart(action string, ch <-chan error) <-chan error {
	resp := make(chan error, 1)
	go func() {
		err := c.client.Send("%s %s", action, c.Name())
		if err != nil {
			resp <- err
			return
		}
		select {
		case err = <-ch:
			resp <- err
		case <-time.After(10 * time.Second):
			resp <- ErrMessageTimeout
		}
	}()
	return resp
}

func (c *Channel) OnMessage(handler func(*ChannelMessage)) func() {
	return c.on("Message", handler, false)
}

func (c *Channel) OnceMessage(handler func(*ChannelMessage)) func() {
	return c.on("Message", handler, true)
}

func (c *Channel) OnJoin(handler func(*ChannelMember)) func() {
	return c.on("Join", handler, false)
}

func (c *Channel) OnceJoin(handler func(*ChannelMember)) func() {
	return c.on("Join", handler, true)
}

func (c *Channel) OnPart(handler func(*ChannelMember)) func() {
	return c.on("Part", handler, false)
}

func (c *Channel) OncePart(handler func(*ChannelMember)) func() {
	return c.on("Part", handler, true)
}
