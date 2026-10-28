package tui

import "testing"

func TestWidthIgnoresANSI(t *testing.T) {
	styled := New().Foreground(RGB(255, 0, 0)).Bold().Render("hello")
	if got := Width(styled); got != 5 {
		t.Errorf("Width(styled %q) = %d, want 5", styled, got)
	}
}

func TestWidthWideRunes(t *testing.T) {
	// A wide CJK glyph counts as two columns; box-drawing counts as one.
	if got := Width("世界"); got != 4 {
		t.Errorf("Width(CJK) = %d, want 4", got)
	}
	if got := Width("╭──╮"); got != 4 {
		t.Errorf("Width(box) = %d, want 4", got)
	}
}

func TestPadRightWidthAware(t *testing.T) {
	SetColorEnabled(true)
	defer SetColorEnabled(false)
	styled := New().Foreground(RGB(0, 255, 0)).Render("ok")
	got := PadRight(styled, 5)
	if Width(got) != 5 {
		t.Errorf("PadRight visible width = %d, want 5", Width(got))
	}
	// The color codes must survive padding.
	if StripANSI(got) != "ok   " {
		t.Errorf("PadRight plain = %q, want %q", StripANSI(got), "ok   ")
	}
}

func TestTruncate(t *testing.T) {
	cases := []struct {
		in   string
		w    int
		want string
	}{
		{"hello world", 5, "hell…"},
		{"hi", 5, "hi"},
		{"abc", 1, "…"},
		{"abc", 0, ""},
	}
	for _, c := range cases {
		if got := Truncate(c.in, c.w); got != c.want {
			t.Errorf("Truncate(%q, %d) = %q, want %q", c.in, c.w, got, c.want)
		}
	}
}

func TestWrapWords(t *testing.T) {
	got := WrapWords([]string{"aaa", "bbb", "ccc", "ddd"}, ", ", 10)
	want := []string{"aaa, bbb", "ccc, ddd"}
	if len(got) != len(want) {
		t.Fatalf("WrapWords lines = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("line %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestSegmentBar(t *testing.T) {
	SetColorEnabled(false)
	bar := SegmentBar(4, 5, New(), New())
	if bar != "■■■■□" {
		t.Errorf("SegmentBar = %q, want ■■■■□", bar)
	}
	// Overfill is clamped to total.
	if got := SegmentBar(9, 3, New(), New()); got != "■■■" {
		t.Errorf("SegmentBar overfill = %q, want ■■■", got)
	}
}

func TestGridRowWidth(t *testing.T) {
	g := NewGrid("  ").Col(6, Left).Col(4, Right)
	row := g.Row("ab", "9")
	if Width(row) != 6+2+4 {
		t.Errorf("grid row width = %d, want %d", Width(row), 12)
	}
	if StripANSI(row) != "ab         9" {
		t.Errorf("grid row = %q", StripANSI(row))
	}
}

func TestPanelLineWidth(t *testing.T) {
	SetColorEnabled(false)
	p := Panel{Inner: 10, Border: New()}
	line := p.Line("hi")
	// "│ " + 10 cols + " │" = 14 columns.
	if Width(line) != 14 {
		t.Errorf("panel line width = %d, want 14", Width(line))
	}
}
