package slack

import (
	"github.com/danryan/hal"
	"github.com/nlopes/slack"
)

func (a *adapter) startConnection() {
	api := slack.New(a.token)

	users, err := api.GetUsers()
	if err != nil {
		hal.Logger.Debugf("%s\n", err)
	}

	for _, user := range users {
		newUser := hal.User{
			ID:   user.ID,
			Name: user.Name,
		}
		u, err := a.Robot.Users.Get(user.ID)
		if err != nil {
			a.Robot.Users.Set(user.ID, newUser)
		}

		if u.Name != user.Name {
			a.Robot.Users.Set(user.ID, newUser)
		}
	}
	hal.Logger.Debugf("Stored users: %s\n", a.Robot.Users.All())

	a.rtm = api.NewRTM()
	go a.rtm.ManageConnection()

	for msg := range a.rtm.IncomingEvents {
		hal.Logger.Debug("Event Received: ")
		switch msg.Data.(type) {
		case slack.HelloEvent:
		case *slack.MessageEvent:
			m := msg.Data.(*slack.MessageEvent)
			hal.Logger.Debugf("Message: %v\n", m)
			msg := a.newMessage(m)
			a.Receive(msg)
		case *slack.PresenceChangeEvent:
			m := msg.Data.(*slack.PresenceChangeEvent)
			hal.Logger.Debugf("Presence Change: %v\n", m)
		case slack.LatencyReport:
			m := msg.Data.(slack.LatencyReport)
			hal.Logger.Debugf("Current latency: %v\n", m.Value)
		case slack.TeamJoinEvent:
			m := msg.Data.(slack.TeamJoinEvent)
			hal.Logger.Debugf("New member joined the team: %v\n", m.User)
			if _, err := a.Robot.Users.Get(m.User.ID); err != nil {
				a.Robot.Users.Set(m.User.ID, hal.User{ID: m.User.ID, Name: m.User.Name})
			}

		default:
			hal.Logger.Debugf("Unexpected: %v\n", msg.Data)
		}
	}
}

func (a *adapter) newMessage(msg *slack.MessageEvent) *hal.Message {
	user, _ := a.Robot.Users.Get(msg.Msg.User)
	return &hal.Message{
		User: user,
		Room: msg.Msg.Channel,
		Text: msg.Text,
	}
}
