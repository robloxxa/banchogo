package banchogo

type ChannelMemberMode string

const (
	IRCUser      ChannelMemberMode = "v"
	IRCModerator                   = "o"
)

type ChannelMember struct {
	Channel *Channel
	User    *User
	Mode    ChannelMemberMode
}
