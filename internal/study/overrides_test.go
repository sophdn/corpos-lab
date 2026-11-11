package study

import (
	"strings"
	"testing"
)

func TestWithOverridesSeedsSetsRuns(t *testing.T) {
	d, err := LoadDef(writeDef(t, validDef, allMaterials()))
	if err != nil {
		t.Fatal(err)
	}
	got, err := d.WithOverrides(Overrides{Seeds: []int{2, 3, 4}})
	if err != nil {
		t.Fatalf("WithOverrides: %v", err)
	}
	if len(got.Sampling.Seeds) != 3 || got.Sampling.Seeds[0] != 2 || got.RunsPerCell != 3 {
		t.Errorf("seeds/runs = %v/%d, want [2 3 4]/3", got.Sampling.Seeds, got.RunsPerCell)
	}
	if len(d.Sampling.Seeds) != 2 {
		t.Errorf("original def changed: %v", d.Sampling.Seeds)
	}
	if got.Applied["seeds"] != "2,3,4" || d.Applied != nil {
		t.Errorf("applied = %v (original %v)", got.Applied, d.Applied)
	}
}

func TestWithOverridesRunsAndImage(t *testing.T) {
	d, err := LoadDef(writeDef(t, validDef, allMaterials()))
	if err != nil {
		t.Fatal(err)
	}
	img := "localhost/lab-grounded-glyph-probe@sha256:" + strings.Repeat("c", 64)
	got, err := d.WithOverrides(Overrides{RunsPerCell: 1, Image: img})
	if err != nil {
		t.Fatalf("WithOverrides: %v", err)
	}
	// One run with two declared seeds keeps the first seed, so run 1 is the run
	// it would have been.
	if got.RunsPerCell != 1 || len(got.Sampling.Seeds) != 1 || got.Sampling.Seeds[0] != 1 || got.Image != img {
		t.Errorf("got runs %d seeds %v image %s", got.RunsPerCell, got.Sampling.Seeds, got.Image)
	}
}

func TestWithOverridesRejects(t *testing.T) {
	d, err := LoadDef(writeDef(t, validDef, allMaterials()))
	if err != nil {
		t.Fatal(err)
	}
	for name, o := range map[string]Overrides{
		"more runs than seeds":    {RunsPerCell: 5},
		"runs and seeds disagree": {Seeds: []int{1, 2}, RunsPerCell: 3},
		"tag image":               {Image: "localhost/lab-grounded-glyph-probe:dev"},
	} {
		if _, err := d.WithOverrides(o); err == nil {
			t.Errorf("%s: want an error", name)
		}
	}
}

func TestOverridesRecord(t *testing.T) {
	if got := (Overrides{}).Record(); got != nil {
		t.Errorf("empty overrides record = %v, want nil", got)
	}
	got := Overrides{Seeds: []int{2, 3}, RunsPerCell: 2, Image: "img"}.Record()
	if got["seeds"] != "2,3" || got["runs_per_cell"] != "2" || got["image"] != "img" {
		t.Errorf("record = %v", got)
	}
}

func TestParseSeeds(t *testing.T) {
	got, err := ParseSeeds("2, 3,4")
	if err != nil || len(got) != 3 || got[2] != 4 {
		t.Errorf("ParseSeeds = %v, %v", got, err)
	}
	for _, bad := range []string{"", "a", "1,,2", "1,-2"} {
		if _, err := ParseSeeds(bad); err == nil {
			t.Errorf("ParseSeeds(%q): want an error", bad)
		}
	}
}
