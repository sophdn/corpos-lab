package lengthdistraction

import (
	"sort"

	"corpos-lab/internal/neutralprefix"
)

// UnreportedConditions returns the conditions present in the key but absent from
// the reported condition list conds, sorted so a caller naming the dropped
// conditions gets a deterministic message. An empty result means the report
// covers every condition the key carries; a non-empty result means Report would
// silently drop those conditions' rows from the table (bug 1369).
func UnreportedConditions(key *neutralprefix.KeyFile, conds []string) []string {
	reported := map[string]bool{}
	for _, c := range conds {
		reported[c] = true
	}
	seen := map[string]bool{}
	var out []string
	for _, id := range key.Order {
		c := key.Meta[id].Condition
		if !reported[c] && !seen[c] {
			seen[c] = true
			out = append(out, c)
		}
	}
	sort.Strings(out)
	return out
}
