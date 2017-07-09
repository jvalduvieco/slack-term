package components

import (
	"fmt"
	"strings"

	"github.com/gizak/termui"

	"github.com/erroneousboat/slack-term/service"
	"sort"
)

// Channels is the definition of a Channels component
type Channels struct {
	list               *termui.List
	channelIDs         map[string]int
	selectedListItemID int // index of which channel is selected from the list
	offset             int // from what offset are channels rendered
	cursorPosition     int // the y position of the 'cursor'
}

// CreateChannels is the constructor for the Channels component
func CreateChannels(inputHeight int) *Channels {
	channels := &Channels{
		list:       termui.NewList(),
		channelIDs: make(map[string]int),
	}

	channels.list.BorderLabel = "Channels"
	channels.list.Height = termui.TermHeight() - inputHeight

	channels.cursorPosition = channels.list.InnerBounds().Min.Y
	channels.selectedListItemID = 0
	channels.offset = 0
	return channels
}

// Buffer implements interface termui.Bufferer
func (c *Channels) Buffer() termui.Buffer {
	buf := c.list.Buffer()

	for i, item := range c.list.Items[c.offset:] {

		y := c.list.InnerBounds().Min.Y + i

		if y > c.list.InnerBounds().Max.Y-1 {
			break
		}

		var cells []termui.Cell
		if y == c.cursorPosition {
			cells = termui.DefaultTxBuilder.Build(
				item, c.list.ItemBgColor, c.list.ItemFgColor)
		} else {
			cells = termui.DefaultTxBuilder.Build(
				item, c.list.ItemFgColor, c.list.ItemBgColor)
		}

		cells = termui.DTrimTxCls(cells, c.list.InnerWidth())

		x := 0
		for _, cell := range cells {
			width := cell.Width()
			buf.Set(c.list.InnerBounds().Min.X+x, y, cell)
			x += width
		}

		// When not at the end of the pane fill it up empty characters
		for x < c.list.InnerBounds().Max.X-1 {
			if y == c.cursorPosition {
				buf.Set(x+1, y,
					termui.Cell{
						Ch: ' ',
						Fg: c.list.ItemBgColor,
						Bg: c.list.ItemFgColor,
					},
				)
			} else {
				buf.Set(
					x+1, y,
					termui.Cell{
						Ch: ' ',
						Fg: c.list.ItemFgColor,
						Bg: c.list.ItemBgColor,
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
	return c.list.Block.GetHeight()
}

// SetWidth implements interface termui.GridBufferer
func (c *Channels) SetWidth(w int) {
	c.list.SetWidth(w)
}

// SetX implements interface termui.GridBufferer
func (c *Channels) SetX(x int) {
	c.list.SetX(x)
}

// SetY implements interface termui.GridBufferer
func (c *Channels) SetY(y int) {
	c.list.SetY(y)
}

// SetChannels sets the channels available
func (c *Channels) SetChannels(channels service.Channels) {
	sort.Sort(channels)
	for i, channel := range channels {
		c.list.Items = append(c.list.Items, fmt.Sprintf(" [%s] %s", channel.ClientID, channel.Name))
		c.channelIDs[channel.ID] = i
	}
}

// SetSelectedItem sets the selectedListItemID given the index
func (c *Channels) SetSelectedItem(index int) {
	c.selectedListItemID = index
}

// GetSelectedChannelID returns the ID of the channel currently in front
func (c *Channels) GetSelectedChannelID() string {
	var result string

	for key, value := range c.channelIDs {
		if value == c.selectedListItemID {
			result = key
			break
		}
	}
	return result
}

// MoveCursorUp will decrease the selectedListItemID by 1
func (c *Channels) MoveCursorUp() {
	if c.selectedListItemID > 0 {
		c.MarkAsRead(c.GetSelectedChannelID())
		c.SetSelectedItem(c.selectedListItemID - 1)
		c.ScrollUp()
		c.MarkAsRead(c.GetSelectedChannelID())
	}
}

// MoveCursorDown will increase the selectedListItemID by 1
func (c *Channels) MoveCursorDown() {
	if c.selectedListItemID < len(c.list.Items)-1 {
		c.MarkAsRead(c.GetSelectedChannelID())
		c.SetSelectedItem(c.selectedListItemID + 1)
		c.ScrollDown()
		c.MarkAsRead(c.GetSelectedChannelID())
	}
}

// MoveCursorTop will move the cursor to the top of the channels
func (c *Channels) MoveCursorTop() {
	c.SetSelectedItem(0)
	c.cursorPosition = c.list.InnerBounds().Min.Y
	c.offset = 0
}

// MoveCursorBottom will move the cursor to the bottom of the channels
func (c *Channels) MoveCursorBottom() {
	c.SetSelectedItem(len(c.list.Items) - 1)

	offset := len(c.list.Items) - (c.list.InnerBounds().Max.Y - 1)

	if offset < 0 {
		c.offset = 0
		c.cursorPosition = c.selectedListItemID + 1
	} else {
		c.offset = offset
		c.cursorPosition = c.list.InnerBounds().Max.Y - 1
	}
}

// ScrollUp enables us to scroll through the channel list when it overflows
func (c *Channels) ScrollUp() {
	if c.cursorPosition == c.list.InnerBounds().Min.Y {
		if c.offset > 0 {
			c.offset--
		}
	} else {
		c.cursorPosition--
	}
}

// ScrollDown enables us to scroll through the channel list when it overflows
func (c *Channels) ScrollDown() {
	if c.cursorPosition == c.list.InnerBounds().Max.Y-1 {
		if c.offset < len(c.list.Items)-1 {
			c.offset++
		}
	} else {
		c.cursorPosition++
	}
}

// MarkAsUnread will be called when a new message arrives and will
// render an asterisk in front of the channel name
func (c *Channels) MarkAsUnread(channelID string) {
	if !strings.Contains(c.list.Items[c.channelIDs[channelID]], "*") {
		// The order of svc.Channels relates to the order of
		// list.Items, index will be the index of the channel
		c.list.Items[c.channelIDs[channelID]] = fmt.Sprintf(
			"*%s",
			strings.TrimSpace(c.list.Items[c.channelIDs[channelID]]))
	}

	// Play terminal bell sound
	fmt.Print("\a")
}

// MarkAsRead will remove the asterisk in front of a channel that
// received a new message. This will happen as one will move up or down the
// cursor for Channels
func (c *Channels) MarkAsRead(channelID string) {
	var listItemID = c.channelIDs[channelID]
	channelName := strings.Split(c.list.Items[listItemID], "*")
	if len(channelName) > 1 {
		c.list.Items[listItemID] = fmt.Sprintf(" %s", channelName[1])
	} else {
		c.list.Items[listItemID] = channelName[0]
	}
}
