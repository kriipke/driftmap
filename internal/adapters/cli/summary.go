package cli

import (
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/kriipke/driftmap/internal/tui"
	"github.com/kriipke/driftmap/pkg/differ"
)

const tagline = "next-gen YAML diff for operators"

// Palette for the dashboard, sampled to match the DriftMap brand look.
var (
	cPurple  = tui.Hex("#a06bff") // driftmap wordmark
	cText    = tui.Hex("#c9d1d9") // primary path text
	cGray    = tui.Hex("#8b929c") // labels / secondary text
	cDim     = tui.Hex("#6b7385") // separators, dashes, hints
	cCyan    = tui.Hex("#4ab8f0") // impacted sections, icons, links
	cAmber   = tui.Hex("#f0b429") // modified / "from" values
	cGreen   = tui.Hex("#4ade80") // added / "to" values
	cRed     = tui.Hex("#fb5c6c") // removed / high impact
	cBorder  = tui.Hex("#3b4556") // panel frame (clearly visible)
	cTrack   = tui.Hex("#2a303c") // empty meter segments
	cPanelBg = tui.Hex("#0d1117") // panel fill, so it reads as a card
)

var (
	stBorder = tui.New().Foreground(cBorder)
	stTrack  = tui.New().Foreground(cTrack)
	stText   = tui.New().Foreground(cText)
	stGray   = tui.New().Foreground(cGray)
	stDim    = tui.New().Foreground(cDim)
	stCyan   = tui.New().Foreground(cCyan)
	stAmber  = tui.New().Foreground(cAmber)
	stGreen  = tui.New().Foreground(cGreen)
	stRed    = tui.New().Foreground(cRed)
)

// runtimeKeywords are path fragments whose change is likely to affect a running
// workload (and therefore hint at a rollout). Used to compute deploy impact.
var runtimeKeywords = []string{
	"image", "tag", "replica", "resources", "limits", "requests",
	"env", "port", "ingress", "host", "service", "autoscaling",
	"redis", "secret", "affinity", "toleration", "nodeselector",
	"command", "args", "probe", "volume",
}

// renderSummary prints the dashboard-style overview of a two-file diff.
func renderSummary(fileA, fileB string, diffs []differ.VariableDiff, elapsed time.Duration) error {
	var modified, added, removed int
	for _, d := range diffs {
		switch d.Status {
		case "changed":
			modified++
		case "added":
			added++
		case "removed":
			removed++
		}
	}
	total := len(diffs)

	sections := impactedSections(diffs)
	runtime := runtimeImpactCount(diffs)
	impact := deployImpact(runtime, removed)

	labelA := columnLabel(fileA)
	labelB := columnLabel(fileB)

	rows, table := buildTable(diffs, labelA, labelB)

	models := buildCards(modified, added, removed, total, sections, impact, runtime)

	// The dashboard is as wide as its widest block (table or card row).
	inner := table.Width()
	if cw := cardsWidth(models); cw > inner {
		inner = cw
	}
	if inner < 100 {
		inner = 100
	}
	p := tui.Panel{Inner: inner, Border: stBorder, Background: &cPanelBg}

	var b strings.Builder
	b.WriteString(commandLine(fileA, fileB) + "\n\n")
	b.WriteString(p.Top() + "\n")
	b.WriteString(banner(p, elapsed) + "\n")
	b.WriteString(p.Sep() + "\n")
	b.WriteString(renderCards(p, models))
	b.WriteString(p.Sep() + "\n")
	b.WriteString(comparedLine(p, fileA, fileB) + "\n")
	b.WriteString(p.Sep() + "\n")
	b.WriteString(tableBlock(p, rows, table, labelA, labelB))
	b.WriteString(p.Sep() + "\n")
	b.WriteString(footer(p, modified, added, removed) + "\n")
	b.WriteString(p.Bottom() + "\n")
	b.WriteString(stCyan.Render("✧ driftmap keeps your desired state in sight.") + "\n")

	fmt.Print(b.String())
	return nil
}

// --- deploy impact ----------------------------------------------------------

type impactLevel struct {
	name string
	fill int // filled bar segments out of 5
	note string
}

func deployImpact(runtime, removed int) impactLevel {
	switch {
	case runtime >= 5 || removed >= 2:
		return impactLevel{"HIGH", 4, "Likely rollout required"}
	case runtime >= 2:
		return impactLevel{"MEDIUM", 3, "Rollout probable"}
	case runtime >= 1:
		return impactLevel{"LOW", 2, "Minimal runtime effect"}
	default:
		return impactLevel{"NONE", 1, "No runtime-impacting changes"}
	}
}

func (i impactLevel) style() tui.Style {
	switch i.name {
	case "HIGH":
		return stRed
	case "MEDIUM":
		return stAmber
	case "LOW":
		return stGreen
	default:
		return stGray
	}
}

func runtimeImpactCount(diffs []differ.VariableDiff) int {
	n := 0
	for _, d := range diffs {
		if d.Status == "added" {
			continue // additive keys rarely force a rollout on their own
		}
		lower := strings.ToLower(d.Name)
		for _, kw := range runtimeKeywords {
			if strings.Contains(lower, kw) {
				n++
				break
			}
		}
	}
	return n
}

// impactedSections returns the sorted set of top-level sections touched.
func impactedSections(diffs []differ.VariableDiff) []string {
	seen := map[string]struct{}{}
	for _, d := range diffs {
		seen[topSection(d.Name)] = struct{}{}
	}
	out := make([]string, 0, len(seen))
	for s := range seen {
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}

// topSection returns the first path segment, e.g. "resources.limits.cpu" ->
// "resources" and "ingress.hosts[0].host" -> "ingress".
func topSection(name string) string {
	if i := strings.IndexAny(name, ".["); i >= 0 {
		return name[:i]
	}
	return name
}

// --- command line -----------------------------------------------------------

func commandLine(fileA, fileB string) string {
	return stGreen.Render("❯") + " " +
		stText.Render(fmt.Sprintf("driftmap -o summary %q %q", fileA, fileB))
}

// --- banner -----------------------------------------------------------------

func banner(p tui.Panel, elapsed time.Duration) string {
	left := strings.Join([]string{
		tui.New().Foreground(cPurple).Bold().Render("driftmap"),
		"  " + stGray.Render(tagline),
	}, " ")
	clock := stCyan.Render("◷ " + formatElapsed(elapsed))
	return p.Split(left, clock)
}

func formatElapsed(d time.Duration) string {
	switch {
	case d < time.Millisecond:
		return fmt.Sprintf("%dµs", d.Microseconds())
	case d < time.Second:
		return fmt.Sprintf("%dms", d.Milliseconds())
	default:
		return fmt.Sprintf("%.2fs", d.Seconds())
	}
}

// --- stat cards -------------------------------------------------------------

// card is one dashboard stat panel: a vertical stack of pre-styled lines plus
// its natural (visible) width, so cards can be laid out variable-width like the
// reference design rather than all squeezed to an equal share.
type card struct {
	lines []string
	width int
}

// cardGap is the styled divider drawn between adjacent cards.
var cardGap = " " + stBorder.Render("│") + " "

// maxCardWidth caps a single card so one long line can't dominate the row.
const maxCardWidth = 30

func newCard(lines []string) card {
	w := 0
	for _, l := range lines {
		if x := tui.Width(l); x > w {
			w = x
		}
	}
	if w > maxCardWidth {
		w = maxCardWidth
	}
	return card{lines: lines, width: w}
}

func buildCards(modified, added, removed, total int, sections []string, impact impactLevel, runtime int) []card {
	pct := func(n int) string {
		if total == 0 {
			return "0% of changes"
		}
		return fmt.Sprintf("%d%% of changes", n*100/total)
	}

	sectionLines := tui.WrapWords(sections, ", ", 22)
	bar := tui.SegmentBar(impact.fill, 5, impact.style(), stTrack)

	return []card{
		newCard(statCardLines("⊟", stAmber, "MODIFIED KEYS", modified, stAmber, pct(modified))),
		newCard(statCardLines("⊞", stGreen, "ADDED KEYS", added, stGreen, pct(added))),
		newCard(statCardLines("⊖", stRed, "REMOVED KEYS", removed, stRed, pct(removed))),
		newCard(statCardLines("⧉", stCyan, "IMPACTED SECTIONS", len(sections), stCyan, sectionLines...)),
		newCard([]string{
			stGray.Render("▤ DEPLOY IMPACT"),
			impact.style().Bold().Render(impact.name),
			bar,
			impact.style().Render(impact.note),
			stGray.Render(fmt.Sprintf("%d runtime-impacting changes", runtime)),
		}),
	}
}

// statCardLines builds the common card body: icon+label title, a bold headline
// count, then any number of gray detail lines.
func statCardLines(icon string, iconStyle tui.Style, label string, headline int, headlineStyle tui.Style, details ...string) []string {
	lines := []string{
		iconStyle.Render(icon) + " " + stGray.Render(label),
	}
	lines = append(lines, tui.BigNumber(fmt.Sprintf("%d", headline), headlineStyle)...)
	for _, d := range details {
		lines = append(lines, stGray.Render(d))
	}
	return lines
}

// cardsWidth is the total rendered width of the card row (cards + dividers).
func cardsWidth(cards []card) int {
	total := 0
	for _, c := range cards {
		total += c.width
	}
	if len(cards) > 1 {
		total += tui.Width(cardGap) * (len(cards) - 1)
	}
	return total
}

// renderCards lays the cards edge-to-edge across the panel, distributing any
// slack (when the table is wider than the cards' natural size) evenly so the
// final card stays flush with the right border.
func renderCards(p tui.Panel, cards []card) string {
	n := len(cards)
	widths := make([]int, n)
	for i, c := range cards {
		widths[i] = c.width
	}
	for slack := p.Inner - cardsWidth(cards); slack > 0; slack-- {
		widths[slack%n]++
	}

	grid := tui.NewGrid(cardGap)
	height := 0
	for i, c := range cards {
		grid.Col(widths[i], tui.Left)
		if len(c.lines) > height {
			height = len(c.lines)
		}
	}

	var b strings.Builder
	for row := 0; row < height; row++ {
		cells := make([]string, n)
		for c := 0; c < n; c++ {
			if row < len(cards[c].lines) {
				cells[c] = cards[c].lines[row]
			}
		}
		b.WriteString(p.Line(grid.Row(cells...)) + "\n")
	}
	return b.String()
}

// --- compared line ----------------------------------------------------------

func comparedLine(p tui.Panel, fileA, fileB string) string {
	left := stDim.Render("ⓘ Compared: ") +
		stCyan.Render(filepath.Base(fileA)) +
		stDim.Render(" → ") +
		stCyan.Render(filepath.Base(fileB))
	stamp := stDim.Render("▤ " + time.Now().Format("Jan 02, 2006 15:04:05"))
	return p.Split(left, stamp)
}

// --- table ------------------------------------------------------------------

type tableRow struct {
	icon        string
	iconStyle   tui.Style
	path        string
	left        string
	leftStyle   tui.Style
	right       string
	rightStyle  tui.Style
	status      string
	statusStyle tui.Style
}

func buildTable(diffs []differ.VariableDiff, labelA, labelB string) ([]tableRow, *tui.Grid) {
	const maxVal = 46

	rows := make([]tableRow, 0, len(diffs))
	for _, d := range diffs {
		icon, iconStyle := rowIcon(d.Name)
		r := tableRow{icon: icon, iconStyle: iconStyle, path: d.Name}
		switch d.Status {
		case "added":
			r.left = "—"
			r.leftStyle = stDim
			r.right, r.rightStyle = tui.Truncate(summaryValue(d.Value), maxVal), stGreen
			r.status, r.statusStyle = "+ added", stGreen
		case "removed":
			r.left, r.leftStyle = tui.Truncate(summaryValue(d.Default), maxVal), stRed
			r.right = "—"
			r.rightStyle = stDim
			r.status, r.statusStyle = "− removed", stRed
		default:
			r.left, r.leftStyle = tui.Truncate(summaryValue(d.Default), maxVal), stAmber
			r.right, r.rightStyle = tui.Truncate(summaryValue(d.Value), maxVal), stGreen
			r.status, r.statusStyle = "↑ modified", stAmber
		}
		rows = append(rows, r)
	}

	// Column widths from the data (icon + 1 space folded into PATH width).
	wPath, wLeft, wRight, wStatus := tui.Width("PATH"), tui.Width(labelA), tui.Width(labelB), tui.Width("STATUS")
	for _, r := range rows {
		wPath = maxi(wPath, 2+tui.Width(r.path)) // icon + space + path
		wLeft = maxi(wLeft, tui.Width(r.left))
		wRight = maxi(wRight, tui.Width(r.right))
		wStatus = maxi(wStatus, tui.Width(r.status))
	}

	grid := tui.NewGrid("  ").
		Col(wPath, tui.Left).
		Col(wLeft, tui.Left).
		Col(wRight, tui.Left).
		Col(wStatus, tui.Left)
	return rows, grid
}

// summaryValue keeps multiline values and control characters inside one cell.
func summaryValue(value interface{}) string {
	text := formatValuePlain(value)
	if strings.ContainsFunc(text, unicode.IsControl) {
		return strconv.Quote(text)
	}
	return text
}

func tableBlock(p tui.Panel, rows []tableRow, grid *tui.Grid, labelA, labelB string) string {
	var b strings.Builder
	header := grid.Row(
		stGray.Render("PATH"),
		stGray.Render(labelA),
		stGray.Render(labelB),
		stGray.Render("STATUS"),
	)
	b.WriteString(p.Line(header) + "\n")
	b.WriteString(p.Line(stBorder.Render(strings.Repeat("─", grid.Width()))) + "\n")

	for _, r := range rows {
		path := r.iconStyle.Render(r.icon) + " " + stText.Render(r.path)
		b.WriteString(p.Line(grid.Row(
			path,
			r.leftStyle.Render(r.left),
			r.rightStyle.Render(r.right),
			r.statusStyle.Render(r.status),
		)) + "\n")
	}
	return b.String()
}

// rowIcon returns a glyph and color hinting at the kind of setting a path
// refers to (network, storage, secret, …). All glyphs are single-width.
func rowIcon(path string) (string, tui.Style) {
	p := strings.ToLower(path)
	switch {
	case strings.Contains(p, "image"), strings.Contains(p, ".tag"):
		return "▣", stCyan
	case strings.HasPrefix(p, "app"), strings.HasPrefix(p, "replica"):
		return "◈", stCyan
	case strings.HasPrefix(p, "service.annotations"):
		return "◍", stCyan
	case strings.HasPrefix(p, "service"):
		return "⤢", stCyan
	case strings.HasPrefix(p, "ingress"):
		return "◉", stCyan
	case strings.HasPrefix(p, "autoscaling"):
		return "▤", stCyan
	case strings.HasPrefix(p, "resources"):
		return "▦", stCyan
	case strings.HasPrefix(p, "env"):
		return "⌗", stCyan
	case strings.HasPrefix(p, "features"), strings.Contains(p, "flag"):
		return "⚑", stAmber
	case strings.HasPrefix(p, "secret"):
		return "⚿", stGray
	case strings.HasPrefix(p, "redis"):
		return "◆", stCyan
	case strings.HasPrefix(p, "nodeselector"):
		return "✛", stGreen
	case strings.HasPrefix(p, "toleration"):
		return "▨", stRed
	case strings.HasPrefix(p, "affinity"):
		return "⬡", stRed
	default:
		return "◦", stCyan
	}
}

// columnLabel derives an uppercased column header from a file path, dropping
// the directory and extension: "…/values/claude.yaml" -> "CLAUDE".
func columnLabel(path string) string {
	base := filepath.Base(path)
	base = strings.TrimSuffix(base, filepath.Ext(base))
	return strings.ToUpper(base)
}

// --- footer -----------------------------------------------------------------

func footer(p tui.Panel, modified, added, removed int) string {
	legend := strings.Join([]string{
		stAmber.Render(fmt.Sprintf("↑ modified (%d)", modified)),
		stGreen.Render(fmt.Sprintf("+ added (%d)", added)),
		stRed.Render(fmt.Sprintf("− removed (%d)", removed)),
	}, "   ")
	tip := stDim.Render("ⓘ Tip: use -o shell for full values")
	return p.Split(legend, tip)
}

func maxi(a, b int) int {
	if a > b {
		return a
	}
	return b
}
