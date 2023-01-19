package banchogo

func (b *BanchoClient) OnConnect(callback func()) {
	b.Event.on("connect", callback)
}

func (b *BanchoClient) OnError(callback func(error)) {
	b.Event.on("error", callback)
}

func (b *BanchoClient) OnPrivateMessage(callback func(BanchoUser, string)) {
	b.Event.on("privateMessage", callback)
}

func (b *BanchoClient) OnChannelMessage(callback func(BanchoUser, string)) {
	b.Event.on("channelMessage", callback)
}
