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
	checkExitCodes(t, []exitCase{
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
	})
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

// TestParseRunStudyArgsEqualsForm pins the -flag=value form run-study gained
// from the shared parser: every valued flag accepts it, mixed with the spaced
// form, an empty value still disables persistence, a value may itself contain
// '=', and a bad override is still a usage error.
func TestParseRunStudyArgsEqualsForm(t *testing.T) {
	opts, msg := parseRunStudyArgs([]string{"-persist", "-toolkit-url=", "-runs=3", "d.toml", "-work=w", "-project", "p",
		"-seeds=4,5", "-image=i@sha256:x=y"})
	if msg != "" || opts.defPath != "d.toml" || opts.workDir != "w" || opts.project != "p" || opts.toolkitURL != "" ||
		!opts.forcePersist || opts.overrides.RunsPerCell != 3 || len(opts.overrides.Seeds) != 2 ||
		opts.overrides.Image != "i@sha256:x=y" {
		t.Fatalf("equals form: %+v %q", opts, msg)
	}
	if _, msg := parseRunStudyArgs([]string{"d.toml", "-runs=0"}); msg == "" {
		t.Fatal("-runs=0 should be a usage error")
	}
	// An unknown -x=y is not a flag of ours; it stays a positional and so
	// trips the one-definition usage check, as it always did.
	if _, msg := parseRunStudyArgs([]string{"d.toml", "-nope=1"}); msg == "" {
		t.Fatal("unknown -nope=1 should be a usage error")
	}
}

// TestSplitValueFlagsEqualsForm pins the shared parser's two value forms.
func TestSplitValueFlagsEqualsForm(t *testing.T) {
	var a, b string
	var on bool
	rest, ok := splitValueFlags([]string{"x", "-a=1=2", "-b", "2", "-on", "-z=9"},
		map[string]*string{"-a": &a, "-b": &b}, map[string]*bool{"-on": &on})
	if !ok || a != "1=2" || b != "2" || !on || len(rest) != 2 || rest[0] != "x" || rest[1] != "-z=9" {
		t.Fatalf("rest=%v a=%q b=%q on=%v ok=%v", rest, a, b, on, ok)
	}
}
