package views

import (
	"github.com/gizak/termui"

	"github.com/erroneousboat/slack-term/components"
	"github.com/erroneousboat/slack-term/config"
	"github.com/erroneousboat/slack-term/service"
)

// View contains all app widgets
type View struct {
	Input    *components.Input
	Chat     *components.Chat
	Channels *components.Channels
	Mode     *components.Mode
	Body     *termui.Grid
}

// CreateUIComponents builds all the widgets needed for the app
func CreateUIComponents(config *config.Config, svc *service.SlackService) *View {

	inputComponent := components.CreateInput()

	channelsComponent := components.CreateChannels(inputComponent.GetHeight())
	channels := svc.GetChannelList()
	channelsComponent.SetChannels(channels)

	chatComponent := components.CreateChat(
		inputComponent.GetHeight(),
		svc.GetChannelName(channelsComponent.GetSelectedChannelID()),
		svc.GetChannelTopic(channelsComponent.GetSelectedChannelID()),
	)

	chatComponent.AddMessages(
		svc.GetMessages(
			channelsComponent.GetSelectedChannelID(),
			chatComponent.GetMaxNumberOfMessagesVisible()))
	modeComponent := components.CreateMode()

	// Setup body
	termui.Body.AddRows(
		termui.NewRow(
			termui.NewCol(config.SidebarWidth, 0, channelsComponent),
			termui.NewCol(config.MainWidth, 0, chatComponent),
		),
		termui.NewRow(
			termui.NewCol(config.SidebarWidth, 0, modeComponent),
			termui.NewCol(config.MainWidth, 0, inputComponent),
		),
	)
	termui.Body.Align()
	termui.Render(termui.Body)

	view := &View{
		Input:    inputComponent,
		Channels: channelsComponent,
		Chat:     chatComponent,
		Mode:     modeComponent,
		Body:     termui.Body,
	}

	return view
}

// Refresh renders all widgets on demand
func (v *View) Refresh() {
	termui.Render(
		v.Input,
		v.Chat,
		v.Channels,
		v.Mode,
	)
}
