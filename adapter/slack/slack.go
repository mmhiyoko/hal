package slack

import (
	"fmt"
	"os"
	"strings"

	"github.com/danryan/env"
	"github.com/danryan/hal"
)

func init() {
	hal.RegisterAdapter("slack", New)
}

type adapter struct {
	hal.BasicAdapter
	token          string
	team           string
	mode           string
	channels       []string
	channelMode    string
	botname        string
	responseMethod string
	iconEmoji      string
	linkNames      int
	api			   Api
	rtm            *RTM
}

type config struct {
	Token          string `env:"key=HAL_SLACK_TOKEN required"`
	Team           string `env:"key=HAL_SLACK_TEAM required"`
	Channels       string `env:"key=HAL_SLACK_CHANNELS"`
	Mode           string `env:"key=HAL_SLACK_MODE"`
	Botname        string `env:"key=HAL_SLACK_BOTNAME default=hal"`
	IconEmoji      string `env:"key=HAL_SLACK_ICON_EMOJI"`
	ResponseMethod string `env:"key=HAL_SLACK_RESPONSE_METHOD default=http"`
	ChannelMode    string `env:"key=HAL_SLACK_CHANNEL_MODE "`
}

// New returns an initialized adapter
func New(r *hal.Robot) (hal.Adapter, error) {
	c := &config{}
	env.MustProcess(c)
	channels := strings.Split(c.Channels, ",")
	a := &adapter{
		token:          c.Token,
		team:           c.Team,
		channels:       channels,
		channelMode:    c.ChannelMode,
		mode:           c.Mode,
		botname:        c.Botname,
		iconEmoji:      c.IconEmoji,
		responseMethod: c.ResponseMethod,
	}
	hal.Logger.Debugf("%v", os.Getenv("HAL_SLACK_CHANNEL_MODE"))
	hal.Logger.Debugf("channel mode: %v", a.channelMode)
	if a.channelMode == "" {
		a.channelMode = "whitelist"
	}
	a.api = NewApi(a.token)
	a.rtm = a.api.NewRTM()
	a.SetRobot(r)
	return a, nil
}

// Send sends a regular response
func (a *adapter) Send(res *hal.Response, strings ...string) error {
	for _, str := range strings {
		a.rtm.SendMessage(a.rtm.NewOutgoingMessage(str, res.Message.Room))
	}

	return nil
}

// Reply sends a direct response
func (a *adapter) Reply(res *hal.Response, strings ...string) error {
	newStrings := make([]string, len(strings))
	for _, str := range strings {
		newStrings = append(newStrings, fmt.Sprintf("%s: %s", res.UserName(), str))
	}

	a.Send(res, newStrings...)

	return nil
}

// Emote is not implemented.
func (a *adapter) Emote(res *hal.Response, strings ...string) error {
	return nil
}

// Topic sets the topic
func (a *adapter) Topic(res *hal.Response, strings ...string) error {
	for _ = range strings {
	}
	return nil
}

// Play is not implemented.
func (a *adapter) Play(res *hal.Response, strings ...string) error {
	return nil
}

// Receive forwards a message to the robot
func (a *adapter) Receive(msg *hal.Message) error {
	hal.Logger.Debug("slack - adapter received message")

	if len(a.channels) > 0 {
		if a.channelMode == "blacklist" {
			if !a.inChannels(msg.Room) {
				hal.Logger.Debugf("slack - %s not in blacklist", msg.Room)
				hal.Logger.Debug("slack - adapter sent message to robot")
				return a.Robot.Receive(msg)
			}
			hal.Logger.Debug("slack - message ignored due to blacklist")
			return nil
		}

		if a.inChannels(msg.Room) {
			hal.Logger.Debugf("slack - %s in whitelist", msg.Room)
			hal.Logger.Debug("slack - adapter sent message to robot")
			return a.Robot.Receive(msg)
		}
		hal.Logger.Debug("slack - message ignored due to whitelist")
		return nil
	}

	hal.Logger.Debug("slack - adapter sent message to robot")
	return a.Robot.Receive(msg)
}

// Run starts the adapter
func (a *adapter) Run() error {
	a.setAllUsers()
	hal.Logger.Debug("slack - starting RTM connection")
	go a.startConnection()
	hal.Logger.Debug("slack - started RTM connection")

	hal.Logger.Debugf("slack - channelmode=%v channels=%v", a.channelMode, a.channels)
	return nil
}

// Stop shuts down the adapter
func (a *adapter) Stop() error {
	return nil
}

func (a *adapter) inChannels(room string) bool {
	for _, r := range a.channels {
		if r == room {
			return true
		}
	}

	return false
}

func (a *adapter) setAllUsers() {
	users, err := a.api.GetUsers()
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
}

