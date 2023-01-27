package ircbanchogo

type Message interface {
	// TODO: Not sure if I am even need this interface, but I will keep it for now
	Content() string
	Action() string
}

type BanchoMessage struct {
	User    BanchoUser
	Message string
}

func (b *BanchoMessage) Content() string {
	return b.Message
}

func (b *BanchoMessage) Action() string {
	// TODO:
	return ""
}

type PrivateMessage struct {
	BanchoMessage
	Self bool
}

type ChannelMessage struct {
	BanchoMessage
	Channel BanchoChannel
}
