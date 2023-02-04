package banchogo

type ChannelMessage struct {
	message

	Channel *Channel
}

func newChannelMessage(b *Client, user *User, channel *Channel, self bool, content string) *ChannelMessage {
	return &ChannelMessage{
		message{b, user, self, content},
		channel,
	}
}

func (c *ChannelMessage) Sender() MessageSender {
	return c.Channel
}
