package service

import (
	"fmt"
	"log"
	"strconv"
	"time"

	slack "github.com/nlopes/slack"
)

// SlackService is the service that manages slack connections
type SlackService struct {
	Client           map[string]*slack.Client
	RTM              map[string]*slack.RTM
	joinedChannels   map[string]Channel
	unjoinedChannels map[string]Channel
	userCache        map[string]string
	currentUserID    map[string]string
}
// Channel represents a slack channel within this app
type Channel struct {
	ID           string
	Name         string
	Topic        string
	SlackChannel interface{}
	ClientID     string
	ChannelType  ChannelType
}

// Channels is an array of Channel, mainly to sort
type Channels []Channel

func (s Channels) Len() int {
	return len(s)
}

func (s Channels) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s Channels) Less(i, j int) bool {
	var first = fmt.Sprintf("%s %d %s", s[i].ClientID, s[i].ChannelType, s[i].Name)
	var second = fmt.Sprintf("%s %d %s", s[j].ClientID, s[j].ChannelType, s[j].Name)
	return first < second
}

// ChannelType Type of the channel, see constant below
type ChannelType uint8

const (
	// CHANNEL Constants that define channel type. Also used to order channels in the channels widget
	CHANNEL ChannelType = iota + 1
	// GROUP for group chats
	GROUP
	// IM for direct messages
	IM
)

// CreateSlackService is the constructor for the SlackService and will initialize
// the RTM and a ClientID
func CreateSlackService(tokens map[string]string) *SlackService {
	svc := &SlackService{
		Client:           make(map[string]*slack.Client),
		RTM:              make(map[string]*slack.RTM),
		joinedChannels:   make(map[string]Channel),
		unjoinedChannels: make(map[string]Channel),
		userCache:        make(map[string]string),
		currentUserID:    make(map[string]string),
	}

	for clientID, token := range tokens {
		svc.Client[clientID] = slack.New(token)

		// Get channelUser associated with token, mainly
		// used to identify channelUser when new messages
		// arrives
		authTest, err := svc.Client[clientID].AuthTest()
		if err != nil {
			log.Fatal("ERROR: not able to authorize client, check your connection and/or slack-token")
		}
		svc.currentUserID[clientID] = authTest.UserID

		// Create RTM
		svc.RTM[clientID] = svc.Client[clientID].NewRTM()
		go svc.RTM[clientID].ManageConnection()

		// Creation of channelUser cache this speeds up
		// the uncovering of usernames of messages
		users, _ := svc.Client[clientID].GetUsers()
		for _, channelUser := range users {
			// only add non-deleted users
			if !channelUser.Deleted {
				svc.userCache[channelUser.ID] = channelUser.Name
			}
		}
	}

	return svc
}

// updateChannels will retrieve all available channels, groups, and im channels.
// We will return different channel collections, first channels the user is a member of
// and secondly a list of unarchived channels the user can join
// Because the channels are of different types, we will append them to
// an []interface as well as to a []Channel which will give us easy access
// to the id and name of the Channel.
func (s *SlackService) updateChannels() {
	// FIXME Check errors
	for currentClientID := range s.Client {
		// Channel
		_ = s.fetchChannels(currentClientID)

		// Groups
		_ = s.fetchGroups(currentClientID)

		// IM
		_ = s.fetchIM(currentClientID)
	}
}

// GetChannelList returns a list of all channels
func (s *SlackService) GetChannelList() []Channel {
	s.updateChannels()
	var result Channels
	for _, channel := range s.joinedChannels {
		result = append(result, channel)
	}

	return result
}

// GetCurrentUserID returns the current user ID
func (s *SlackService) GetCurrentUserID(clientID string) string {
	return s.currentUserID[clientID]
}

func (s *SlackService) fetchIM(currentClientID string) error {
	slackIM, err := s.Client[currentClientID].GetIMChannels()
	if err != nil {
		//chans = append(chans, Channel{})
	}
	for _, im := range slackIM {

		// Uncover name, when we can't uncover name for
		// IM channel this is then probably a deleted
		// user, because we wont add deleted users
		// to the userCache, so we skip it
		name, ok := s.userCache[im.User]
		if ok {
			s.joinedChannels[im.ID] = Channel{im.ID, name, "", im, currentClientID, IM}
		}
	}
	return err
}

func (s *SlackService) fetchGroups(currentClientID string) error {
	slackGroups, err := s.Client[currentClientID].GetGroups(true)
	if err != nil {
		//chans = append(chans, Channel{})
	}
	for _, grp := range slackGroups {
		s.joinedChannels[grp.ID] = Channel{grp.ID, grp.Name, grp.Topic.Value, grp, currentClientID, GROUP}
	}
	return err
}

func (s *SlackService) fetchChannels(currentClientID string) error {
	slackChans, err := s.Client[currentClientID].GetChannels(true)
	if err != nil {
		//chans = append(chans, Channel{})
	}
	for _, chn := range slackChans {
		if chn.IsMember {
			s.joinedChannels[chn.ID] = Channel{chn.ID, chn.Name, chn.Topic.Value, chn, currentClientID, CHANNEL}
		} else {
			s.unjoinedChannels[chn.ID] = Channel{chn.ID, chn.Name, chn.Topic.Value, chn, currentClientID, CHANNEL}
		}
	}
	return err
}

// SetChannelReadMark will set the read mark for a channel, group, and im
// channel based on the current time.
func (s *SlackService) SetChannelReadMark(channelID string) {
	selectedChannel := s.joinedChannels[channelID]
	switch channel := selectedChannel.SlackChannel.(type) {
	case slack.Channel:
		s.Client[selectedChannel.ClientID].SetChannelReadMark(
			channel.ID, fmt.Sprintf("%f",
				float64(time.Now().Unix())),
		)
	case slack.Group:
		s.Client[selectedChannel.ClientID].SetGroupReadMark(
			channel.ID, fmt.Sprintf("%f",
				float64(time.Now().Unix())),
		)
	case slack.IM:
		s.Client[selectedChannel.ClientID].MarkIMChannel(
			channel.ID, fmt.Sprintf("%f",
				float64(time.Now().Unix())),
		)
	}
}

// SendMessage will send a message to a particular channel
func (s *SlackService) SendMessage(channelID string, message string) {
	currentChannel := s.joinedChannels[channelID]
	// https://godoc.org/github.com/nlopes/slack#PostMessageParameters
	postParams := slack.PostMessageParameters{
		AsUser: true,
	}

	// https://godoc.org/github.com/nlopes/slack#Client.PostMessage
	s.Client[currentChannel.ClientID].PostMessage(channelID, message, postParams)
}

// GetMessages will get messages for a channel, group or im channel delimited
// by a count.
func (s *SlackService) GetMessages(channelID string, count int) []string {
	channel := s.joinedChannels[channelID]
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
		history, err = s.Client[channel.ClientID].GetChannelHistory(chnType.ID, historyParams)
		if err != nil {
			log.Fatal(err) // FIXME
		}
	case slack.Group:
		history, err = s.Client[channel.ClientID].GetGroupHistory(chnType.ID, historyParams)
		if err != nil {
			log.Fatal(err) // FIXME
		}
	case slack.IM:
		history, err = s.Client[channel.ClientID].GetIMHistory(chnType.ID, historyParams)
		if err != nil {
			log.Fatal(err) // FIXME
		}
	}

	// Construct the messages
	var messages []string
	for _, message := range history.Messages {
		msg := s.CreateMessage(message, channel.ClientID)
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
func (s *SlackService) CreateMessage(message slack.Message, clientID string) []string {
	var msgs []string
	var name string

	// Get username from cache
	name, ok := s.userCache[message.User]

	// Name not in cache
	if !ok {
		if message.BotID != "" {
			// Name not found, perhaps a bot, use Username
			name, ok = s.userCache[message.BotID]
			if !ok {
				// Not found in cache, add it
				name = message.Username
				s.userCache[message.BotID] = message.Username
			}
		} else {
			// Not a bot, not in cache, get user info
			channelUser, err := s.Client[clientID].GetUserInfo(message.User)
			if err != nil {
				name = "unknown"
				s.userCache[message.User] = name
			} else {
				name = channelUser.Name
				s.userCache[message.User] = channelUser.Name
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
	msg := s.formatMessage(intTime, name, message.Text)

	msgs = append(msgs, msg)

	return msgs
}

func (s *SlackService) formatMessage(intTime int64, name string, message string) string {
	msg := fmt.Sprintf(
		"[%s] <[%s](fg-green)> %s",
		time.Unix(intTime, 0).Format("15:04"),
		name,
		message,
	)
	return msg
}

// CreateMessageFromMessageEvent creates a message from an event
func (s *SlackService) CreateMessageFromMessageEvent(message *slack.MessageEvent, clientID string) []string {

	var msgs []string
	var name string

	// Append (edited) when an edited message is received
	if message.SubType == "message_changed" {
		message = &slack.MessageEvent{Msg: *message.SubMessage}
		message.Text = fmt.Sprintf("%s (edited)", message.Text)
	}

	// Get username from cache
	name, ok := s.userCache[message.User]

	// Name not in cache
	if !ok {
		if message.BotID != "" {
			// Name not found, perhaps a bot, use Username
			name, ok = s.userCache[message.BotID]
			if !ok {
				// Not found in cache, add it
				name = message.Username
				s.userCache[message.BotID] = message.Username
			}
		} else {
			// Not a bot, not in cache, get user info
			user, err := s.Client[clientID].GetUserInfo(message.User)
			if err != nil {
				name = "unknown"
				s.userCache[message.User] = name
			} else {
				name = user.Name
				s.userCache[message.User] = user.Name
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
	msg := s.formatMessage(intTime, name, message.Text)

	msgs = append(msgs, msg)

	return msgs
}

// GetChannelName returns the channel name
func (s *SlackService) GetChannelName(channelID string) string {
	return s.joinedChannels[channelID].Name
}

// GetChannelTopic returns the channel topic
func (s *SlackService) GetChannelTopic(channelID string) string {
	return s.joinedChannels[channelID].Topic
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
