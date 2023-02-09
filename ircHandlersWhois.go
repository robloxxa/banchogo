package banchogo

import (
	"regexp"
	"strconv"
)

var (
	osuLinkRegex = regexp.MustCompile(`^https?://osu\.ppy\.sh/u/(\d+)$`)
)

func handleWhoisUserCommand(b *Client, splits []string) {
	user := b.GetUser(splits[3])
	userId, _ := strconv.Atoi(osuLinkRegex.FindStringSubmatch(splits[4])[1])

	select {
	case whoisObj := <-user.whoisChan:
		whoisObj.UserId = userId
		b.currentWhois = whoisObj
	default:
	}
}

func handleWhoisChannelsCommand(b *Client, splits []string) {
	user := b.GetUser(splits[3])
	channelSplits := splits[4 : len(splits)-1]
	channelSplits[0] = channelSplits[0][1:len(channelSplits[0])]

	channels := make([]*Channel, len(channelSplits), cap(channelSplits))

	for i, name := range channelSplits {
		channel, err := b.GetChannel(name)
		if err == nil {
			channels[i] = channel
		}
	}

	if b.currentWhois != nil {
		if b.currentWhois.user != user {
			// TODO: For tests only, will delete later
			panic("Whois implementation is wrong, dumbass")
		}
		b.currentWhois.Channels = append(b.currentWhois.Channels, channels...)
	}
}

func handleWhoisEndCommand(b *Client, splits []string) {
	user := b.GetUser(splits[3])

	if b.currentWhois != nil {
		if b.currentWhois.user != user {
			// TODO: For tests only, will delete later
			panic("Whois implementation is wrong, dumbass")
		}
		close(b.currentWhois.c)
		b.currentWhois = nil
	}
}
