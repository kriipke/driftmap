package tui

import "strings"

// Panel draws a rounded, single-line box. Inner is the content width between
// the one-space inner margins; the drawn box is Inner+4 columns wide. When
// Background is set, every row is filled edge to edge so the panel reads as a
// solid card against the surrounding terminal.
type Panel struct {
	Inner      int
	Border     Style
	Background *Color
}

// border styles the given box-drawing text with the panel's border style.
func (p Panel) border(s string) string { return p.Border.Render(s) }

// fill paints the panel background across the full width of a rendered row. It
// re-asserts the background after every inner reset so nested styles (which end
// in a plain reset) don't punch holes in the fill.
func (p Panel) fill(s string) string {
	if p.Background == nil || !colorEnabled {
		return s
	}
	seq := p.Background.bgSeq()
	return seq + strings.ReplaceAll(s, reset, reset+seq) + reset
}

// Top renders the top edge: ╭────╮.
func (p Panel) Top() string {
	return p.fill(p.border("╭" + strings.Repeat("─", p.Inner+2) + "╮"))
}

// Bottom renders the bottom edge: ╰────╯.
func (p Panel) Bottom() string {
	return p.fill(p.border("╰" + strings.Repeat("─", p.Inner+2) + "╯"))
}

// Sep renders an interior horizontal divider: ├────┤.
func (p Panel) Sep() string {
	return p.fill(p.border("├" + strings.Repeat("─", p.Inner+2) + "┤"))
}

// Line renders one content row, padding content (which may contain ANSI codes)
// to the inner width and framing it with vertical borders.
func (p Panel) Line(content string) string {
	return p.fill(p.border("│") + " " + PadRight(content, p.Inner) + " " + p.border("│"))
}

// Split places left and right content on one row, filling the gap between them
// so right sits flush against the inner right edge.
func (p Panel) Split(left, right string) string {
	gap := p.Inner - Width(left) - Width(right)
	if gap < 1 {
		gap = 1
	}
	return p.Line(left + strings.Repeat(" ", gap) + right)
}
