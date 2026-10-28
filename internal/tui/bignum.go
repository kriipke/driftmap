package tui

import "strings"

// bigDigitHeight is the number of terminal rows each block glyph occupies.
const bigDigitHeight = 5

// bigDigits is a 3-column, 5-row block font. Terminals can't scale a font per
// glyph, so "big" numbers are drawn as multi-row art: each rune maps to
// bigDigitHeight strings that stack vertically. Digits are 3 cells wide.
var bigDigits = map[rune][bigDigitHeight]string{
	'0': {"███", "█ █", "█ █", "█ █", "███"},
	'1': {" █ ", "██ ", " █ ", " █ ", "███"},
	'2': {"███", "  █", "███", "█  ", "███"},
	'3': {"███", "  █", "███", "  █", "███"},
	'4': {"█ █", "█ █", "███", "  █", "  █"},
	'5': {"███", "█  ", "███", "  █", "███"},
	'6': {"███", "█  ", "███", "█ █", "███"},
	'7': {"███", "  █", "  █", "  █", "  █"},
	'8': {"███", "█ █", "███", "█ █", "███"},
	'9': {"███", "█ █", "███", "  █", "███"},
	'%': {"█ █", "  █", " █ ", "█  ", "█ █"},
	'-': {"   ", "   ", "███", "   ", "   "},
}

// BigNumber renders s (digits and the handful of symbols above) as a slice of
// bigDigitHeight styled rows, ready to drop into a stacked card body. Unknown
// runes render as blank 3-cell columns. Each row is styled as a whole so ANSI
// codes stay balanced.
func BigNumber(s string, style Style) []string {
	rows := make([]string, bigDigitHeight)
	for r := 0; r < bigDigitHeight; r++ {
		var parts []string
		for _, ch := range s {
			glyph, ok := bigDigits[ch]
			if !ok {
				parts = append(parts, "   ")
				continue
			}
			parts = append(parts, glyph[r])
		}
		rows[r] = style.Render(strings.Join(parts, " "))
	}
	return rows
}
