package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestStatsGolden freezes statsCmd's output or error for every mode and every
// rejection path.
func TestStatsGolden(t *testing.T) {
	dir := t.TempDir()
	write := func(name, body string) string {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
		return p
	}
	a := write("a.json", `{"1":"C","2":"I","3":"C","4":"N","9":"C"}`)
	b := write("b.json", `{"1":"C","2":"C","3":"C","4":"N","8":"I"}`)
	same := write("same.json", `{"1":"C","2":"C"}`)
	other := write("other.json", `{"x":"C"}`)
	bad := write("bad.json", `[1,2]`)
	cases := [][]string{
		{},
		{"nope"},
		{"fisher", "3", "1", "1", "3"},
		{"fisher", "1", "2"},
		{"fisher", "1", "2", "x", "4"},
		{"fisher", "1", "2", "-1", "4"},
		{"wilson", "7", "10"},
		{"wilson", "7"},
		{"wilson", "11", "10"},
		{"wilson", "0", "0"},
		{"newcombe", "8", "10", "3", "10"},
		{"newcombe", "-conf", "90", "8", "10", "3", "10"},
		{"newcombe", "-conf", "80", "8", "10", "3", "10"},
		{"newcombe", "-bogus", "8", "10", "3", "10"},
		{"newcombe", "8", "10", "3"},
		{"newcombe", "11", "10", "3", "10"},
		{"newcombe", "8", "10", "11", "10"},
		{"kappa", a, b},
		{"kappa", a},
		{"kappa", filepath.Join(dir, "absent.json"), b},
		{"kappa", a, filepath.Join(dir, "absent.json")},
		{"kappa", bad, b},
		{"kappa", a, other},
		{"kappa", same, same},
	}
	var out strings.Builder
	for _, c := range cases {
		var w bytes.Buffer
		err := statsCmd(&w, c)
		msg := "ok"
		if err != nil {
			msg = strings.ReplaceAll(err.Error(), dir, "<dir>")
		}
		args := strings.ReplaceAll(strings.Join(c, " "), dir, "<dir>")
		fmt.Fprintf(&out, "$ %s\n%s=> %s\n", args, w.String(), msg)
	}
	got := out.String()
	golden := filepath.Join("testdata", "stats.golden")
	if os.Getenv("UPDATE_GOLDEN") == "1" {
		if err := os.WriteFile(golden, []byte(got), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatal(err)
	}
	if got != string(want) {
		t.Fatalf("statsCmd drifted from %s:\n%s", golden, got)
	}
}
