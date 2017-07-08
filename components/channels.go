package components

import (
	"fmt"
	"strings"

	"github.com/gizak/termui"

	"github.com/erroneousboat/slack-term/service"
)

// Channels is the definition of a Channels component
type Channels struct {
	List               *termui.List
	channelIds		   map[string]int
	SelectedListItemId int // index of which channel is selected from the List
	Offset             int    // from what offset are channels rendered
	CursorPosition     int    // the y position of the 'cursor'
}

// CreateChannels is the constructor for the Channels component
func CreateChannels(inputHeight int) *Channels {
	channels := &Channels{
		List: termui.NewList(),
		channelIds: make(map[string]int),
	}

	channels.List.BorderLabel = "Channels"
	channels.List.Height = termui.TermHeight() - inputHeight

	channels.CursorPosition = channels.List.InnerBounds().Min.Y
	channels.SelectedListItemId = 0
	channels.Offset = 0
	return channels
}

// Buffer implements interface termui.Bufferer
func (c *Channels) Buffer() termui.Buffer {
	buf := c.List.Buffer()

	for i, item := range c.List.Items[c.Offset:] {

		y := c.List.InnerBounds().Min.Y + i

		if y > c.List.InnerBounds().Max.Y-1 {
			break
		}

		var cells []termui.Cell
		if y == c.CursorPosition {
			cells = termui.DefaultTxBuilder.Build(
				item, c.List.ItemBgColor, c.List.ItemFgColor)
		} else {
			cells = termui.DefaultTxBuilder.Build(
				item, c.List.ItemFgColor, c.List.ItemBgColor)
		}

		cells = termui.DTrimTxCls(cells, c.List.InnerWidth())

		x := 0
		for _, cell := range cells {
			width := cell.Width()
			buf.Set(c.List.InnerBounds().Min.X+x, y, cell)
			x += width
		}

		// When not at the end of the pane fill it up empty characters
		for x < c.List.InnerBounds().Max.X-1 {
			if y == c.CursorPosition {
				buf.Set(x+1, y,
					termui.Cell{
						Ch: ' ',
						Fg: c.List.ItemBgColor,
						Bg: c.List.ItemFgColor,
					},
				)
			} else {
				buf.Set(
					x+1, y,
					termui.Cell{
						Ch: ' ',
						Fg: c.List.ItemFgColor,
						Bg: c.List.ItemBgColor,
					},
				)
			}
			x++
		}
	}

	return buf
}

// GetHeight implements interface termui.GridBufferer
func (c *Channels) GetHeight() int {
	return c.List.Block.GetHeight()
}

// SetWidth implements interface termui.GridBufferer
func (c *Channels) SetWidth(w int) {
	c.List.SetWidth(w)
}

// SetX implements interface termui.GridBufferer
func (c *Channels) SetX(x int) {
	c.List.SetX(x)
}

// SetY implements interface termui.GridBufferer
func (c *Channels) SetY(y int) {
	c.List.SetY(y)
}

func (c *Channels) SetChannels(channels []service.Channel) {
	for i, channel := range channels {
		c.List.Items = append(c.List.Items, fmt.Sprintf(" [%s] %s", channel.ClientId, channel.Name))
		c.channelIds[channel.Id] = i
	}
}

// SetSelectedItem sets the SelectedListItemId given the index
func (c *Channels) SetSelectedItem(index int) {
	c.SelectedListItemId = index
}

func (c *Channels) GetSelectedChannelId() string {
	var result string

	for key,value := range c.channelIds {
		if value == c.SelectedListItemId {
			result = key
			break
		}
	}
	return result
}
// MoveCursorUp will decrease the SelectedListItemId by 1
func (c *Channels) MoveCursorUp() {
	if c.SelectedListItemId > 0 {
		c.SetSelectedItem(c.SelectedListItemId - 1)
		c.ScrollUp()
		c.MarkAsRead()
	}
}

// MoveCursorDown will increase the SelectedListItemId by 1
func (c *Channels) MoveCursorDown() {
	if c.SelectedListItemId < len(c.List.Items)-1 {
		c.SetSelectedItem(c.SelectedListItemId + 1)
		c.ScrollDown()
		c.MarkAsRead()
	}
}

// MoveCursorTop will move the cursor to the top of the channels
func (c *Channels) MoveCursorTop() {
	c.SetSelectedItem(0)
	c.CursorPosition = c.List.InnerBounds().Min.Y
	c.Offset = 0
}

// MoveCursorBottom will move the cursor to the bottom of the channels
func (c *Channels) MoveCursorBottom() {
	c.SetSelectedItem(len(c.List.Items) - 1)

	offset := len(c.List.Items) - (c.List.InnerBounds().Max.Y - 1)

	if offset < 0 {
		c.Offset = 0
		c.CursorPosition = c.SelectedListItemId + 1
	} else {
		c.Offset = offset
		c.CursorPosition = c.List.InnerBounds().Max.Y - 1
	}
}

// ScrollUp enables us to scroll through the channel list when it overflows
func (c *Channels) ScrollUp() {
	if c.CursorPosition == c.List.InnerBounds().Min.Y {
		if c.Offset > 0 {
			c.Offset--
		}
	} else {
		c.CursorPosition--
	}
}

// ScrollDown enables us to scroll through the channel list when it overflows
func (c *Channels) ScrollDown() {
	if c.CursorPosition == c.List.InnerBounds().Max.Y-1 {
		if c.Offset < len(c.List.Items)-1 {
			c.Offset++
		}
	} else {
		c.CursorPosition++
	}
}

// MarkAsUnread will be called when a new message arrives and will
// render an asterisk in front of the channel name
func (c *Channels) MarkAsUnread(channelID string) {
	if !strings.Contains(c.List.Items[c.channelIds[channelID]], "*") {
		// The order of svc.Channels relates to the order of
		// List.Items, index will be the index of the channel
		c.List.Items[c.channelIds[channelID]] = fmt.Sprintf(
			"* %s",
			strings.TrimSpace(c.List.Items[c.channelIds[channelID]]))
	}

	// Play terminal bell sound
	fmt.Print("\a")
}

// MarkAsRead will remove the asterisk in front of a channel that
// received a new message. This will happen as one will move up or down the
// cursor for Channels
func (c *Channels) MarkAsRead() {
	channelName := strings.Split(c.List.Items[c.SelectedListItemId], "* ")
	if len(channelName) > 1 {
		c.List.Items[c.SelectedListItemId] = fmt.Sprintf("  %s", channelName[1])
	} else {
		c.List.Items[c.SelectedListItemId] = channelName[0]
	}
}
