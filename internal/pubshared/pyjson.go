package pubshared

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

// omember is one key/value pair of an ordered JSON object.
type omember struct {
	key string
	val any
}

// oobject is a JSON object with its keys kept in input order, so a round-trip
// through DumpsPy reproduces CPython's insertion-order json.dumps output.
type oobject []omember

func (o oobject) get(key string) (any, bool) {
	for _, m := range o {
		if m.key == key {
			return m.val, true
		}
	}
	return nil, false
}

func (o oobject) keys() []string {
	ks := make([]string, len(o))
	for i, m := range o {
		ks[i] = m.key
	}
	return ks
}

// ParseOrdered parses JSON into an ordered value tree: objects become oobject
// (order-preserving), arrays []any, numbers json.Number, and strings/bools/null
// their Go equivalents.
func ParseOrdered(data []byte) (any, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	v, err := parseValue(dec)
	if err != nil {
		return nil, err
	}
	if dec.More() {
		return nil, fmt.Errorf("trailing data after JSON value")
	}
	return v, nil
}

func parseValue(dec *json.Decoder) (any, error) {
	tok, err := dec.Token()
	if err != nil {
		return nil, err
	}
	return parseFromToken(dec, tok)
}

func parseFromToken(dec *json.Decoder, tok json.Token) (any, error) {
	switch t := tok.(type) {
	case json.Delim:
		switch t {
		case '{':
			return parseObject(dec)
		case '[':
			return parseArray(dec)
		default:
			return nil, fmt.Errorf("unexpected delimiter %q", t)
		}
	default:
		return tok, nil // string, json.Number, bool, or nil
	}
}

func parseObject(dec *json.Decoder) (oobject, error) {
	obj := oobject{}
	for dec.More() {
		keyTok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		key, ok := keyTok.(string)
		if !ok {
			return nil, fmt.Errorf("object key is not a string: %v", keyTok)
		}
		val, err := parseValue(dec)
		if err != nil {
			return nil, err
		}
		obj = append(obj, omember{key, val})
	}
	if _, err := dec.Token(); err != nil { // consume '}'
		return nil, err
	}
	return obj, nil
}

func parseArray(dec *json.Decoder) ([]any, error) {
	arr := []any{}
	for dec.More() {
		v, err := parseValue(dec)
		if err != nil {
			return nil, err
		}
		arr = append(arr, v)
	}
	if _, err := dec.Token(); err != nil { // consume ']'
		return nil, err
	}
	return arr, nil
}

// DumpsPy serializes an ordered value tree the way CPython's json.dumps does with
// the default separators (", " and ": "), no indent. ensureAscii selects between
// json.dumps default (\uXXXX escaping of non-ASCII) and ensure_ascii=False.
func DumpsPy(v any, ensureAscii bool) string {
	var b strings.Builder
	dumpValue(&b, v, ensureAscii)
	return b.String()
}

func dumpValue(b *strings.Builder, v any, ensureAscii bool) {
	switch t := v.(type) {
	case nil:
		b.WriteString("null")
	case bool:
		if t {
			b.WriteString("true")
		} else {
			b.WriteString("false")
		}
	case string:
		b.WriteString(pyJSONString(t, ensureAscii))
	case json.Number:
		b.WriteString(string(t))
	case oobject:
		b.WriteByte('{')
		for i, m := range t {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(pyJSONString(m.key, ensureAscii))
			b.WriteString(": ")
			dumpValue(b, m.val, ensureAscii)
		}
		b.WriteByte('}')
	case []any:
		b.WriteByte('[')
		for i, e := range t {
			if i > 0 {
				b.WriteString(", ")
			}
			dumpValue(b, e, ensureAscii)
		}
		b.WriteByte(']')
	default:
		// Fallback for values built in Go (not from ParseOrdered).
		bts, _ := json.Marshal(v)
		b.Write(bts)
	}
}

// pyJSONString escapes a string as CPython's json encoder does. With
// ensureAscii=false it matches ensure_ascii=False (only \, ", the C0 controls
// escaped; every other rune passed through). With ensureAscii=true every
// non-ASCII rune is emitted as a \uXXXX escape (surrogate pair above U+FFFF),
// matching json.dumps' default.
func pyJSONString(s string, ensureAscii bool) string {
	var b strings.Builder
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '\\':
			b.WriteString(`\\`)
		case '"':
			b.WriteString(`\"`)
		case '\b':
			b.WriteString(`\b`)
		case '\f':
			b.WriteString(`\f`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		default:
			switch {
			case r < 0x20:
				fmt.Fprintf(&b, `\u%04x`, r)
			case ensureAscii && r > 0x7f:
				if r > 0xffff {
					r1, r2 := utf16Pair(r)
					fmt.Fprintf(&b, `\u%04x\u%04x`, r1, r2)
				} else {
					fmt.Fprintf(&b, `\u%04x`, r)
				}
			default:
				b.WriteRune(r)
			}
		}
	}
	b.WriteByte('"')
	return b.String()
}

// utf16Pair splits an astral rune into its UTF-16 surrogate pair.
func utf16Pair(r rune) (rune, rune) {
	r -= 0x10000
	return 0xd800 + (r >> 10), 0xdc00 + (r & 0x3ff)
}

// escapeScriptClose applies python's .replace("</", "<\\/") that keeps inline
// <script> data from closing the tag early.
func escapeScriptClose(s string) string {
	return strings.ReplaceAll(s, "</", `<\/`)
}
