package study

import (
	"fmt"
	"strconv"
	"strings"

	"corpos-lab/internal/image"
)

// Overrides are run-time changes to a study definition, given on the run-study
// command line so a smoke check or a re-run of a few seeds needs no TOML copy.
// Copying a TOML to /tmp broke its relative material paths, and every image
// change meant re-pinning near-identical TOMLs. The run record names each
// override that was applied (record what ran).
type Overrides struct {
	// Seeds replaces the seed list. When RunsPerCell is unset it also sets
	// runs_per_cell to len(Seeds), so run N uses Seeds[N-1] as always.
	Seeds []int
	// RunsPerCell replaces runs_per_cell. With no Seeds override, the declared
	// seed list is cut to its first RunsPerCell entries, so each run keeps the
	// seed it had.
	RunsPerCell int
	// Image replaces the pinned image. It must be digest-pinned.
	Image string
}

// Empty reports whether no override is set.
func (o Overrides) Empty() bool {
	return len(o.Seeds) == 0 && o.RunsPerCell == 0 && o.Image == ""
}

// Record renders the overrides for the run record, nil when none is set.
func (o Overrides) Record() map[string]string {
	if o.Empty() {
		return nil
	}
	m := map[string]string{}
	if len(o.Seeds) > 0 {
		parts := make([]string, len(o.Seeds))
		for i, s := range o.Seeds {
			parts[i] = strconv.Itoa(s)
		}
		m["seeds"] = strings.Join(parts, ",")
	}
	if o.RunsPerCell > 0 {
		m["runs_per_cell"] = strconv.Itoa(o.RunsPerCell)
	}
	if o.Image != "" {
		m["image"] = o.Image
	}
	return m
}

// WithOverrides returns a copy of d with o applied, validated as a whole
// definition. d itself is unchanged.
func (d Def) WithOverrides(o Overrides) (Def, error) {
	if o.Image != "" {
		if _, _, ok := image.ParsePinnedDigest(o.Image); !ok {
			return Def{}, fmt.Errorf("study: image override %q is not digest-pinned (want <repo>@sha256:<digest>)", o.Image)
		}
		d.Image = o.Image
	}
	switch {
	case len(o.Seeds) > 0:
		if o.RunsPerCell > 0 && o.RunsPerCell != len(o.Seeds) {
			return Def{}, fmt.Errorf("study: %d seeds but runs_per_cell override %d — give one seed per run", len(o.Seeds), o.RunsPerCell)
		}
		d.Sampling.Seeds = append([]int(nil), o.Seeds...)
		d.RunsPerCell = len(o.Seeds)
	case o.RunsPerCell > 0:
		if len(d.Sampling.Seeds) > 0 {
			if o.RunsPerCell > len(d.Sampling.Seeds) {
				return Def{}, fmt.Errorf("study: runs_per_cell override %d exceeds the %d declared seeds — add -seeds", o.RunsPerCell, len(d.Sampling.Seeds))
			}
			d.Sampling.Seeds = append([]int(nil), d.Sampling.Seeds[:o.RunsPerCell]...)
		}
		d.RunsPerCell = o.RunsPerCell
	}
	if err := d.validate(); err != nil {
		return Def{}, err
	}
	d.Applied = o.Record()
	return d, nil
}

// ParseSeeds reads a comma-separated seed list such as "2,3,4".
func ParseSeeds(s string) ([]int, error) {
	var seeds []int
	for _, part := range strings.Split(s, ",") {
		n, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil || n < 0 {
			return nil, fmt.Errorf("study: bad seed %q in %q (want non-negative integers, comma-separated)", part, s)
		}
		seeds = append(seeds, n)
	}
	return seeds, nil
}
