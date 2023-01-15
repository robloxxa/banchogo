package main

type ConnectState uint8

type BanchoChannelMemberMode string

const (
	BANCHOHOST = "irc.ppy.sh"
	BANCHOPORT = "6667"
)

const (
	Disconnected ConnectState = iota
	Reconnecting
	Connecting
	Connected
)

const (
	IRCUser      BanchoChannelMemberMode = "v"
	IRCModerator                         = "o"
)

var IgnoredCodes = []string{
	"312",
	"333",
	"366",
	"372",
	"375",
	"376",
}
