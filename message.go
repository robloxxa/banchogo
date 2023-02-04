package banchogo

type Message interface {
	Sender() MessageSender
	Content() string
	Action() string
}

type message struct {
	Bancho *Client

	User    *User
	Self    bool
	Message string
}

func (b *message) Content() string {
	return b.Message
}

func (b *message) Action() string {
	// TODO:
	return ""
}
