package banchogo

type PrivateMessage struct {
	message

	Recipient *User
}

func newPrivateMessage(b *Client, user *User, recipient *User, self bool, content string) *PrivateMessage {
	return &PrivateMessage{
		message{b, user, self, content},
		recipient,
	}
}

func (p *PrivateMessage) Sender() MessageSender {
	return p.User
}
