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

var ircHandlers = map[string]func(*Client, []string){
	"001":     handleWelcomeCommand,
	"353":     handleJoinCommand,
	"464":     handleBadAuthCommand,
	"PRIVMSG": handlePrivmsgCommand,
	"JOIN":    handleJoinCommand,
	"QUIT":    handleQuitCommand,
}

func handleWelcomeCommand(b *Client, splits []string) {
	b.setConnectState(Connected)
	b.connectSignal <- nil
}

func handleBadAuthCommand(b *Client, splits []string) {
	b.connectSignal <- ErrBadAuthentication
}

func handlePrivmsgCommand(b *Client, splits []string) {
	username := b.GetUser(splits[0][1:strings.Index(splits[0], "!")])
	content := strings.Join(splits[3:], " ")[1:]

	if strings.ToLower(splits[2]) == strings.ToLower(b.Username) {
		pm := newPrivateMessage(b, username, b.GetSelf(), false, content)
		b.ev.Handle("PrivateMessage", pm)
		b.ev.Handle("Message", Message(pm))
	} else if strings.Index(splits[2], "#") == -1 {
		b.ev.Handle("RejectedMessage", newPrivateMessage(b, username, b.GetSelf(), true, content))
	} else {
		channel, _ := b.GetChannel(splits[2])
		cm := newChannelMessage(b, username, channel, true, content)
		b.ev.Handle("ChannelMessage", cm)
		b.ev.Handle("Message", Message(cm))
	}
}

func handleJoinCommand(b *Client, splits []string) {

}

func handleQuitCommand(b *Client, splits []string) {
	username := splits[0][1:strings.Index(splits[0], "!")]
	user := b.GetUser(username)

	b.ev.Handle("QUIT", user) // TODO: make event type for QUIT

	b.Channels.Range(func(_ string, v *Channel) bool {
		v.Members.Delete(user.IrcUsername)
		return true
	})
}
