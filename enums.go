package main

type ConnectState int8

const (
	Disconnected ConnectState = iota
	Connecting
	Reconnecting
	Connected
)

// https://bancho.js.org/lib_BanchoClient.js.html
var IgnoredCodes = []string{
	"312",
	"333",
	"366",
	"372",
	"375",
	"376",
}
