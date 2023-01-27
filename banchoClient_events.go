package ircbanchogo

func (b *BanchoClient) OnConnect(callback func()) {
	b.Event.on("connect", callback)
}

func (b *BanchoClient) OnStateChanged(callback func(ConnectState)) {
	b.Event.on("stateChanged", callback)
}

func (b *BanchoClient) OnError(callback func(error)) {
	b.Event.on("error", callback)
}

func (b *BanchoClient) OnRawMessage(callback func(string)) {
	b.Event.on("rawMessage", callback)
}

func (b *BanchoClient) OnPrivateMessage(callback func(BanchoUser, string)) {
	b.Event.on("privateMessage", callback)
}

func (b *BanchoClient) OnChannelMessage(callback func(BanchoUser, string)) {
	b.Event.on("channelMessage", callback)
}
