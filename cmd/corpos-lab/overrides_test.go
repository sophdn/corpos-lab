package main

import (
	"path/filepath"
	"testing"

	"corpos-lab/internal/study"
)

func TestSetOverride(t *testing.T) {
	var o study.Overrides
	for _, kv := range [][2]string{{"-seeds", "2,3"}, {"-runs", "2"}, {"-image", "x@sha256:y"}} {
		if err := setOverride(&o, kv[0], kv[1]); err != nil {
			t.Fatalf("%s %s: %v", kv[0], kv[1], err)
		}
	}
	if len(o.Seeds) != 2 || o.RunsPerCell != 2 || o.Image != "x@sha256:y" {
		t.Errorf("overrides = %+v", o)
	}
	for _, kv := range [][2]string{{"-seeds", "a"}, {"-runs", "0"}} {
		if err := setOverride(&o, kv[0], kv[1]); err == nil {
			t.Errorf("%s %s: want an error", kv[0], kv[1])
		}
	}
}

func TestRunStudyOverrideFlagErrors(t *testing.T) {
	for _, args := range [][]string{{"run-study", "x.toml", "-seeds"}, {"run-study", "x.toml", "-runs", "zero"}} {
		if code := run(args); code != 2 {
			t.Errorf("run(%v) = %d, want 2", args, code)
		}
	}
}

func TestShouldPersist(t *testing.T) {
	over := map[string]string{"runs_per_cell": "1"}
	cases := []struct {
		url     string
		applied map[string]string
		force   bool
		want    bool
		why     bool
	}{
		{"http://t", nil, false, true, false},
		{"http://t", over, false, false, true},
		{"http://t", over, true, true, false},
		{"", nil, true, false, false},
	}
	for _, c := range cases {
		got, why := shouldPersist(c.url, c.applied, c.force)
		if got != c.want || (why != "") != c.why {
			t.Errorf("shouldPersist(%q, %v, %v) = %v, %q", c.url, c.applied, c.force, got, why)
		}
	}
}

// TestRunStudyArgExitCodes pins run-study's argument handling, all of which runs
// before any container work: 2 for a flag missing its value, a bad override, or
// the wrong number of positional args; 1 when the definition cannot load (every
// flag accepted, in any position).
func TestRunStudyArgExitCodes(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "absent.toml")
	cases := []struct {
		args []string
		want int
	}{
		{[]string{"run-study"}, 2},
		{[]string{"run-study", "a.toml", "b.toml"}, 2},
		{[]string{"run-study", "a.toml", "-work"}, 2},
		{[]string{"run-study", "a.toml", "-toolkit-url"}, 2},
		{[]string{"run-study", "a.toml", "-project"}, 2},
		{[]string{"run-study", "a.toml", "-image"}, 2},
		{[]string{"run-study", "a.toml", "-runs", "0"}, 2},
		{[]string{"run-study", "a.toml", "-seeds", "x"}, 2},
		{[]string{"run-study", missing}, 1},
		{[]string{"run-study", "-work", "w", "-toolkit-url", "", "-project", "p", "-persist",
			"-runs", "2", "-seeds", "1,2", "-image", "img@sha256:abc", missing}, 1},
	}
	for _, c := range cases {
		if got := run(c.args); got != c.want {
			t.Errorf("run(%v) = %d, want %d", c.args, got, c.want)
		}
	}
}

// TestParseRunStudyArgs pins the parsed options: defaults, every flag in any
// position, and an empty -toolkit-url kept as empty (persistence off).
func TestParseRunStudyArgs(t *testing.T) {
	opts, msg := parseRunStudyArgs([]string{"d.toml"})
	if msg != "" || opts.defPath != "d.toml" || opts.workDir != "" || opts.project != "glyph-research" ||
		opts.toolkitURL == "" || opts.forcePersist || !opts.overrides.Empty() {
		t.Fatalf("defaults: %+v %q", opts, msg)
	}
	opts, msg = parseRunStudyArgs([]string{"-persist", "-toolkit-url", "", "-runs", "3", "d.toml", "-work", "w", "-project", "p", "-seeds", "4,5", "-image", "i@sha256:x"})
	if msg != "" || opts.defPath != "d.toml" || opts.workDir != "w" || opts.project != "p" || opts.toolkitURL != "" ||
		!opts.forcePersist || opts.overrides.RunsPerCell != 3 || len(opts.overrides.Seeds) != 2 || opts.overrides.Image != "i@sha256:x" {
		t.Fatalf("all flags: %+v %q", opts, msg)
	}
	if _, msg := parseRunStudyArgs([]string{"d.toml", "-seeds"}); msg != "corpos-lab: -seeds needs a value" {
		t.Fatalf("missing value msg = %q", msg)
	}
}
