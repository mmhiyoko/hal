package slack

import (
	"github.com/danryan/hal"
	"github.com/nlopes/slack"
)

type RTMService interface {
	SendMessage(*slack.OutgoingMessage)
	ManageConnection()
	NewOutgoingMessage(string, string) *slack.OutgoingMessage
}

type RTM struct {
	rtmService RTMService
	IncomingEvents chan slack.RTMEvent
}

func (r *RTM) SendMessage(msg *slack.OutgoingMessage) {
	r.rtmService.SendMessage(msg)
}

func (r *RTM) ManageConnection() {
	r.rtmService.ManageConnection()
}

func (r *RTM) NewOutgoingMessage(text string, channelID string) *slack.OutgoingMessage {
	return r.rtmService.NewOutgoingMessage(text, channelID)
}

func (a *adapter) startConnection() {
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
