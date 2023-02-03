package banchogo

type OutgoingBanchoMessage struct {
	MessageSender

	client *BanchoClient

	Content string
	C       chan error
}

func newOutgoingBanchoMessage(client *BanchoClient, sender MessageSender, message string) *OutgoingBanchoMessage {
	return &OutgoingBanchoMessage{
		sender,
		client,
		message,
		nil,
	}
}

func (o *OutgoingBanchoMessage) Send() <-chan error {
	o.C = make(chan error, 1)
	if o.client.IsConnected() {
		o.client.messageQueue <- o
	} else {
		o.C <- ErrConnectionClosed
	}
	return o.C
}
