package views

import (
	"github.com/gizak/termui"

	"github.com/erroneousboat/slack-term/components"
	"github.com/erroneousboat/slack-term/service"
)

type View struct {
	Input    *components.Input
	Chat     *components.Chat
	Channels *components.Channels
	Mode     *components.Mode
}

func CreateUIComponents(svc *service.SlackService) *View {
	inputComponent := components.CreateInput()

	channelsComponent := components.CreateChannels(inputComponent.Par.Height)
	joined, _ := svc.GetChannels()
	channelsComponent.SetChannels(joined)

	chatComponent := components.CreateChat(
		inputComponent.Par.Height,
		svc.JoinedSlackChannels[channelsComponent.SelectedChannel],
		svc.JoinedChannels[channelsComponent.SelectedChannel],
	)


	chatComponent.SetMessages(
		svc.GetMessages(
			svc.JoinedSlackChannels[channelsComponent.SelectedChannel],
			chatComponent.GetNumberOfMessagesVisible()))
	modeComponent := components.CreateMode()

	view := &View{
		Input:    inputComponent,
		Channels: channelsComponent,
		Chat:     chatComponent,
		Mode:     modeComponent,
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
