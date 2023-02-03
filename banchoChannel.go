package banchogo

import (
	"github.com/puzpuzpuz/xsync/v2"
)

type BanchoChannel struct {
	client *BanchoClient

	ChannelName string
	Topic       string
	Joined      bool
	Members     *xsync.MapOf[string, *BanchoChannelMember]

	joinSignal chan error
}

func NewBanchoChannel(b *BanchoClient, name string) *BanchoChannel {
	return &BanchoChannel{
		client:      b,
		ChannelName: name,
		Topic:       "",
		Joined:      false,
		Members:     xsync.NewMapOf[*BanchoChannelMember](),
	}
}

func (c *BanchoChannel) Name() string {
	return c.ChannelName
}

func (c *BanchoChannel) SendMessage(message string) <-chan error {
	return newOutgoingBanchoMessage(c.client, c, message).Send()
}

func (c *BanchoChannel) SendAction(message string) <-chan error {
	return newOutgoingBanchoMessage(c.client, c, "ACTION "+message).Send()
}
