// Package neutralprefix ports the neutral-prefix-control provenance scoring
// python to Go (chain eradicate-python-rewrite-in-go). It backs the published
// paper "Comprehension as Compliance" (10.5281/zenodo.22846123).
//
// The logic is sans-IO: callers read the study's committed inputs and pass the
// data in; cmd/corpos-lab/pubscore_neutralprefix.go does the wiring. The two
// serialization organs here — a Python-compatible JSON encoder (pyjson.go) and a
// faithful CPython Mersenne-Twister port (pyrandom.go) — let the Go rewrites
// reproduce the exact committed artifacts byte-for-byte, so the python oracles
// can be deleted.
package neutralprefix

import (
	"fmt"
	"strconv"
	"strings"
)

// OMap is an insertion-ordered string-keyed map. Go's built-in map iterates in a
// randomized order and json.Marshal sorts keys, so neither reproduces a Python
// dict's insertion order; OMap does, which is what byte-identical output needs.
type OMap struct {
	keys []string
	vals []any
}

// NewOMap returns an empty ordered map.
func NewOMap() *OMap { return &OMap{} }

// Set appends key k with value v and returns the map for chaining.
func (m *OMap) Set(k string, v any) *OMap {
	m.keys = append(m.keys, k)
	m.vals = append(m.vals, v)
	return m
}

// Len reports the number of entries.
func (m *OMap) Len() int { return len(m.keys) }

// PyDumps renders v the way CPython's json.dumps does. indent<0 selects the
// compact form (separators ", " and ": "); indent>=0 selects the newline form
// with that many spaces of indentation per nesting level (separators "," and
// ": "). ensureASCII mirrors the flag of the same name: when true, every rune
// above 0x7e is \u-escaped (with a surrogate pair above U+FFFF); when false such
// runes are emitted as raw UTF-8. Supported value types are string, int, int64,
// *OMap, and []any; anything else panics, which is a programming error in a
// caller, not a runtime input fault.
func PyDumps(v any, indent int, ensureASCII bool) string {
	var b strings.Builder
	pyEncode(&b, v, indent, ensureASCII, 0)
	return b.String()
}

func pyEncode(b *strings.Builder, v any, indent int, ascii bool, depth int) {
	switch x := v.(type) {
	case string:
		pyEncodeString(b, x, ascii)
	case int:
		b.WriteString(strconv.Itoa(x))
	case int64:
		b.WriteString(strconv.FormatInt(x, 10))
	case *OMap:
		pyEncodeObj(b, x, indent, ascii, depth)
	case []any:
		pyEncodeArr(b, x, indent, ascii, depth)
	default:
		panic(fmt.Sprintf("pyjson: unsupported type %T", v))
	}
}

func pyEncodeObj(b *strings.Builder, m *OMap, indent int, ascii bool, depth int) {
	if m.Len() == 0 {
		b.WriteString("{}")
		return
	}
	b.WriteByte('{')
	for i, k := range m.keys {
		if i > 0 {
			if indent < 0 {
				b.WriteString(", ")
			} else {
				b.WriteByte(',')
			}
		}
		if indent >= 0 {
			writeNL(b, indent, depth+1)
		}
		pyEncodeString(b, k, ascii)
		b.WriteString(": ")
		pyEncode(b, m.vals[i], indent, ascii, depth+1)
	}
	if indent >= 0 {
		writeNL(b, indent, depth)
	}
	b.WriteByte('}')
}

func pyEncodeArr(b *strings.Builder, arr []any, indent int, ascii bool, depth int) {
	if len(arr) == 0 {
		b.WriteString("[]")
		return
	}
	b.WriteByte('[')
	for i, e := range arr {
		if i > 0 {
			if indent < 0 {
				b.WriteString(", ")
			} else {
				b.WriteByte(',')
			}
		}
		if indent >= 0 {
			writeNL(b, indent, depth+1)
		}
		pyEncode(b, e, indent, ascii, depth+1)
	}
	if indent >= 0 {
		writeNL(b, indent, depth)
	}
	b.WriteByte(']')
}

func writeNL(b *strings.Builder, indent, level int) {
	b.WriteByte('\n')
	if indent > 0 {
		b.WriteString(strings.Repeat(" ", indent*level))
	}
}

// pyEncodeString writes s as a JSON string literal matching CPython's encoder:
// the mandatory escapes (" \ and the C0 controls, with short forms for the named
// ones), and, when ascii is true, \u escapes for every rune above 0x7e. It never
// escapes '<', '>', '&', or '/', which Go's encoding/json escapes by default.
func pyEncodeString(b *strings.Builder, s string, ascii bool) {
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		case '\b':
			b.WriteString(`\b`)
		case '\f':
			b.WriteString(`\f`)
		default:
			switch {
			case r < 0x20:
				fmt.Fprintf(b, `\u%04x`, r)
			case ascii && r > 0x7e:
				if r > 0xffff {
					r2 := r - 0x10000
					hi := 0xd800 + (r2 >> 10)
					lo := 0xdc00 + (r2 & 0x3ff)
					fmt.Fprintf(b, `\u%04x\u%04x`, hi, lo)
				} else {
					fmt.Fprintf(b, `\u%04x`, r)
				}
			default:
				b.WriteRune(r)
			}
		}
	}
	b.WriteByte('"')
}
