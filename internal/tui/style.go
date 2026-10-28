package tui

import (
	"os"
	"strings"
)

const reset = "\x1b[0m"

// colorEnabled gates all ANSI emission. It is disabled when NO_COLOR is set or
// when stdout is not a terminal (e.g. piped), so redirected output stays clean.
var colorEnabled = true

func init() {
	// Precedence: NO_COLOR disables; CLICOLOR_FORCE forces on; otherwise emit
	// color only to a terminal so redirected/piped output stays clean.
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		colorEnabled = false
		return
	}
	if _, ok := os.LookupEnv("CLICOLOR_FORCE"); ok {
		colorEnabled = true
		return
	}
	if fi, err := os.Stdout.Stat(); err == nil {
		if fi.Mode()&os.ModeCharDevice == 0 {
			colorEnabled = false
		}
	}
}

// SetColorEnabled forces color output on or off, overriding auto-detection.
func SetColorEnabled(b bool) { colorEnabled = b }

// ColorEnabled reports whether styled output currently emits ANSI codes.
func ColorEnabled() bool { return colorEnabled }

// Style is an immutable set of text attributes. Methods return a copy, so
// styles compose fluently: tui.New().Foreground(c).Bold().
type Style struct {
	fg, bg              *Color
	bold, faint, italic bool
}

// New returns an empty style.
func New() Style { return Style{} }

// Foreground sets the text color.
func (s Style) Foreground(c Color) Style { s.fg = &c; return s }

// Background sets the fill color.
func (s Style) Background(c Color) Style { s.bg = &c; return s }

// Bold enables bold weight.
func (s Style) Bold() Style { s.bold = true; return s }

// Faint enables the dim/faint attribute.
func (s Style) Faint() Style { s.faint = true; return s }

// Italic enables italics.
func (s Style) Italic() Style { s.italic = true; return s }

// Render wraps text in this style's escape codes (a no-op when color is off).
func (s Style) Render(text string) string {
	if !colorEnabled {
		return text
	}
	var b strings.Builder
	if s.bold {
		b.WriteString("\x1b[1m")
	}
	if s.faint {
		b.WriteString("\x1b[2m")
	}
	if s.italic {
		b.WriteString("\x1b[3m")
	}
	if s.fg != nil {
		b.WriteString(s.fg.fgSeq())
	}
	if s.bg != nil {
		b.WriteString(s.bg.bgSeq())
	}
	if b.Len() == 0 {
		return text
	}
	b.WriteString(text)
	b.WriteString(reset)
	return b.String()
}
