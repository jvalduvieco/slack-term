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
	input := components.CreateInput()

	channels := components.CreateChannels(input.Par.Height)
	joined, _ := svc.GetChannels()
	channels.SetChannels(joined)

	chat := components.CreateChat(
		svc,
		input.Par.Height,
		svc.JoinedSlackChannels[channels.SelectedChannel],
		svc.JoinedChannels[channels.SelectedChannel],
	)

	mode := components.CreateMode()

	view := &View{
		Input:    input,
		Channels: channels,
		Chat:     chat,
		Mode:     mode,
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
