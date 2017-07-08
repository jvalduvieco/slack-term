package views

import (
	"github.com/gizak/termui"

	"github.com/erroneousboat/slack-term/components"
	"github.com/erroneousboat/slack-term/config"
	"github.com/erroneousboat/slack-term/service"
)

type View struct {
	Input    *components.Input
	Chat     *components.Chat
	Channels *components.Channels
	Mode     *components.Mode
	Body     *termui.Grid
}

func CreateUIComponents(config *config.Config, svc *service.SlackService) *View {

	inputComponent := components.CreateInput()

	channelsComponent := components.CreateChannels(inputComponent.Par.Height)
	joined, _ := svc.GetChannels()
	channelsComponent.SetChannels(joined)

	selectedChannel := svc.JoinedChannels[channelsComponent.SelectedChannel]
	chatComponent := components.CreateChat(
		inputComponent.Par.Height,
		selectedChannel.Name,
		selectedChannel.Topic,
	)

	chatComponent.SetMessages(
		svc.GetMessages(
			svc.JoinedChannels[channelsComponent.SelectedChannel].SlackChannel,
			chatComponent.GetNumberOfMessagesVisible()))
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

func (v *View) Refresh() {
	termui.Render(
		v.Input,
		v.Chat,
		v.Channels,
		v.Mode,
	)
}
