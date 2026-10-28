package tui

import "strings"

// Grid lays out fixed-width columns separated by a gap string. Cells may be
// pre-styled; padding is display-width aware, so ANSI codes don't skew columns.
type Grid struct {
	widths []int
	aligns []Align
	gap    string
}

// NewGrid creates a grid whose columns are joined by gap.
func NewGrid(gap string) *Grid { return &Grid{gap: gap} }

// Col appends a column of the given width and alignment.
func (g *Grid) Col(width int, a Align) *Grid {
	g.widths = append(g.widths, width)
	g.aligns = append(g.aligns, a)
	return g
}

// Width returns the total rendered width of a full row, gaps included.
func (g *Grid) Width() int {
	total := 0
	for _, w := range g.widths {
		total += w
	}
	if len(g.widths) > 1 {
		total += Width(g.gap) * (len(g.widths) - 1)
	}
	return total
}

// Row renders cells across the grid's columns. Extra cells are ignored; missing
// cells render as empty. Each cell is padded to its column width.
func (g *Grid) Row(cells ...string) string {
	parts := make([]string, len(g.widths))
	for i := range g.widths {
		cell := ""
		if i < len(cells) {
			cell = cells[i]
		}
		parts[i] = Pad(cell, g.widths[i], g.aligns[i])
	}
	return strings.Join(parts, g.gap)
}
