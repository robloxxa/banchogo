package banchogo

import (
	"github.com/thehowl/go-osuapi"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"
)

var (
	whereRegex = regexp.MustCompile("(.+) is in (.+)")
)

type WhoisResponse struct {
	user *User
	c    chan struct{}

	UserId   int
	Channels []*Channel
	Error    error
}

type WhereResponse struct {
	Country string
	Error   error
}

type User struct {
	mu sync.Mutex

	ev     *EventEmitter
	client *Client

	whoisChan       chan *WhoisResponse
	handlerRemovers [1]func()

	ircUsername string
	data        *osuapi.User
}

func newBanchoUser(client *Client, username string) *User {
	return &User{
		client:      client,
		ircUsername: username,

		whoisChan: make(chan *WhoisResponse, 1),
	}
}

func (u *User) Name() string {
	u.mu.Lock()
	defer u.mu.Unlock()
	return u.ircUsername
}

func (u *User) SendMessage(message string) error {
	return newOutgoingBanchoMessage(u.client, u, message).Send()
}

func (u *User) SendAction(message string) error {
	return newOutgoingBanchoMessage(u.client, u, "ACTION "+message).Send()
}

func (u *User) Type() string {
	return "user"
}

func (u *User) FetchFromAPI(osuMode ...int) (osuapi.User, error) {
	mode := 0
	if len(osuMode) > 0 {
		mode = osuMode[0]
	}
	data, err := u.client.Api.GetUser(osuapi.GetUserOpts{Username: u.ircUsername, Mode: osuapi.Mode(mode)})
	if err != nil {
		return osuapi.User{}, err
	}
	u.mu.Lock()
	u.data = data
	u.mu.Unlock()
	return *data, nil
}

// Data returns copy of API user data.
func (u *User) Data() osuapi.User {
	u.mu.Lock()
	defer u.mu.Unlock()
	if u.data == nil {
		return osuapi.User{}
	}
	return *u.data
}

func (u *User) on(name string, handler interface{}, once bool) func() {
	if u.ev == nil {
		u.ev = &EventEmitter{}

		u.handlerRemovers = [1]func(){
			u.client.OnPrivateMessage(func(m *PrivateMessage) {
				if m.User != u {
					return
				}
				u.ev.Emit("Message", m)
			})}

		// TODO: Figure out the proper way to clear events when object is gced
		// Note: SetFinalizer prevent object to be freed when gc tries to free it for the first time
		runtime.SetFinalizer(u, func() {
			for _, f := range u.handlerRemovers {
				f()
			}
		})
	}
	if once {
		return u.ev.Once(name, handler)
	} else {
		return u.ev.On(name, handler)
	}
}

func (u *User) OnMessage(handler func(*PrivateMessage)) func() {
	return u.on("Message", handler, false)
}

func (u *User) OnceMessage(handler func(*PrivateMessage)) func() {
	return u.on("Message", handler, true)
}

func (u *User) IsClient() bool {
	return strings.ToLower(u.client.Username) == strings.ToLower(u.ircUsername)
}

func (u *User) Where() <-chan WhereResponse {
	var clearEvent func()

	resp := make(chan WhereResponse, 1)

	timer := time.AfterFunc(10*time.Second, func() {
		select {
		case resp <- WhereResponse{"", ErrMessageTimeout}:
		default:
		}
	})

	afterResponse := func() {
		timer.Stop()
		clearEvent()
	}

	banchoBot := u.client.GetUser("BanchoBot")
	clearEvent = banchoBot.OnMessage(func(m *PrivateMessage) {
		if m.Message == "The user is currently not offline" {
			resp <- WhereResponse{"", ErrUserOffline}
			afterResponse()
			return
		}

		country := whereRegex.FindStringSubmatch(m.Message)
		if country != nil && country[1] == u.ircUsername {
			resp <- WhereResponse{country[2], nil}
			afterResponse()
			return
		}
	})

	err := banchoBot.SendMessage("!where " + u.ircUsername)
	if err != nil {
		resp <- WhereResponse{Error: err}
		afterResponse()
	}

	return resp
}

func (u *User) Whois() <-chan WhoisResponse {
	resp := make(chan WhoisResponse, 1)
	whoisObj := &WhoisResponse{user: u, c: make(chan struct{})}
	go func(response chan WhoisResponse, whoisObj *WhoisResponse) {
		for {
			select {
			case u.whoisChan <- whoisObj:
				u.whoisChan <- whoisObj
				err := u.client.Send("WHOIS " + u.ircUsername)
				if err != nil {
					resp <- WhoisResponse{Error: err}
					return
				}
			case <-whoisObj.c:
				resp <- *whoisObj
			case <-time.After(10 * time.Second):
				resp <- WhoisResponse{Error: ErrMessageTimeout}
				return
			}
		}
	}(resp, whoisObj)

	return resp
}

func (u *User) Stats() <-chan BanchoBotStatsResponse {
	return newBanchoBotStatsCommand(u).Send()
}
