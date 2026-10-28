package tui

import "regexp"

// ansiSGR matches Select-Graphic-Rendition escape sequences (colors, bold, …).
var ansiSGR = regexp.MustCompile("\x1b\\[[0-9;]*m")

// StripANSI removes SGR escape sequences, leaving only visible text.
func StripANSI(s string) string { return ansiSGR.ReplaceAllString(s, "") }

// Width returns the printed column width of s. ANSI SGR codes contribute
// nothing, zero-width runes count as 0, and East-Asian wide runes count as 2.
func Width(s string) int {
	s = StripANSI(s)
	w := 0
	for _, r := range s {
		w += runeWidth(r)
	}
	return w
}

func runeWidth(r rune) int {
	switch {
	case r == 0:
		return 0
	case isZeroWidth(r):
		return 0
	case isWide(r):
		return 2
	default:
		return 1
	}
}

// isZeroWidth reports whether r is a combining mark or zero-width control.
func isZeroWidth(r rune) bool {
	switch {
	case r >= 0x0300 && r <= 0x036F, // combining diacriticals
		r >= 0x1AB0 && r <= 0x1AFF,
		r >= 0x1DC0 && r <= 0x1DFF,
		r >= 0x20D0 && r <= 0x20FF, // combining marks for symbols
		r >= 0xFE20 && r <= 0xFE2F,
		r == 0x200B, r == 0x200C, r == 0x200D, r == 0xFEFF:
		return true
	}
	return false
}

// isWide reports whether r occupies two terminal cells (CJK, most emoji, …).
func isWide(r rune) bool {
	switch {
	case r >= 0x1100 && r <= 0x115F, // Hangul Jamo
		r >= 0x2E80 && r <= 0x303E, // CJK radicals, Kangxi
		r >= 0x3041 && r <= 0x33FF, // Hiragana … CJK symbols
		r >= 0x3400 && r <= 0x4DBF, // CJK Ext A
		r >= 0x4E00 && r <= 0x9FFF, // CJK Unified
		r >= 0xA000 && r <= 0xA4CF, // Yi
		r >= 0xAC00 && r <= 0xD7A3, // Hangul syllables
		r >= 0xF900 && r <= 0xFAFF, // CJK compat
		r >= 0xFE30 && r <= 0xFE4F, // CJK compat forms
		r >= 0xFF00 && r <= 0xFF60, // fullwidth forms
		r >= 0xFFE0 && r <= 0xFFE6,
		r >= 0x1F300 && r <= 0x1FAFF, // emoji & pictographs
		r >= 0x20000 && r <= 0x3FFFD: // CJK Ext B+
		return true
	}
	return false
}
