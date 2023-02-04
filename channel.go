package banchogo

import (
	"github.com/puzpuzpuz/xsync/v2"
)

type Channel struct {
	ev     *EventEmitter
	client *Client

	ChannelName string
	Topic       string
	Joined      bool
	Members     *xsync.MapOf[string, *ChannelMember]

	joinSignal chan error
}

func NewBanchoChannel(b *Client, name string) *Channel {
	return &Channel{
		ev:          NewEmitter(),
		client:      b,
		ChannelName: name,
		Topic:       "",
		Joined:      false,
		Members:     xsync.NewMapOf[*ChannelMember](),
	}
}

func (c *Channel) Name() string {
	return c.ChannelName
}

func (c *Channel) SendMessage(message string) <-chan error {
	return newOutgoingBanchoMessage(c.client, c, message).Send()
}

func (c *Channel) SendAction(message string) <-chan error {
	return newOutgoingBanchoMessage(c.client, c, "ACTION "+message).Send()
}
