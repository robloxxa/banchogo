package banchogo

import (
	"strconv"
	"strings"
	"sync"
)

// Lobby a bancho multiplayer lobby
type Lobby struct {
	mu     sync.Mutex
	Client *Client

	Id      int
	Channel *Channel

	name      string
	beatmapId int
	beatmap   string
	//teamMode
	//winCondition
	//mods
	freemod bool
	playing bool
	//slots [16]
	size int

	//players []

}

func NewLobby(c *Channel) (l *Lobby) {
	l = &Lobby{
		Client:  c.client,
		Channel: c,

		size: 16,
	}
	l.Id, _ = strconv.Atoi(c.Name()[4:len(c.Name())])

	l.Channel.OnJoin(func(m *ChannelMember) {
		if m.User.IsClient() {
			// TODO:
		}
	})

	l.Channel.OnMessage(func(m *ChannelMessage) {
		if strings.ToLower(m.User.ircUsername) == "banchobot" {

		}
	})

	return &Lobby{Channel: c}
}

func (l *Lobby) Name() string {
	return l.Channel.Name()
}

func (l *Lobby) SendMessage(message string) error {
	return newOutgoingBanchoMessage(l.Client, l, message).Send()
}

func (l *Lobby) SendAction(message string) error {
	return newOutgoingBanchoMessage(l.Client, l, message).Send()
}

func (l *Lobby) Type() string {
	return "mp"
}
