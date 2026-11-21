// Package blindslice holds the pieces the build_slices.py ports share: the
// opaque response id and the fixed-size chunking of a shuffled class.
package blindslice

import (
	"crypto/sha1" //nolint:gosec // the committed opaque-id scheme, not security
	"encoding/hex"
	"fmt"
)

// Oid is build_slices.py's opaque response id: "r" + the first 12 hex chars of
// sha1("<cls>|<sc>|<model>|<cond>|<run>").
func Oid(cls, sc, model, cond string, run int) string {
	h := sha1.Sum([]byte(fmt.Sprintf("%s|%s|%s|%s|%d", cls, sc, model, cond, run))) //nolint:gosec // committed id scheme
	return "r" + hex.EncodeToString(h[:])[:12]
}

// Chunk calls fn with each consecutive run of up to size items, in order, and the
// run's index. The last run holds the remainder. Each chunk shares items' backing
// array.
func Chunk[T any](items []T, size int, fn func(index int, chunk []T)) {
	for i := 0; i < len(items); i += size {
		end := min(i+size, len(items))
		fn(i/size, items[i:end])
	}
}
