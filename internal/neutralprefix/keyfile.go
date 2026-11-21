package neutralprefix

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// KeyMeta is one response's metadata as stored in scoring/key.json. Scenario is a
// string ("1"/"2") in the committed file, matching build_slices.py's output;
// build_anchor.py, aggregate.py and reconcile_rerate.py all read it as such.
type KeyMeta struct {
	Cls       string `json:"cls"`
	Scenario  string `json:"scenario"`
	Model     string `json:"model"`
	Condition string `json:"condition"`
	Run       int    `json:"run"`
}

// KeyFile is scoring/key.json with its rid order preserved. reconcile_rerate.py
// and aggregate.py iterate the key in insertion order, and that order is
// load-bearing for reconcile's per-condition consensus histogram, so the Go port
// must not fall back to a Go map's randomized iteration.
type KeyFile struct {
	Order []string
	Meta  map[string]KeyMeta
}

// ParseKeyFile parses key.json bytes into a KeyFile, preserving the object's key
// order via a streaming decoder (a plain map decode would lose it).
func ParseKeyFile(data []byte) (*KeyFile, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	tok, err := dec.Token()
	if err != nil {
		return nil, err
	}
	if d, ok := tok.(json.Delim); !ok || d != '{' {
		return nil, fmt.Errorf("key.json: expected object, got %v", tok)
	}
	kf := &KeyFile{Meta: map[string]KeyMeta{}}
	for dec.More() {
		kt, err := dec.Token()
		if err != nil {
			return nil, err
		}
		rid, ok := kt.(string)
		if !ok {
			return nil, fmt.Errorf("key.json: expected string key, got %v", kt)
		}
		var m KeyMeta
		if err := dec.Decode(&m); err != nil {
			return nil, fmt.Errorf("key.json[%s]: %w", rid, err)
		}
		kf.Order = append(kf.Order, rid)
		kf.Meta[rid] = m
	}
	// Require the closing '}' so a truncated object is rejected rather than
	// silently accepted mid-stream.
	if _, err := dec.Token(); err != nil {
		return nil, err
	}
	return kf, nil
}
