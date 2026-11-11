package cartographer

import (
	"fmt"
	"math/rand"
)

// Duty is one produced response ready to be blinded: its condition, run index,
// and text. The caller reads these from a run's out/responses/ tree.
type Duty struct {
	Condition string
	Run       int
	Text      string
}

// BlindResult is one blinded duty: an opaque id, the verbatim text to write under
// that id, and the (condition, run) the id maps back to.
type BlindResult struct {
	ID        string
	Text      string
	Condition string
	Run       int
}

// BuildBlindSet assigns each duty an opaque, shuffled id (d0001, d0002, …) and
// returns the id->text payloads plus the id->(condition, run) mapping, the two
// halves build_blind_set.py writes as scores/blind/<id>.md and
// scores/blind_map.json. The shuffle uses a seeded Go RNG (Fisher-Yates), so the
// id assignment is deterministic for a seed but is NOT the same permutation
// python's Mersenne Twister produces; blinding correctness does not depend on the
// permutation, only on the mapping round-tripping, which it does for any order.
//
// The input is taken in the caller's sorted-filename order (matching python's
// sorted(os.listdir)); BuildBlindSet copies before shuffling so it does not
// mutate the caller's slice.
func BuildBlindSet(duties []Duty, seed int64) []BlindResult {
	items := make([]Duty, len(duties))
	copy(items, duties)
	rng := rand.New(rand.NewSource(seed)) //nolint:gosec // deterministic parity RNG, not security
	rng.Shuffle(len(items), func(i, j int) { items[i], items[j] = items[j], items[i] })
	out := make([]BlindResult, len(items))
	for i, d := range items {
		out[i] = BlindResult{
			ID:        fmt.Sprintf("d%04d", i+1),
			Text:      d.Text,
			Condition: d.Condition,
			Run:       d.Run,
		}
	}
	return out
}
