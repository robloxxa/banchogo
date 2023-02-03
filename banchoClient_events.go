package banchogo

func (b *BanchoClient) HandleConnect(handler func()) func() {
	return b.AddHandler("Connect", handler)
}

func (b *BanchoClient) HandleConnectOnce(handler func()) func() {
	return b.AddHandlerOnce("Connect", handler)
}

func (b *BanchoClient) HandleDisconnect(handler func(error)) func() {
	return b.AddHandler("Disconnect", handler)
}

func (b *BanchoClient) HandleDisconnectOnce(handler func(error)) func() {
	return b.AddHandlerOnce("Disconnect", handler)
}

func (b *BanchoClient) HandleError(handler func(error)) func() {
	return b.AddHandler("Error", handler)
}

func (b *BanchoClient) HandleErrorOnce(handler func(error)) func() {
	return b.AddHandlerOnce("Error", handler)
}

func (b *BanchoClient) HandleRawMessage(handler func([]string)) func() {
	return b.AddHandler("RawMessage", handler)
}

func (b *BanchoClient) HandleRawMessageOnce(handler func([]string)) func() {
	return b.AddHandlerOnce("RawMessage", handler)
}

func (b *BanchoClient) HandlePrivateMessage(handler func(*PrivateMessage)) func() {
	return b.AddHandler("PrivateMessage", handler)
}

func (b *BanchoClient) HandlePrivateMessageOnce(handler func(*PrivateMessage)) func() {
	return b.AddHandlerOnce("PrivateMessage", handler)
}

func (b *BanchoClient) HandleMessage(handler func(BanchoMessage)) func() {
	return b.AddHandler("Message", handler)
}

func (b *BanchoClient) HandleMessageOnce(handler func(BanchoMessage)) func() {
	return b.AddHandlerOnce("Message", handler)
}
