package study

import (
	"strings"
	"testing"
)

func TestScenarioSelfContainmentWarnings(t *testing.T) {
	cases := []struct {
		name    string
		text    string
		wantN   int
		wantSub string
	}{
		{
			name:  "self-contained scenario is clean",
			text:  "Rename the file report.go to summary.go. Do only that.",
			wantN: 0,
		},
		{
			name:    "references an absent doc by name",
			text:    "Follow the steps in GUIDE.md to complete the migration.",
			wantN:   1,
			wantSub: "guide.md",
		},
		{
			name:    "verbal pointer to an external document",
			text:    "Complete the task as described in the document.",
			wantN:   1,
			wantSub: "points outside the prompt",
		},
		{
			name:  "dedupes repeated references to the same doc",
			text:  "Read GUIDE.md. The GUIDE.md has the rules. Open GUIDE.md again.",
			wantN: 1,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := ScenarioSelfContainmentWarnings(c.text)
			if len(got) != c.wantN {
				t.Fatalf("got %d warnings %v, want %d", len(got), got, c.wantN)
			}
			if c.wantSub != "" {
				joined := strings.ToLower(strings.Join(got, " | "))
				if !strings.Contains(joined, c.wantSub) {
					t.Errorf("warnings %v do not mention %q", got, c.wantSub)
				}
			}
		})
	}
}

func TestScenarioSelfContainmentStableOrder(t *testing.T) {
	text := "See the appendix. Also read SPEC.md and NOTES.txt for details."
	a := ScenarioSelfContainmentWarnings(text)
	b := ScenarioSelfContainmentWarnings(text)
	if len(a) == 0 {
		t.Fatal("expected warnings for a scenario with multiple dangling refs")
	}
	if strings.Join(a, "\n") != strings.Join(b, "\n") {
		t.Errorf("warning order is not stable: %v vs %v", a, b)
	}
}
