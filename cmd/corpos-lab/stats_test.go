package main

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

// Golden values are the ones the published papers print, so the subcommand is
// checked against the record, not against itself.
func TestStatsMatchesPublishedValues(t *testing.T) {
	cases := []struct {
		args []string
		want []string
	}{
		// behavioral-equivalence: duty 23/24 vs corpus 4/24, two-sided Fisher p=2.3e-8.
		{[]string{"fisher", "23", "1", "4", "20"}, []string{"p=2.3e-08"}},
		// register-shift-followup: imperative 5/24 vs domain directive 24/24, printed p≈7e-9.
		{[]string{"fisher", "5", "19", "24", "0"}, []string{"p=7.4e-09"}},
		// register-shift-followup: Newcombe 95% on 24/24 minus 5/24 = +0.79 [+0.55, +0.91].
		{[]string{"newcombe", "24", "24", "5", "24"}, []string{"diff=+0.79", "[+0.55, +0.91]"}},
		// behavioral-equivalence: 90% interval on corpus 4/24 minus duty 23/24 = [-0.89, -0.59].
		{[]string{"newcombe", "-conf", "90", "4", "24", "23", "24"}, []string{"[-0.89, -0.59]"}},
		// q2-register-shift: Wilson 4/8 spans 0.22 to 0.78, 7/8 0.53 to 0.98.
		{[]string{"wilson", "4", "8"}, []string{"[0.22, 0.78]"}},
		{[]string{"wilson", "7", "8"}, []string{"[0.53, 0.98]"}},
		// behavioral-equivalence Table: 23/24 [0.80, 0.99], 4/24 [0.07, 0.36].
		{[]string{"wilson", "23", "24"}, []string{"[0.80, 0.99]"}},
		{[]string{"wilson", "4", "24"}, []string{"[0.07, 0.36]"}},
	}
	for _, c := range cases {
		var out bytes.Buffer
		if err := statsCmd(&out, c.args); err != nil {
			t.Fatalf("stats %v: %v", c.args, err)
		}
		for _, w := range c.want {
			if !strings.Contains(out.String(), w) {
				t.Errorf("stats %v = %q, want it to contain %q", c.args, out.String(), w)
			}
		}
	}
}

func TestStatsKappa(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.json")
	b := filepath.Join(dir, "b.json")
	writeFileT(t, a, `{"1": "C", "2": "C", "3": "I", "4": "I", "5": "C"}`)
	writeFileT(t, b, `{"1": "C", "2": "I", "3": "I", "4": "I", "6": "C"}`)
	var out bytes.Buffer
	if err := statsCmd(&out, []string{"kappa", a, b}); err != nil {
		t.Fatal(err)
	}
	// 4 shared ids, 3 agree: po=0.75, pe=0.5 -> kappa=0.50; ids 5 and 6 unpaired.
	for _, w := range []string{"n=4", "agree=0.75", "kappa=0.50", "unpaired=2"} {
		if !strings.Contains(out.String(), w) {
			t.Errorf("kappa = %q, want %q", out.String(), w)
		}
	}
}

func TestStatsErrors(t *testing.T) {
	for _, args := range [][]string{
		{}, {"bogus"}, {"fisher", "1", "2"}, {"fisher", "1", "2", "3", "x"}, {"fisher", "-1", "2", "3", "4"},
		{"wilson", "5", "0"}, {"wilson", "6", "5"}, {"newcombe", "-conf", "80", "1", "2", "1", "2"},
		{"kappa", "nope.json"}, {"kappa", "a.json", "b.json"},
	} {
		if err := statsCmd(&bytes.Buffer{}, args); err == nil {
			t.Errorf("stats %v: want an error", args)
		}
	}
	if code := run([]string{"stats", "wilson", "1", "2"}); code != 0 {
		t.Errorf("run stats wilson = %d", code)
	}
	if code := run([]string{"stats"}); code != 2 {
		t.Errorf("run stats (no mode) = %d, want 2", code)
	}
}
