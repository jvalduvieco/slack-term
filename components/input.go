package components

import (
	"github.com/gizak/termui"
)

// Input is the definition of an Input component
type Input struct {
	par            *termui.Par
	text           []rune
	cursorPosition int
}

// CreateInput is the constructor of the Input struct
func CreateInput() *Input {
	input := &Input{
		par:            termui.NewPar(""),
		text:           make([]rune, 0),
		cursorPosition: 0,
	}

	input.par.Height = 3

	return input
}

// Buffer implements interface termui.Bufferer
func (i *Input) Buffer() termui.Buffer {
	buf := i.par.Buffer()

	// Set visible cursor
	char := buf.At(i.par.InnerX()+i.cursorPosition, i.par.Block.InnerY())
	buf.Set(
		i.par.InnerX()+i.cursorPosition,
		i.par.Block.InnerY(),
		termui.Cell{
			Ch: char.Ch,
			Fg: i.par.TextBgColor,
			Bg: i.par.TextFgColor,
		},
	)

	return buf
}

// GetHeight implements interface termui.GridBufferer
func (i *Input) GetHeight() int {
	return i.par.Block.GetHeight()
}

// SetWidth implements interface termui.GridBufferer
func (i *Input) SetWidth(w int) {
	i.par.SetWidth(w)
}

// SetX implements interface termui.GridBufferer
func (i *Input) SetX(x int) {
	i.par.SetX(x)
}

// SetY implements interface termui.GridBufferer
func (i *Input) SetY(y int) {
	i.par.SetY(y)
}

// Insert will insert a given key at the place of the current cursorPosition
func (i *Input) Insert(key rune) {
	if len(i.text) < i.par.InnerBounds().Dx()-1 {

		first := append(i.text[0:i.cursorPosition], key)
		i.text = append(first, i.text[i.cursorPosition:]...)

		i.par.Text = string(i.text)
		i.MoveCursorRight()
	}
}

// Backspace will remove a character in front of the cursorPosition
func (i *Input) Backspace() {
	if i.cursorPosition > 0 {
		i.text = append(i.text[0:i.cursorPosition-1], i.text[i.cursorPosition:]...)
		i.par.Text = string(i.text)
		i.MoveCursorLeft()
	}
}

// Delete will remove a character at the cursorPosition
func (i *Input) Delete() {
	if i.cursorPosition < len(i.text) {
		i.text = append(i.text[0:i.cursorPosition], i.text[i.cursorPosition+1:]...)
		i.par.Text = string(i.text)
	}
}

// MoveCursorRight will increase the current cursorPosition with 1
func (i *Input) MoveCursorRight() {
	if i.cursorPosition < len(i.text) {
		i.cursorPosition++
	}
}

// MoveCursorLeft will decrease the current cursorPosition with 1
func (i *Input) MoveCursorLeft() {
	if i.cursorPosition > 0 {
		i.cursorPosition--
	}
}

// IsEmpty will return true when the input is empty
func (i *Input) IsEmpty() bool {
	if i.par.Text == "" {
		return true
	}
	return false
}

// Clear will empty the input and move the cursor to the start position
func (i *Input) Clear() {
	i.text = make([]rune, 0)
	i.par.Text = ""
	i.cursorPosition = 0
}

// GetText returns the text currently in the input
func (i *Input) GetText() string {
	return i.par.Text
}
