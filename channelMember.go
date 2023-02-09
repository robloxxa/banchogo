package banchogo

import "strings"

type ChannelMemberMode string

const (
	IRCUser      ChannelMemberMode = "v"
	IRCModerator                   = "o"
)

func (c ChannelMemberMode) ToSymbol() string {
	switch c {
	case "v":
		return "+"
	case "o":
		return "@"
	default:
		return ""
	}
}

type ChannelMember struct {
	Channel *Channel
	User    *User
	Mode    ChannelMemberMode
}

func newChannelMember(b *Client, channel *Channel, username string) (c *ChannelMember) {
	c = &ChannelMember{
		Channel: channel,
	}
	if strings.Index(username, "@") == 0 {
		c.Mode = IRCModerator
		username = username[1:]
	} else if strings.Index(username, "+") == 0 {
		c.Mode = IRCUser
	} else {
		c.Mode = ""
	}
	c.User = b.GetUser(username)
	return

}
