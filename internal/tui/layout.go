package tui

import "strings"

// Align is a horizontal alignment within a fixed-width cell.
type Align int

const (
	Left Align = iota
	Right
	Center
)

// PadRight pads s on the right with spaces to reach display width w.
func PadRight(s string, w int) string {
	if d := w - Width(s); d > 0 {
		return s + strings.Repeat(" ", d)
	}
	return s
}

// PadLeft pads s on the left with spaces to reach display width w.
func PadLeft(s string, w int) string {
	if d := w - Width(s); d > 0 {
		return strings.Repeat(" ", d) + s
	}
	return s
}

// PadCenter centers s within display width w, biasing extra space to the right.
func PadCenter(s string, w int) string {
	d := w - Width(s)
	if d <= 0 {
		return s
	}
	left := d / 2
	return strings.Repeat(" ", left) + s + strings.Repeat(" ", d-left)
}

// Pad aligns s to width w per a.
func Pad(s string, w int, a Align) string {
	switch a {
	case Right:
		return PadLeft(s, w)
	case Center:
		return PadCenter(s, w)
	default:
		return PadRight(s, w)
	}
}

// Truncate shortens plain (unstyled) text to at most w columns, adding an
// ellipsis when clipped. Callers truncate before styling so color codes stay
// intact.
func Truncate(s string, w int) string {
	if w <= 0 {
		return ""
	}
	if Width(s) <= w {
		return s
	}
	if w == 1 {
		return "…"
	}
	var b strings.Builder
	used := 0
	for _, r := range StripANSI(s) {
		rw := runeWidth(r)
		if used+rw > w-1 {
			break
		}
		b.WriteRune(r)
		used += rw
	}
	b.WriteRune('…')
	return b.String()
}

// WrapWords greedily wraps space-separated words to lines of at most w columns.
func WrapWords(words []string, sep string, w int) []string {
	var lines []string
	cur := ""
	for _, word := range words {
		candidate := word
		if cur != "" {
			candidate = cur + sep + word
		}
		if Width(candidate) > w && cur != "" {
			lines = append(lines, cur)
			cur = word
			continue
		}
		cur = candidate
	}
	if cur != "" {
		lines = append(lines, cur)
	}
	return lines
}
