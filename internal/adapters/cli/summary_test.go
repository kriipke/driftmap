package cli

import (
	"reflect"
	"strings"
	"testing"

	"github.com/kriipke/driftmap/pkg/differ"
)

func TestSummaryTableEscapesControlCharacters(t *testing.T) {
	rows, grid := buildTable([]differ.VariableDiff{
		{Name: "message", Default: "first\nsecond", Value: "\x1b[31mred\ttext", Status: "changed"},
	}, "BEFORE", "AFTER")
	if len(rows) != 1 {
		t.Fatalf("expected one row, got %d", len(rows))
	}
	row := rows[0]
	if row.left != `"first\nsecond"` || row.right != `"\x1b[31mred\ttext"` {
		t.Fatalf("unexpected cell values: %q, %q", row.left, row.right)
	}
	if strings.ContainsAny(grid.Row(row.path, row.left, row.right, row.status), "\n\t\x1b") {
		t.Fatal("table row contains unescaped control characters")
	}
}

func TestColumnLabel(t *testing.T) {
	cases := map[string]string{
		"claude.yaml":                        "CLAUDE",
		"~/infra/k8s/values/production.yaml": "PRODUCTION",
		"values-dev.yml":                     "VALUES-DEV",
		"noext":                              "NOEXT",
		"/abs/path/values-before.yaml":       "VALUES-BEFORE",
	}
	for in, want := range cases {
		if got := columnLabel(in); got != want {
			t.Errorf("columnLabel(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestTopSection(t *testing.T) {
	cases := map[string]string{
		"resources.limits.cpu":  "resources",
		"ingress.hosts[0].host": "ingress",
		"replicaCount":          "replicaCount",
		"tolerations[0]":        "tolerations",
	}
	for in, want := range cases {
		if got := topSection(in); got != want {
			t.Errorf("topSection(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestImpactedSections(t *testing.T) {
	diffs := []differ.VariableDiff{
		{Name: "nginx.enabled", Status: "changed"},
		{Name: "nginx.image.tag", Status: "changed"},
		{Name: "replicaCount", Status: "changed"},
		{Name: "featureFlags.enableBeta", Status: "changed"},
	}
	got := impactedSections(diffs)
	want := []string{"featureFlags", "nginx", "replicaCount"} // deduped + sorted
	if !reflect.DeepEqual(got, want) {
		t.Errorf("impactedSections = %v, want %v", got, want)
	}
}

func TestRuntimeImpactCount(t *testing.T) {
	diffs := []differ.VariableDiff{
		{Name: "nginx.image.tag", Status: "changed"},         // runtime (image/tag)
		{Name: "resources.limits.cpu", Status: "changed"},    // runtime (resources)
		{Name: "featureFlags.enableBeta", Status: "changed"}, // not runtime
		{Name: "autoscaling.minReplicas", Status: "added"},   // added -> ignored
		{Name: "nginx.ports[1]", Status: "removed"},          // runtime (port)
	}
	if got, want := runtimeImpactCount(diffs), 3; got != want {
		t.Errorf("runtimeImpactCount = %d, want %d", got, want)
	}
}

func TestDeployImpact(t *testing.T) {
	cases := []struct {
		runtime, removed int
		wantName         string
		wantFill         int
	}{
		{8, 2, "HIGH", 4},
		{0, 2, "HIGH", 4}, // two removals alone escalate to HIGH
		{3, 0, "MEDIUM", 3},
		{1, 0, "LOW", 2},
		{0, 0, "NONE", 1},
	}
	for _, c := range cases {
		got := deployImpact(c.runtime, c.removed)
		if got.name != c.wantName || got.fill != c.wantFill {
			t.Errorf("deployImpact(%d,%d) = {%s,%d}, want {%s,%d}",
				c.runtime, c.removed, got.name, got.fill, c.wantName, c.wantFill)
		}
	}
}
