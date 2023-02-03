package banchogo

import "strings"

var ignoredCodes = []string{
	"312",
	"333",
	"366",
	"372",
	"375",
	"376",
}

var ircHandlers = map[string]func(*BanchoClient, []string){
	"001":     handleWelcomeCommand,
	"353":     handleJoinCommand,
	"464":     handleBadAuthCommand,
	"PRIVMSG": handlePrivmsgCommand,
	"JOIN":    handleJoinCommand,
	"QUIT":    handleQuitCommand,
}

func handleWelcomeCommand(b *BanchoClient, splits []string) {
	b.setConnectState(Connected)
	b.connectSignal <- nil
}

func handleBadAuthCommand(b *BanchoClient, splits []string) {
	b.connectSignal <- ErrBadAuthentication
}

func handlePrivmsgCommand(b *BanchoClient, splits []string) {
	username := b.GetUser(splits[0][1:strings.Index(splits[0], "!")])
	message := strings.Join(splits[3:], " ")[1:]

	if strings.ToLower(splits[2]) == strings.ToLower(b.Username) {
		pm := newPrivateMessage(b, username, b.GetSelf(), false, message)
		b.handle("PrivateMessage", pm)
		b.handle("Message", BanchoMessage(pm))
	} else if strings.Index(splits[2], "#") == -1 {
		b.handle("RejectedMessage", newPrivateMessage(b, username, b.GetSelf(), true, message))
	} else {
		channel, _ := b.GetChannel(splits[2])
		cm := newChannelMessage(b, username, channel, true, message)
		b.handle("ChannelMessage", cm)
		b.handle("Message", BanchoMessage(cm))
	}
}

func handleJoinCommand(b *BanchoClient, splits []string) {

}

func handleQuitCommand(b *BanchoClient, splits []string) {
	username := splits[0][1:strings.Index(splits[0], "!")]
	user := b.GetUser(username)

	b.handle("QUIT", user) // TODO: make event type for QUIT

	b.Channels.Range(func(k string, v *BanchoChannel) bool {
		v.Members.Delete(user.IrcUsername)
		return true
	})
}
