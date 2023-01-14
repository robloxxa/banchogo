package main

type ConnectState uint8

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

var IgnoredCodes = []string{
	"312",
	"333",
	"366",
	"372",
	"375",
	"376",
}
