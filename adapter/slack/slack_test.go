package slack

import (
	"testing"
	"reflect"

	"github.com/danryan/hal"
	"github.com/nlopes/slack"
)

type fakeRTM struct {
	RTMService
}

func (rtm *fakeRTM) SendMessage(message *slack.OutgoingMessage) {}

func (rtm *fakeRTM) NewOutgoingMessage(text string, channelID string) *slack.OutgoingMessage {
	return &slack.OutgoingMessage{}
}

type fakeApi struct {
	Api
}

var users = []User{{ID: "123", Name: "user1"}, {ID: "124", Name: "user2"}}

func (api *fakeApi) GetUsers() ([]User, error) {
	return users, nil
}

type fakeStore struct {
	hal.Store
}

func (s *fakeStore) Set(key string, data []byte) error {
	return nil
}


func TestSend(t *testing.T) {
	rtm := &RTM{rtmService: &fakeRTM{}}
	res := &hal.Response{Message: &hal.Message{Room: "general"}}
	a := &adapter{rtm: rtm}
	err := a.Send(res, "abc")
	if err != nil {
		t.Error(err)
	}
}

func TestReply(t *testing.T) {
	rtm := &RTM{rtmService: &fakeRTM{}}
	res := &hal.Response{
		Message: &hal.Message{Room: "general"},
		Envelope: &hal.Envelope{User: &hal.User{Name: "hal"}},
	}
	a := &adapter{rtm: rtm}
	err := a.Reply(res, "abc")
	if err != nil {
		t.Error(err)
	}
}

func TestEmote(t *testing.T) {
	t.Skip()
}

func TestTopic(t *testing.T) {
	t.Skip()
}

func TestPlay(t *testing.T) {
	t.Skip()
}

func TestRun(t *testing.T) {
	t.Skip()
}

func TestStop(t *testing.T) {
	a := &adapter{}
	err := a.Stop()
	if err != nil {
		t.Error(err)
	}
}

func TestInChannels(t *testing.T) {
	a := &adapter{channels: []string{"a", "b", "c"}}
	var roomCases = []struct{
		room string
		want bool
	}{
		{"a", true},
		{"b", true},
		{"d", false},
	}

	for _, c := range roomCases {
		got := a.inChannels(c.room)
		if got != c.want {
			t.Errorf("inChannels(%q) == %t, want %t", c.room, got, c.want)
		}
	}
}

func TestSetAllUsers(t *testing.T) {
	a := &adapter{api: &fakeApi{}}
	robot := &hal.Robot{}
	robot.Users = hal.NewUserMap(robot)
	robot.Store = &fakeStore{}
	a.SetRobot(robot)

	a.setAllUsers()

	var setUsers []User
	for _, u := range robot.Users.All() {
		setUsers = append(setUsers, User{ID: u.ID, Name: u.Name})
	}

	if reflect.DeepEqual(setUsers, users) {
		t.Errorf("setUsers %v, want %v", setUsers, users)
	}
}
