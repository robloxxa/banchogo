package banchogo

type OutgoingMessage struct {
	MessageSender

	client *Client

	Content string
	C       chan error
}

func newOutgoingBanchoMessage(client *Client, sender MessageSender, message string) *OutgoingMessage {
	return &OutgoingMessage{
		sender,
		client,
		message,
		nil,
	}
}

func (o *OutgoingMessage) Send() <-chan error {
	o.C = make(chan error, 1)
	if o.client.IsConnected() {
		o.client.messageQueue <- o
	} else {
		o.C <- ErrConnectionClosed
	}
	return o.C
}
