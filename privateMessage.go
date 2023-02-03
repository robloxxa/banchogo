package banchogo

type PrivateMessage struct {
	banchoMessage

	Recipient *BanchoUser
}

func newPrivateMessage(b *BanchoClient, user *BanchoUser, recipient *BanchoUser, self bool, content string) *PrivateMessage {
	return &PrivateMessage{
		banchoMessage{b, user, self, content},
		recipient,
	}
}

func (p *PrivateMessage) Sender() MessageSender {
	return p.User
}
