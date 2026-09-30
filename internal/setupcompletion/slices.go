package setupcompletion

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// Conditions is the grounded-glyph condition set, in the order build_slices.py
// and tally.py iterate it.
var Conditions = []string{"baseline", "glyph_only", "imperative_only"}

// OpaqueID returns a 16-hex-char content hash of a run's identity, matching
// build_slices.opaque_id: sha256 over study, cond, run, text joined by NUL bytes,
// truncated to the first 16 hex chars. build_slices passes setup+glyph as `study`,
// so the id leaks neither the condition nor the setup.
func OpaqueID(study, cond string, run int, text string) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%s\x00%s\x00%d\x00%s", study, cond, run, text)))
	return hex.EncodeToString(sum[:])[:16]
}

// RunRecord is one collected run response, the pure input to BuildSlices. The FS
// walk that discovers these lives in the cmd IO seam.
type RunRecord struct {
	Glyph     string
	Setup     string
	Condition string
	Run       int
	Text      string
}

// SliceItem is one line of a per-glyph blind rating slice.
type SliceItem struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

// KeyEntry de-anonymises one opaque id after scoring.
type KeyEntry struct {
	Glyph     string `json:"glyph"`
	Condition string `json:"condition"`
	Setup     string `json:"setup"`
	Run       int    `json:"run"`
}

// BuildSlices turns collected run records into per-glyph blind slices and the
// de-anonymising key, matching build_slices.main. Each item's id is the opaque
// content hash over setup+glyph, condition, run, text; items within a glyph are
// sorted by id (stable — order leaks nothing). The returned glyphs slice is
// sorted, for deterministic iteration by the writer. Callers do any --glyph
// filtering before passing records in.
func BuildSlices(records []RunRecord) (glyphs []string, perGlyph map[string][]SliceItem, key map[string]KeyEntry) {
	perGlyph = map[string][]SliceItem{}
	key = map[string]KeyEntry{}
	for _, r := range records {
		rid := OpaqueID(r.Setup+r.Glyph, r.Condition, r.Run, r.Text)
		perGlyph[r.Glyph] = append(perGlyph[r.Glyph], SliceItem{ID: rid, Text: r.Text})
		key[rid] = KeyEntry{Glyph: r.Glyph, Condition: r.Condition, Setup: r.Setup, Run: r.Run}
	}
	for g := range perGlyph {
		items := perGlyph[g]
		sort.SliceStable(items, func(i, j int) bool { return items[i].ID < items[j].ID })
		glyphs = append(glyphs, g)
	}
	sort.Strings(glyphs)
	return glyphs, perGlyph, key
}

// SetupFromStudy maps a study dir name to its setup arm, matching build_slices'
// prefix test: "setup-raw-*" -> raw, "setup-loop-*" -> loop, else ok is false
// (the study is skipped).
func SetupFromStudy(study string) (setup string, ok bool) {
	switch {
	case strings.HasPrefix(study, "setup-raw-"):
		return "raw", true
	case strings.HasPrefix(study, "setup-loop-"):
		return "loop", true
	default:
		return "", false
	}
}

// IsSmokeStudy reports whether a study dir is a smoke run to skip, matching
// build_slices' `"smoke" in study or study.startswith("sm-")`.
func IsSmokeStudy(study string) bool {
	return strings.Contains(study, "smoke") || strings.HasPrefix(study, "sm-")
}

// RunFromStem parses the run number from a response filename stem (no extension),
// matching build_slices' int(stem.rsplit("_", 1)[1]): the integer after the last
// underscore. It errors if there is no underscore or the tail is not an integer.
func RunFromStem(stem string) (int, error) {
	i := strings.LastIndex(stem, "_")
	if i < 0 {
		return 0, fmt.Errorf("no underscore in stem %q", stem)
	}
	return strconv.Atoi(stem[i+1:])
}
