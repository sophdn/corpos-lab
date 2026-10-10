package main

import (
	"path/filepath"
	"testing"
)

func TestRunProvenance(t *testing.T) {
	dir := t.TempDir()
	writeFileT(t, filepath.Join(dir, "runs", "s1", "run-record.json"), `{"image_digest": "sha256:a", "results": {"model_id": "m", "rows": [], "sampler": {}, "server": {}}}`)
	for _, args := range [][]string{{"provenance", dir}, {"provenance", dir, "-json"}, {"provenance", "-json", dir}} {
		if code := run(args); code != 0 {
			t.Errorf("run(%v) = %d", args, code)
		}
	}
	for _, args := range [][]string{{"provenance"}, {"provenance", t.TempDir()}, {"provenance", dir, "-nope"}} {
		if code := run(args); code == 0 {
			t.Errorf("run(%v) = 0, want non-zero", args)
		}
	}
}

// TestDirReportExitCodes pins the exact exit codes cells and provenance share:
// 2 for a usage or flag error, 1 when the directory cannot be loaded.
func TestDirReportExitCodes(t *testing.T) {
	for _, cmd := range []string{"cells", "provenance"} {
		empty := t.TempDir()
		cases := []struct {
			args []string
			want int
		}{
			{[]string{cmd}, 2},
			{[]string{cmd, "-json"}, 2},
			// A leading dir is taken first; a stray extra argument is then ignored.
			{[]string{cmd, "a", "b"}, 1},
			{[]string{cmd, "-json", "a", "b"}, 2},
			{[]string{cmd, empty, "-nope"}, 2},
			{[]string{cmd, empty}, 1},
			{[]string{cmd, "-json", empty}, 1},
		}
		for _, c := range cases {
			if got := run(c.args); got != c.want {
				t.Errorf("run(%v) = %d, want %d", c.args, got, c.want)
			}
		}
	}
}
