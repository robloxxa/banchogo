package banchogo

import "regexp"

var actionRegex = regexp.MustCompile("^\x01ACTION( (.+)?)?\x01")

type MessageSender interface {
	Name() string
	SendMessage(string) error
	SendAction(string) error
	Type() string
}

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
	action := actionRegex.FindStringSubmatch(b.Content())
	if action != nil {
		return action[2]
	}
	return ""
}
