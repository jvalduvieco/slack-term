package components

import "github.com/gizak/termui"

// Mode is the definition of Mode component
type Mode struct {
	par *termui.Par
}

// CreateMode is the constructor of the Mode struct
func CreateMode() *Mode {
	mode := &Mode{
		par: termui.NewPar("COMMAND"),
	}

	mode.par.Height = 3

	return mode
}

// Buffer implements interface termui.Bufferer
func (m *Mode) Buffer() termui.Buffer {
	buf := m.par.Buffer()

	// Center text
	space := m.par.InnerWidth()
	word := len(m.par.Text)

	midSpace := space / 2
	midWord := word / 2

	start := midSpace - midWord

	cells := termui.DefaultTxBuilder.Build(
		m.par.Text, m.par.TextFgColor, m.par.TextBgColor)

	i, j := 0, 0
	x := m.par.InnerBounds().Min.X
	for x < m.par.InnerBounds().Max.X {
		if i < start {
			buf.Set(
				x, m.par.InnerY(),
				termui.Cell{
					Ch: ' ',
					Fg: m.par.TextFgColor,
					Bg: m.par.TextBgColor,
				},
			)
			x++
			i++
		} else {
			if j < len(cells) {
				buf.Set(x, m.par.InnerY(), cells[j])
				i++
				j++
			}
			x++
		}
	}

	return buf
}

// GetHeight implements interface termui.GridBufferer
func (m *Mode) GetHeight() int {
	return m.par.Block.GetHeight()
}

// SetWidth implements interface termui.GridBufferer
func (m *Mode) SetWidth(w int) {
	m.par.SetWidth(w)
}

// SetX implements interface termui.GridBufferer
func (m *Mode) SetX(x int) {
	m.par.SetX(x)
}

// SetY implements interface termui.GridBufferer
func (m *Mode) SetY(y int) {
	m.par.SetY(y)
}

// SetText sets the widget text
func (m *Mode) SetText(text string) {
	m.par.Text = text
}
