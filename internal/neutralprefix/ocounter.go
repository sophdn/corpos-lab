package neutralprefix

import (
	"fmt"
	"strings"
)

// OCounter is an insertion-ordered counter, the shape of a collections.Counter
// whose repr and iteration follow first-seen key order. reconcile_rerate.py and
// build_slices.py both print a dict(Counter(...)), whose byte form depends on
// that order.
type OCounter struct {
	Keys []string
	vals map[string]int
}

// NewOCounter returns an empty ordered counter.
func NewOCounter() *OCounter { return &OCounter{vals: map[string]int{}} }

// Inc adds one to k, recording k's position on first sight.
func (c *OCounter) Inc(k string) {
	if _, ok := c.vals[k]; !ok {
		c.Keys = append(c.Keys, k)
	}
	c.vals[k]++
}

// Get returns k's count (zero if absent).
func (c *OCounter) Get(k string) int { return c.vals[k] }

// ToOMap converts the counter to an ordered map of key -> int count.
func (c *OCounter) ToOMap() *OMap {
	m := NewOMap()
	for _, k := range c.Keys {
		m.Set(k, c.vals[k])
	}
	return m
}

// PyDictRepr renders the counter as a Python dict repr, e.g. {'C': 12, 'N': 3},
// matching what print(dict(Counter(...))) emits (single-quoted string keys, ", "
// separators, first-seen order). An empty counter renders as {}.
func (c *OCounter) PyDictRepr() string {
	if len(c.Keys) == 0 {
		return "{}"
	}
	parts := make([]string, len(c.Keys))
	for i, k := range c.Keys {
		parts[i] = fmt.Sprintf("'%s': %d", k, c.vals[k])
	}
	return "{" + strings.Join(parts, ", ") + "}"
}
