package banchogo

func (b *Client) HandleConnect(handler func()) func() {
	return b.ev.AddHandler("Connect", handler)
}

func (b *Client) HandleConnectOnce(handler func()) func() {
	return b.ev.AddHandlerOnce("Connect", handler)
}

func (b *Client) HandleDisconnect(handler func(error)) func() {
	return b.ev.AddHandler("Disconnect", handler)
}

func (b *Client) HandleDisconnectOnce(handler func(error)) func() {
	return b.ev.AddHandlerOnce("Disconnect", handler)
}

func (b *Client) HandleError(handler func(error)) func() {
	return b.ev.AddHandler("Error", handler)
}

func (b *Client) HandleErrorOnce(handler func(error)) func() {
	return b.ev.AddHandlerOnce("Error", handler)
}

func (b *Client) HandleRawMessage(handler func([]string)) func() {
	return b.ev.AddHandler("RawMessage", handler)
}

func (b *Client) HandleRawMessageOnce(handler func([]string)) func() {
	return b.ev.AddHandlerOnce("RawMessage", handler)
}

func (b *Client) HandlePrivateMessage(handler func(*PrivateMessage)) func() {
	return b.ev.AddHandler("PrivateMessage", handler)
}

func (b *Client) HandlePrivateMessageOnce(handler func(*PrivateMessage)) func() {
	return b.ev.AddHandlerOnce("PrivateMessage", handler)
}

func (b *Client) HandleMessage(handler func(Message)) func() {
	return b.ev.AddHandler("Message", handler)
}

func (b *Client) HandleMessageOnce(handler func(Message)) func() {
	return b.ev.AddHandlerOnce("Message", handler)
}
