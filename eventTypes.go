package banchogo

// Only add here func types that will be used for EventEmitter
// After adding a new type, run `go generate` to generate Call and NumField function for EventHandle interface
//go:generate go run tools/cmd/eventhandler/main.go

type EmptyHandlerType func()

type EllipseInterfaceHandlerType func(...interface{})

type WithErrorHandlerType func(error)

type RawMessageHandlerType func([]string)

type ConnectStateHandlerType func(ConnectState)

type MessageHandlerType func(Message)

type PrivateMessageHandlerType func(*PrivateMessage)

type ChannelMessageHandlerType func(*ChannelMessage)

type ChannelMemberHandlerType func(*ChannelMember)
