package banchogo

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

// TODO: make ingame status

var (
	banchoStatsForRegex = regexp.MustCompile(
		`Stats for \((.+)\)\[https://osu\.ppy\.sh/u/(d+)\](?: is (.+))?`,
	)
	banchoStatsScoreRegex = regexp.MustCompile(
		`Score: {4}(.+) \(#(\d+)\)`,
	)
	banchoStatsPlaysRegex = regexp.MustCompile(
		`Plays: {4}(\d+) \(lv(\d+)\)`,
	)
	banchoStatsAccuracyRegex = regexp.MustCompile(
		`Accuracy: (\d+(\.\d+)?)%`,
	)
)

type BanchoBotStatsResponse struct {
	UserID      int
	Username    string
	RankedScore int64
	Rank        int
	Level       int
	Accuracy    float64
	Status      string
	Online      bool
	Error       error
}

type banchoBotStatsCommand struct {
	User *User

	response     BanchoBotStatsResponse
	responseChan chan BanchoBotStatsResponse

	handlers        [5]func(m *PrivateMessage)
	handlerRemovers [5]func()
}

func newBanchoBotStatsCommand(user *User) (s *banchoBotStatsCommand) {

	s = &banchoBotStatsCommand{
		User:         user,
		responseChan: make(chan BanchoBotStatsResponse, 1),
	}

	userNotFoundHandler := func(m *PrivateMessage) {
		if m.Message == "User not found" {
			s.responseChan <- BanchoBotStatsResponse{
				Error: ErrUserNotFound,
			}
			s.removeHandlers()
		}
	}

	statusHandler := func(m *PrivateMessage) {
		r := banchoStatsForRegex.FindStringSubmatch(m.Message)
		if r != nil && s.matchIrcUsername(r[1]) {
			s.response.Status = r[4]
			s.response.Online = r[4] != ""

			s.User.mu.Lock()
			s.User.data.Username = r[1]
			s.User.data.UserID, _ = strconv.Atoi(r[2])
			s.User.mu.Unlock()
		}
	}

	scoreHandler := func(m *PrivateMessage) {
		r := banchoStatsScoreRegex.FindStringSubmatch(m.Message)
		if r != nil {
			rankedScore, _ := strconv.Atoi(strings.ReplaceAll(r[1], ",", ""))
			rank, _ := strconv.Atoi(r[2])

			s.response.RankedScore = int64(rankedScore)
			s.response.Rank = rank

			s.User.mu.Lock()
			s.User.data.RankedScore = int64(rankedScore)
			s.User.data.Rank = rank
			s.User.mu.Unlock()
		}
	}

	playsHandler := func(m *PrivateMessage) {
		r := banchoStatsPlaysRegex.FindStringSubmatch(m.Message)
		if r != nil {
			s.User.data.Playcount, _ = strconv.Atoi(r[1])
			s.response.Level, _ = strconv.Atoi(r[2])
		}
	}

	accuracyHandler := func(m *PrivateMessage) {
		r := banchoStatsAccuracyRegex.FindStringSubmatch(m.Message)
		if r != nil {
			s.response.Accuracy, _ = strconv.ParseFloat(r[1], 64)
			s.responseChan <- s.response
			s.removeHandlers()
		}
	}

	s.handlers = [5]func(m *PrivateMessage){userNotFoundHandler, statusHandler, scoreHandler, playsHandler, accuracyHandler}

	return
}

func (s *banchoBotStatsCommand) Send() chan BanchoBotStatsResponse {
	s.registerHandlers()
	s.User.client.GetUser("BanchoBot").SendMessage("!stats " + s.User.Name())

	time.AfterFunc(10*time.Second, func() {
		select {
		case s.responseChan <- BanchoBotStatsResponse{Error: ErrMessageTimeout}:
			s.removeHandlers()
		default:
		}
	})
	return s.responseChan
}

func (s *banchoBotStatsCommand) registerHandlers() {
	for i, h := range s.handlers {
		s.handlerRemovers[i] = s.User.client.GetUser("BanchoBot").OnMessage(h)
	}
}

func (s *banchoBotStatsCommand) removeHandlers() {
	for _, h := range s.handlerRemovers {
		h()
	}
}

func (s *banchoBotStatsCommand) matchIrcUsername(username string) bool {
	return strings.ReplaceAll(username, " ", "_") == strings.ToLower(s.User.Name())
}
