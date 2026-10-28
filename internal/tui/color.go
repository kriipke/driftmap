// Package tui is a small terminal-UI toolkit: truecolor styling, display-width
// aware layout, rounded panels, grids, and bars. It has no dependencies beyond
// the standard library and is used to render DriftMap's summary dashboard.
package tui

import (
	"fmt"
	"strconv"
	"strings"
)

// Color is a 24-bit RGB color.
type Color struct{ R, G, B uint8 }

// RGB constructs a Color from red, green, and blue components.
func RGB(r, g, b uint8) Color { return Color{r, g, b} }

// Hex parses a "#rrggbb" (or "rrggbb") string into a Color. Malformed input
// yields black; callers pass compile-time constants, so this stays quiet.
func Hex(s string) Color {
	s = strings.TrimPrefix(s, "#")
	if len(s) != 6 {
		return Color{}
	}
	return Color{hexByte(s[0:2]), hexByte(s[2:4]), hexByte(s[4:6])}
}

func hexByte(s string) uint8 {
	v, _ := strconv.ParseUint(s, 16, 8)
	return uint8(v)
}

// fgSeq returns the SGR sequence that sets this color as the foreground.
func (c Color) fgSeq() string { return fmt.Sprintf("\x1b[38;2;%d;%d;%dm", c.R, c.G, c.B) }

// bgSeq returns the SGR sequence that sets this color as the background.
func (c Color) bgSeq() string { return fmt.Sprintf("\x1b[48;2;%d;%d;%dm", c.R, c.G, c.B) }
