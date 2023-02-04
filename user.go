package banchogo

import (
	"regexp"
	"strings"
)

var (
	whereRegex = regexp.MustCompile("/(.+) is in (.+)")
)

type whereResponse struct {
	Country string
	Error   error
}

type User struct {
	ev EventEmitter

	client *Client

	whereChan       chan whereResponse
	runningHandlers *[1]func()

	IrcUsername string

	//TODO: Implement existing or create my own osu api wrapper
	//Id       int
	//Username string
	//JoinDate time.Time
	//Count300 int
	//Count100 int
	//Count50  int
}

func newBanchoUser(client *Client, username string) *User {
	return &User{
		client:      client,
		IrcUsername: username,
	}
}

func (u *User) Name() string {
	return u.IrcUsername
}

func (u *User) SendMessage(message string) <-chan error {
	return newOutgoingBanchoMessage(u.client, u, message).Send()
}

func (u *User) SendAction(message string) <-chan error {
	return newOutgoingBanchoMessage(u.client, u, "ACTION "+message).Send()
}

func (u *User) on(name string, handler interface{}, once bool) func() {
	if u.runningHandlers == nil {
		pmHandler := u.client.HandlePrivateMessage(func(m *PrivateMessage) {
			if m.User != u {
				return
			}
			u.ev.Handle("Message", m)
		})

		u.runningHandlers = &[1]func(){pmHandler}

		// TODO: Figure out the proper way to clear events when object is gced
		// Note: SetFinalizer prevent object to be freed when gc tries to free it for the first time
		//runtime.SetFinalizer(u, func() {
		//	for _, f := range u.runningHandlers {
		//		f()
		//	}
		//})
	}
	if once {
		return u.ev.AddHandlerOnce(name, handler)
	} else {
		return u.ev.AddHandler(name, handler)
	}
}

func (u *User) HandleMessage(handler func(*PrivateMessage)) func() {
	return u.on("Message", handler, false)
}

func (u *User) HandleMessageOnce(handler func(*PrivateMessage)) func() {
	return u.on("Message", handler, true)
}

func (u *User) IsClient() bool {
	return strings.ToLower(u.client.Username) == strings.ToLower(u.IrcUsername)
}

func (u *User) Where() <-chan whereResponse {
	if u.whereChan == nil {
		var clearEvent func()
		u.whereChan = make(chan whereResponse, 1)

		banchoBot := u.client.GetUser("BanchoBot")

		afterAnswer := func() {
			clearEvent()
			u.whereChan = nil
		}

		clearEvent = banchoBot.HandleMessage(func(m *PrivateMessage) {
			if m.Message == "The user is currently not offline" {
				u.whereChan <- whereResponse{"", ErrUserOffline}
				afterAnswer()
				return
			}

			country := whereRegex.FindStringSubmatch(m.Message)
			if country != nil && country[1] == m.User.IrcUsername {
				u.whereChan <- whereResponse{country[2], nil}
				afterAnswer()
				return
			}
		})
		banchoBot.SendMessage("!where " + u.IrcUsername)
	}

	return u.whereChan
}
