package banchogo

type ChannelMessage struct {
	banchoMessage

	Channel *BanchoChannel
}

func newChannelMessage(b *BanchoClient, user *BanchoUser, channel *BanchoChannel, self bool, content string) *ChannelMessage {
	return &ChannelMessage{
		banchoMessage{b, user, self, content},
		channel,
	}
}

func (c *ChannelMessage) Sender() MessageSender {
	return c.Channel
}
