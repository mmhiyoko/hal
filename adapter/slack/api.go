package slack

import (
	"github.com/nlopes/slack"
)

type Api interface {
	GetUsers() ([]User, error)
	NewRTM() *RTM
}

type User struct {
	ID string
	Name string
}

type api struct {
	Api
	client *slack.Client
}

func (a *api) GetUsers() ([]User, error) {
	users, err := a.client.GetUsers()
	if err != nil {
		return []User{}, err
	}

	var results []User
	for _, user := range users {
		results = append(results, User{user.ID, user.Name})
	}
	return results, nil
}

func (a *api) NewRTM() *RTM {
	rtm := a.client.NewRTM()
	return &RTM{rtmService: rtm, IncomingEvents: rtm.IncomingEvents}
}

func NewApi(token string) Api {
	return &api{client: slack.New(token)}
}
