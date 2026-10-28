package tui

import "strings"

// SegmentBar renders a segmented meter, e.g. ■■■■□ for filled=4, total=5.
// The filled segments use on; the remainder use off.
func SegmentBar(filled, total int, on, off Style) string {
	if filled < 0 {
		filled = 0
	}
	if filled > total {
		filled = total
	}
	var b strings.Builder
	b.WriteString(on.Render(strings.Repeat("■", filled)))
	if rest := total - filled; rest > 0 {
		b.WriteString(off.Render(strings.Repeat("□", rest)))
	}
	return b.String()
}
