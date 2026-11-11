package cartographer

import "strconv"

// BlindMapEntry is one blind_map.json row: the opaque id and the condition/run it
// unblinds to. The caller passes these in blind_map order so the missing-id list
// keeps that order.
type BlindMapEntry struct {
	ID        string
	Condition string
	Run       int
}

// CoverageRow is one unblinded coverage-grid row, keyed downstream by
// "<condition>_<run>". The coverage map carries one 0/1 call per registered slug.
type CoverageRow struct {
	Condition string         `json:"condition"`
	Run       int            `json:"run"`
	Coverage  map[string]int `json:"coverage"`
}

// Unblind attaches the condition to each judged blind id, producing the coverage
// grid analyze consumes. judge maps a blind id to its per-slug 0/1 calls; a slug
// the judge omitted counts as 0, and any non-zero call is normalised to 1
// (python's int(bool(...))). An id absent from judge is skipped and reported in
// missing, in blind_map order. Ported from unblind.main; the returned grid is
// keyed "<condition>_<run>".
func Unblind(mapping []BlindMapEntry, judge map[string]map[string]int) (grid map[string]CoverageRow, missing []string) {
	grid = make(map[string]CoverageRow, len(mapping))
	for _, m := range mapping {
		calls, ok := judge[m.ID]
		if !ok {
			missing = append(missing, m.ID)
			continue
		}
		cov := make(map[string]int, len(C2Slugs))
		for _, s := range C2Slugs {
			if calls[s] != 0 {
				cov[s] = 1
			} else {
				cov[s] = 0
			}
		}
		key := gridKey(m.Condition, m.Run)
		grid[key] = CoverageRow{Condition: m.Condition, Run: m.Run, Coverage: cov}
	}
	return grid, missing
}

// gridKey builds the "<condition>_<run>" key used across the grid outputs.
func gridKey(cond string, run int) string {
	return cond + "_" + strconv.Itoa(run)
}
