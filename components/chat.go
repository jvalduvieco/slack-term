package components

import (
	"fmt"
	"html"
	"sort"
	"strings"

	"github.com/gizak/termui"

	"github.com/erroneousboat/slack-term/config"
)

// Chat is the definition of a Chat component
type Chat struct {
	list   *termui.List
	offset int
}

// CreateChat is the constructor for the Chat struct
func CreateChat(inputHeight int, name string, topic string) *Chat {
	chat := &Chat{
		list:   termui.NewList(),
		offset: 0,
	}

	chat.list.Height = termui.TermHeight() - inputHeight
	chat.list.Overflow = "wrap"

	chat.SetBorderLabel(name, topic)

	return chat
}

// Buffer implements interface termui.Bufferer
func (c *Chat) Buffer() termui.Buffer {
	// Build cells, after every item put a newline
	cells := termui.DefaultTxBuilder.Build(
		strings.Join(c.list.Items, "\n"),
		c.list.ItemFgColor, c.list.ItemBgColor,
	)

	// We will create an array of Line structs, this allows us
	// to more easily render the items in a list. We will range
	// over the cells we've created and create a Line within
	// the bounds of the Chat pane
	type Line struct {
		cells []termui.Cell
	}

	lines := []Line{}
	line := Line{}

	x := 0
	for _, cell := range cells {

		if cell.Ch == '\n' {
			lines = append(lines, line)
			line = Line{}
			x = 0
			continue
		}

		if x+cell.Width() > c.list.InnerBounds().Dx() {
			lines = append(lines, line)
			line = Line{}
			x = 0
		}

		line.cells = append(line.cells, cell)
		x++
	}
	lines = append(lines, line)

	// We will print lines bottom up, it will loop over the lines
	// backwards and for every line it'll set the cell in that line.
	// offset is the number which allows us to begin printing the
	// line above the last line.
	buf := c.list.Buffer()
	linesHeight := len(lines)
	paneMinY := c.list.InnerBounds().Min.Y
	paneMaxY := c.list.InnerBounds().Max.Y

	currentY := paneMaxY - 1
	for i := (linesHeight - 1) - c.offset; i >= 0; i-- {
		if currentY < paneMinY {
			break
		}

		x := c.list.InnerBounds().Min.X
		for _, cell := range lines[i].cells {
			buf.Set(x, currentY, cell)
			x += cell.Width()
		}

		// When we're not at the end of the pane, fill it up
		// with empty characters
		for x < c.list.InnerBounds().Max.X {
			buf.Set(
				x, currentY,
				termui.Cell{
					Ch: ' ',
					Fg: c.list.ItemFgColor,
					Bg: c.list.ItemBgColor,
				},
			)
			x++
		}
		currentY--
	}

	// If the space above currentY is empty we need to fill
	// it up with blank lines, otherwise the list object will
	// render the items top down, and the result will mix.
	for currentY >= paneMinY {
		x := c.list.InnerBounds().Min.X
		for x < c.list.InnerBounds().Max.X {
			buf.Set(
				x, currentY,
				termui.Cell{
					Ch: ' ',
					Fg: c.list.ItemFgColor,
					Bg: c.list.ItemBgColor,
				},
			)
			x++
		}
		currentY--
	}

	return buf
}

// GetHeight implements interface termui.GridBufferer
func (c *Chat) GetHeight() int {
	return c.list.Block.GetHeight()
}

// SetWidth implements interface termui.GridBufferer
func (c *Chat) SetWidth(w int) {
	c.list.SetWidth(w)
}

// SetX implements interface termui.GridBufferer
func (c *Chat) SetX(x int) {
	c.list.SetX(x)
}

// SetY implements interface termui.GridBufferer
func (c *Chat) SetY(y int) {
	c.list.SetY(y)
}

// GetMaxNumberOfMessagesVisible returns the maximum numner of messages visible within the widget
func (c *Chat) GetMaxNumberOfMessagesVisible() int {
	return c.list.InnerBounds().Max.Y - c.list.InnerBounds().Min.Y
}

// AddMessages adds an array of mesages into the widget
func (c *Chat) AddMessages(messages []string) {
	for _, message := range messages {
		c.AddMessage(message)
	}
}

// AddMessage adds a single message to list.Items
func (c *Chat) AddMessage(message string) {
	c.list.Items = append(c.list.Items, html.UnescapeString(message))
}

// ClearMessages clear the list.Items
func (c *Chat) ClearMessages() {
	c.list.Items = []string{}
}

// ScrollUp will render the chat messages based on the offset of the Chat
// pane.
//
// offset is 0 when scrolled down. (we loop backwards over the array, so we
// start with rendering last item in the list at the maximum y of the Chat
// pane). Increasing the offset will thus result in substracting the offset
// from the len(Chat.list.Items).
func (c *Chat) ScrollUp() {
	c.offset = c.offset + 10

	// Protect overscrolling
	if c.offset > len(c.list.Items)-1 {
		c.offset = len(c.list.Items) - 1
	}
}

// ScrollDown will render the chat messages based on the offset of the Chat
// pane.
//
// offset is 0 when scrolled down. (we loop backwards over the array, so we
// start with rendering last item in the list at the maximum y of the Chat
// pane). Increasing the offset will thus result in substracting the offset
// from the len(Chat.list.Items).
func (c *Chat) ScrollDown() {
	c.offset = c.offset - 10

	// Protect overscrolling
	if c.offset < 0 {
		c.offset = 0
	}
}

// SetBorderLabel will set Label of the Chat pane to the specified string
func (c *Chat) SetBorderLabel(name string, topic string) {
	var channelName string
	if topic != "" {
		channelName = fmt.Sprintf("%s - %s",
			name,
			topic,
		)
	} else {
		channelName = name
	}
	c.list.BorderLabel = channelName
}

// ShowHelp shows the usage and key bindings in the chat pane
func (c *Chat) ShowHelp(cfg *config.Config) {
	help := []string{
		"slack-term - slack client for your terminal",
		"",
		"USAGE:",
		"    slack-term -config [path-to-config]",
		"",
		"KEY BINDINGS:",
		"",
	}

	for mode, mapping := range cfg.KeyMap {
		help = append(help, fmt.Sprintf("    %s", strings.ToUpper(mode)))
		help = append(help, "")

		var keys []string
		for k := range mapping {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			help = append(help, fmt.Sprintf("    %-12s%-15s", k, mapping[k]))
		}
		help = append(help, "")
	}

	c.list.Items = help
}
