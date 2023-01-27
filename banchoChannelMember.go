package ircbanchogo

type BanchoChannelMemberMode string

const (
	IRCUser      BanchoChannelMemberMode = "v"
	IRCModerator                         = "o"
)

type BanchoChannelMember struct {
	Channel *BanchoChannel
	User    *BanchoUser
	Mode    BanchoChannelMemberMode
}
