package banchogo

import (
	"strings"
)

var ignoredCodes = []string{
	"312",
	"333",
	"366",
	"372",
	"375",
	"376",
}

var IrcHandlers = map[string]func(*Client, []string){
	"001":     handleWelcomeCommand,
	"311":     handleWhoisUserCommand,
	"319":     handleWhoisChannelsCommand,
	"318":     handleWhoisEndCommand,
	"332":     handleChannelTopicCommand,
	"353":     handleNamesCommand,
	"403":     handleChannelNotFoundCommand,
	"464":     handleBadAuthCommand,
	"PRIVMSG": handlePrivmsgCommand,
	"MODE":    handleModeCommand,
	"JOIN":    handleJoinCommand,
	"PART":    handlePartCommand,
	"QUIT":    handleQuitCommand,
}

func handleWelcomeCommand(b *Client, _ []string) {
	b.connectSignal <- nil
	b.setConnectState(Connected)
}

func handleBadAuthCommand(b *Client, _ []string) {
	b.connectSignal <- ErrBadAuthentication
}

func handlePrivmsgCommand(b *Client, splits []string) {
	username := b.GetUser(splits[0][1:strings.Index(splits[0], "!")])
	content := strings.Join(splits[3:], " ")[1:]

	if strings.ToLower(splits[2]) == strings.ToLower(b.Username) {
		pm := newPrivateMessage(b, username, b.GetSelf(), false, content)
		b.ev.Emit("PrivateMessage", pm)
		b.ev.Emit("Message", Message(pm))
	} else if strings.Index(splits[2], "#") == -1 {
		b.ev.Emit("RejectedMessage", newPrivateMessage(b, username, b.GetSelf(), true, content))
	} else {
		channel, _ := b.GetChannel(splits[2])
		cm := newChannelMessage(b, username, channel, true, content)
		b.ev.Emit("ChannelMessage", cm)
		b.ev.Emit("Message", Message(cm))
	}
}

func handleJoinCommand(b *Client, splits []string) {
	channel, _ := b.GetChannel(splits[2][1:])
	user := b.GetUser(splits[0][1:strings.Index(splits[0], "!")])

	member := newChannelMember(b, channel, user.Name())
	channel.Members.Store(user.Name(), member)
	b.ev.Emit("Join", member)

	if user.IsClient() {
		channel.Joined = true
		select {
		case channel.joinSignal <- nil:
		default:
		}
	}
}

func handlePartCommand(b *Client, splits []string) {
	channel, _ := b.GetChannel(splits[2][1:])
	user := b.GetUser(splits[0][1:strings.Index(splits[0], "!")])
	emitPart(b, user, channel)
}

func handleQuitCommand(b *Client, splits []string) {
	username := splits[0][1:strings.Index(splits[0], "!")]
	user := b.GetUser(username)

	b.ev.Emit("Quit", user)

	b.Channels.Range(func(_ string, v *Channel) bool {
		v.Members.Delete(user.Name())
		return true
	})
}

func handleModeCommand(b *Client, splits []string) {
	channel, _ := b.GetChannel(splits[2])
	mode := ChannelMemberMode(splits[3][1:2])
	user := b.GetUser(splits[4])

	channel.Members.Compute(user.Name(), func(oldV *ChannelMember, loaded bool) (newV *ChannelMember, delete bool) {
		if !loaded {
			newV = newChannelMember(b, channel, splits[4])
		} else {
			oldV.Mode = mode
			newV = oldV
		}
		return
	})
}

func handleChannelTopicCommand(b *Client, splits []string) {
	topicSplits := splits[4:]
	topicSplits[0] = topicSplits[0][1:]

	b.Channels.Compute(splits[3], func(c *Channel, loaded bool) (nc *Channel, delete bool) {
		if !loaded {
			nc = NewChannel(b, splits[3])
		} else {
			nc = c
		}
		nc.Topic = strings.Join(topicSplits, " ")
		return
	})
}

func handleNamesCommand(b *Client, splits []string) {
	var names []string
	channel, _ := b.GetChannel(splits[4])

	names = splits[4:]
	names[0] = names[0][1:]

	for _, n := range names {
		member := newChannelMember(b, channel, n)
		channel.Members.Store(member.User.Name(), member)
	}
}

func handleChannelNotFoundCommand(b *Client, splits []string) {
	channel, _ := b.GetChannel(splits[3])

	b.ev.Emit("ChannelNotFound", channel)
	for {
		select {
		case channel.joinSignal <- ErrChannelNotFound:
		case channel.partSignal <- ErrChannelNotFound:
		default:
			return
		}
	}
}

func emitPart(b *Client, u *User, c *Channel) {
	if member, ok := c.Members.Load(u.Name()); ok {
		b.ev.Emit("PART", member)
	}

	if u.IsClient() {
		c.Joined = false
		select {
		case c.partSignal <- nil:
		default:
		}
	}

	b.Channels.Delete(u.Name())
}
