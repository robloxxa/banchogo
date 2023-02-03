package banchogo

// Only add here func types that will be used for EventEmitter
// After adding a new type, run `go generate` to generate Call and NumField function for EventHandle interface
//go:generate go run tools/cmd/eventhandler/main.go

type EmptyHandlerType func()

type WithErrorHandlerType func(error)

type RawMessageHandlerType func([]string)

type MessageHandlerType func(BanchoMessage)

type PrivateMessageHandlerType func(*PrivateMessage)

type ChannelMessageHandlerType func(*ChannelMessage)
