package banchogo

func (b *Client) OnConnect(handler func()) func() {
	return b.ev.On("Connect", handler)
}

func (b *Client) OnceConnect(handler func()) func() {
	return b.ev.Once("Connect", handler)
}

func (b *Client) OnDisconnect(handler func(error)) func() {
	return b.ev.On("Disconnect", handler)
}

func (b *Client) OnceDisconnect(handler func(error)) func() {
	return b.ev.Once("Disconnect", handler)
}

func (b *Client) OnConnectState(handler func(ConnectState)) func() {
	return b.ev.On("ConnectState", handler)
}

func (b *Client) OnceConnectState(handler func(ConnectState)) func() {
	return b.ev.Once("ConnectState", handler)
}

func (b *Client) OnError(handler func(error)) func() {
	return b.ev.On("Error", handler)
}

func (b *Client) OnceError(handler func(error)) func() {
	return b.ev.Once("Error", handler)
}

func (b *Client) OnRawMessage(handler func([]string)) func() {
	return b.ev.On("RawMessage", handler)
}

func (b *Client) OnceRawMessage(handler func([]string)) func() {
	return b.ev.Once("RawMessage", handler)
}

func (b *Client) OnPrivateMessage(handler func(*PrivateMessage)) func() {
	return b.ev.On("PrivateMessage", handler)
}

func (b *Client) OncePrivateMessage(handler func(*PrivateMessage)) func() {
	return b.ev.Once("PrivateMessage", handler)
}

func (b *Client) OnChannelMessage(handler func(*ChannelMessage)) func() {
	return b.ev.On("ChannelMessage", handler)
}

func (b *Client) OnceChannelMessage(handler func(*ChannelMessage)) func() {
	return b.ev.Once("ChannelMessage", handler)
}

func (b *Client) OnMessage(handler func(Message)) func() {
	return b.ev.On("Message", handler)
}

func (b *Client) OnceMessage(handler func(Message)) func() {
	return b.ev.Once("Message", handler)
}

func (b *Client) OnJoin(handler func(*ChannelMember)) func() {
	return b.ev.On("Join", handler)
}

func (b *Client) OnceJoin(handler func(*ChannelMember)) func() {
	return b.ev.Once("Join", handler)
}

func (b *Client) OnPart(handler func(*ChannelMember)) func() {
	return b.ev.On("Part", handler)
}

func (b *Client) OncePart(handler func(*ChannelMember)) func() {
	return b.ev.Once("Part", handler)
}

func (b *Client) OnQuit(handler func(*User)) func() {
	return b.ev.On("Quit", handler)
}

func (b *Client) OnceQuit(handler func(*User)) func() {
	return b.ev.Once("Quit", handler)
}
