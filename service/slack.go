package service

import (
	"fmt"
	"log"
	"strconv"
	"time"

	slack "github.com/nlopes/slack"
)

type SlackService struct {
	Client           map[string]*slack.Client
	RTM              map[string]*slack.RTM
	JoinedChannels   map[string]Channel
	UnjoinedChannels map[string]Channel
	UserCache        map[string]string
	CurrentUserID    map[string]string
}

type Channel struct {
	Id           string
	Name         string
	Topic        string
	SlackChannel interface{}
	ClientId     string
	ChannelType  ChannelType
}

type Channels []Channel

func (s Channels) Len() int {
	return len(s)
}

func (s Channels) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s Channels) Less(i, j int) bool {
	var first string = fmt.Sprintf("%s %d %s", s[i].ClientId, s[i].ChannelType, s[i].Name)
	var second string = fmt.Sprintf("%s %d %s", s[j].ClientId, s[j].ChannelType, s[j].Name)
	return first < second
}

type ChannelType uint8

const (
	CHANNEL ChannelType = iota + 1
	GROUP
	IM
)

// CreateSlackService is the constructor for the SlackService and will initialize
// the RTM and a ClientId
func CreateSlackService(tokens map[string]string) *SlackService {
	svc := &SlackService{
		Client:           make(map[string]*slack.Client),
		RTM:              make(map[string]*slack.RTM),
		JoinedChannels:   make(map[string]Channel),
		UnjoinedChannels: make(map[string]Channel),
		UserCache:        make(map[string]string),
		CurrentUserID:    make(map[string]string),
	}

	for clientId, token := range tokens {
		svc.Client[clientId] = slack.New(token)

		// Get channelUser associated with token, mainly
		// used to identify channelUser when new messages
		// arrives
		authTest, err := svc.Client[clientId].AuthTest()
		if err != nil {
			log.Fatal("ERROR: not able to authorize client, check your connection and/or slack-token")
		}
		svc.CurrentUserID[clientId] = authTest.UserID

		// Create RTM
		svc.RTM[clientId] = svc.Client[clientId].NewRTM()
		go svc.RTM[clientId].ManageConnection()

		// Creation of channelUser cache this speeds up
		// the uncovering of usernames of messages
		users, _ := svc.Client[clientId].GetUsers()
		for _, channelUser := range users {
			// only add non-deleted users
			if !channelUser.Deleted {
				svc.UserCache[channelUser.ID] = channelUser.Name
			}
		}
	}

	return svc
}

// UpdateChannels will retrieve all available channels, groups, and im channels.
// We will return different channel collections, first channels the user is a member of
// and secondly a list of unarchived channels the user can join
// Because the channels are of different types, we will append them to
// an []interface as well as to a []Channel which will give us easy access
// to the id and name of the Channel.
func (s *SlackService) UpdateChannels() {
	// FIXME Check errors
	for currentClientId := range s.Client {
		// Channel
		_ = s.FetchChannels(currentClientId)

		// Groups
		_ = s.FetchGroups(currentClientId)

		// IM
		_ = s.FetchIM(currentClientId)
	}
}

func (s *SlackService) GetChannelList() []Channel {
	s.UpdateChannels()
	var result Channels
	for _, channel := range s.JoinedChannels {
		result = append(result, channel)
	}

	return result
}
func (s *SlackService) FetchIM(currentClientId string) error {
	slackIM, err := s.Client[currentClientId].GetIMChannels()
	if err != nil {
		//chans = append(chans, Channel{})
	}
	for _, im := range slackIM {

		// Uncover name, when we can't uncover name for
		// IM channel this is then probably a deleted
		// user, because we wont add deleted users
		// to the UserCache, so we skip it
		name, ok := s.UserCache[im.User]
		if ok {
			s.JoinedChannels[im.ID] = Channel{im.ID, name, "", im, currentClientId, IM}
		}
	}
	return err
}

func (s *SlackService) FetchGroups(currentClientId string) error {
	slackGroups, err := s.Client[currentClientId].GetGroups(true)
	if err != nil {
		//chans = append(chans, Channel{})
	}
	for _, grp := range slackGroups {
		s.JoinedChannels[grp.ID] = Channel{grp.ID, grp.Name, grp.Topic.Value, grp, currentClientId, GROUP}
	}
	return err
}
func (s *SlackService) FetchChannels(currentClientId string) error {
	slackChans, err := s.Client[currentClientId].GetChannels(true)
	if err != nil {
		//chans = append(chans, Channel{})
	}
	for _, chn := range slackChans {
		if chn.IsMember {
			s.JoinedChannels[chn.ID] = Channel{chn.ID, chn.Name, chn.Topic.Value, chn, currentClientId, CHANNEL}
		} else {
			s.UnjoinedChannels[chn.ID] = Channel{chn.ID, chn.Name, chn.Topic.Value, chn, currentClientId, CHANNEL}
		}
	}
	return err
}

// SetChannelReadMark will set the read mark for a channel, group, and im
// channel based on the current time.
func (s *SlackService) SetChannelReadMark(channelId string) {
	selectedChannel := s.JoinedChannels[channelId]
	switch channel := selectedChannel.SlackChannel.(type) {
	case slack.Channel:
		s.Client[selectedChannel.ClientId].SetChannelReadMark(
			channel.ID, fmt.Sprintf("%f",
				float64(time.Now().Unix())),
		)
	case slack.Group:
		s.Client[selectedChannel.ClientId].SetGroupReadMark(
			channel.ID, fmt.Sprintf("%f",
				float64(time.Now().Unix())),
		)
	case slack.IM:
		s.Client[selectedChannel.ClientId].MarkIMChannel(
			channel.ID, fmt.Sprintf("%f",
				float64(time.Now().Unix())),
		)
	}
}

// SendMessage will send a message to a particular channel
func (s *SlackService) SendMessage(channelId string, message string) {
	currentChannel := s.JoinedChannels[channelId]
	// https://godoc.org/github.com/nlopes/slack#PostMessageParameters
	postParams := slack.PostMessageParameters{
		AsUser: true,
	}

	// https://godoc.org/github.com/nlopes/slack#Client.PostMessage
	s.Client[currentChannel.ClientId].PostMessage(channelId, message, postParams)
}

// GetMessages will get messages for a channel, group or im channel delimited
// by a count.
func (s *SlackService) GetMessages(channelId string, count int) []string {
	channel := s.JoinedChannels[channelId]
	// https://api.slack.com/methods/channels.history
	historyParams := slack.HistoryParameters{
		Count:     count,
		Inclusive: false,
		Unreads:   false,
	}

	// https://godoc.org/github.com/nlopes/slack#History
	history := new(slack.History)
	var err error
	switch chnType := channel.SlackChannel.(type) {
	case slack.Channel:
		history, err = s.Client[channel.ClientId].GetChannelHistory(chnType.ID, historyParams)
		if err != nil {
			log.Fatal(err) // FIXME
		}
	case slack.Group:
		history, err = s.Client[channel.ClientId].GetGroupHistory(chnType.ID, historyParams)
		if err != nil {
			log.Fatal(err) // FIXME
		}
	case slack.IM:
		history, err = s.Client[channel.ClientId].GetIMHistory(chnType.ID, historyParams)
		if err != nil {
			log.Fatal(err) // FIXME
		}
	}

	// Construct the messages
	var messages []string
	for _, message := range history.Messages {
		msg := s.CreateMessage(message, channel.ClientId)
		messages = append(messages, msg...)
	}

	// Reverse the order of the messages, we want the newest in
	// the last place
	var messagesReversed []string
	for i := len(messages) - 1; i >= 0; i-- {
		messagesReversed = append(messagesReversed, messages[i])
	}

	return messagesReversed
}

// CreateMessage will create a string formatted message that can be rendered
// in the Chat pane.
//
// [23:59] <erroneousboat> Hello world!
//
// This returns an array of string because we will try to uncover attachments
// associated with messages.
func (s *SlackService) CreateMessage(message slack.Message, clientId string) []string {
	var msgs []string
	var name string

	// Get username from cache
	name, ok := s.UserCache[message.User]

	// Name not in cache
	if !ok {
		if message.BotID != "" {
			// Name not found, perhaps a bot, use Username
			name, ok = s.UserCache[message.BotID]
			if !ok {
				// Not found in cache, add it
				name = message.Username
				s.UserCache[message.BotID] = message.Username
			}
		} else {
			// Not a bot, not in cache, get user info
			channelUser, err := s.Client[clientId].GetUserInfo(message.User)
			if err != nil {
				name = "unknown"
				s.UserCache[message.User] = name
			} else {
				name = channelUser.Name
				s.UserCache[message.User] = channelUser.Name
			}
		}
	}

	if name == "" {
		name = "unknown"
	}

	// When there are attachments append them
	if len(message.Attachments) > 0 {
		msgs = append(msgs, createMessageFromAttachments(message.Attachments)...)
	}

	// Parse time
	floatTime, err := strconv.ParseFloat(message.Timestamp, 64)
	if err != nil {
		floatTime = 0.0
	}
	intTime := int64(floatTime)

	// Format message
	msg := s.FormatMessage(intTime, name, message.Text)

	msgs = append(msgs, msg)

	return msgs
}
func (s *SlackService) FormatMessage(intTime int64, name string, message string) string {
	msg := fmt.Sprintf(
		"[%s] <[%s](fg-green)> %s",
		time.Unix(intTime, 0).Format("15:04"),
		name,
		message,
	)
	return msg
}

func (s *SlackService) CreateMessageFromMessageEvent(message *slack.MessageEvent, clientId string) []string {

	var msgs []string
	var name string

	// Append (edited) when an edited message is received
	if message.SubType == "message_changed" {
		message = &slack.MessageEvent{Msg: *message.SubMessage}
		message.Text = fmt.Sprintf("%s (edited)", message.Text)
	}

	// Get username from cache
	name, ok := s.UserCache[message.User]

	// Name not in cache
	if !ok {
		if message.BotID != "" {
			// Name not found, perhaps a bot, use Username
			name, ok = s.UserCache[message.BotID]
			if !ok {
				// Not found in cache, add it
				name = message.Username
				s.UserCache[message.BotID] = message.Username
			}
		} else {
			// Not a bot, not in cache, get user info
			user, err := s.Client[clientId].GetUserInfo(message.User)
			if err != nil {
				name = "unknown"
				s.UserCache[message.User] = name
			} else {
				name = user.Name
				s.UserCache[message.User] = user.Name
			}
		}
	}

	if name == "" {
		name = "unknown"
	}

	// When there are attachments append them
	if len(message.Attachments) > 0 {
		msgs = append(msgs, createMessageFromAttachments(message.Attachments)...)
	}

	// Parse time
	floatTime, err := strconv.ParseFloat(message.Timestamp, 64)
	if err != nil {
		floatTime = 0.0
	}
	intTime := int64(floatTime)

	// Format message
	msg := s.FormatMessage(intTime, name, message.Text)

	msgs = append(msgs, msg)

	return msgs
}

func (s *SlackService) GetChannelName(channelId string) string {
	return s.JoinedChannels[channelId].Name
}

func (s *SlackService) GetChannelTopic(channelId string) string {
	return s.JoinedChannels[channelId].Topic
}

// createMessageFromAttachments will construct a array of string of the Field
// values of Attachments from a Message.
func createMessageFromAttachments(atts []slack.Attachment) []string {
	var msgs []string
	for _, att := range atts {
		for i := len(att.Fields) - 1; i >= 0; i-- {
			msgs = append(msgs,
				fmt.Sprintf(
					"%s %s",
					att.Fields[i].Title,
					att.Fields[i].Value,
				),
			)
		}

		if att.Text != "" {
			msgs = append(msgs, att.Text)
		}

		if att.Title != "" {
			msgs = append(msgs, att.Title)
		}
	}

	return msgs
}
