package banchogo

type BanchoMessage interface {
	Sender() MessageSender
	Content() string
	Action() string
}

type banchoMessage struct {
	Bancho *BanchoClient

	User    *BanchoUser
	Self    bool
	Message string
}

func (b *banchoMessage) Content() string {
	return b.Message
}

func (b *banchoMessage) Action() string {
	// TODO:
	return ""
}
