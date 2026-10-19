// Package disclosure implements the default-closed progressive-disclosure rule
// for corpos-lab.
//
// The full glyph canon lives in the private corpus repo. Only an explicitly
// approved subset is published to the public tree. Nothing is public unless it
// appears in the approved manifest AND is certified: the default is closed, so
// a glyph reaches the public GitHub mirror only by deliberate approval, never
// by omission.
//
// The package is pure logic and embeds no corpus data. Selection and public-tree
// validation take slug sets as arguments, so the public-repo CI can validate the
// public tree against the approved manifest without any access to the private
// canon.
package disclosure

import (
	"fmt"
	"sort"
	"strings"
)

// Selection is the outcome of applying an approved-subset manifest to the set of
// certified glyphs.
//
//   - Publish is the subset that may go public: approved AND certified.
//   - Withheld is certified but not approved. It stays private. This is the
//     default-closed majority.
//   - Missing is approved but not certified. It is an error condition: an
//     uncertified glyph must never be published, so a non-empty Missing means
//     the manifest names a slug that the canon cannot back.
type Selection struct {
	Publish  []string
	Withheld []string
	Missing  []string
}

// Select applies the approved manifest to the certified set and returns the
// partition. It is pure: it takes the two slug sets as data and returns the
// result, sorted for determinism. A slug repeated in either input counts once.
func Select(certified, approved []string) Selection {
	certSet := toSet(certified)
	apprSet := toSet(approved)

	sel := Selection{}
	for slug := range certSet {
		if _, ok := apprSet[slug]; ok {
			sel.Publish = append(sel.Publish, slug)
		} else {
			sel.Withheld = append(sel.Withheld, slug)
		}
	}
	for slug := range apprSet {
		if _, ok := certSet[slug]; !ok {
			sel.Missing = append(sel.Missing, slug)
		}
	}

	sort.Strings(sel.Publish)
	sort.Strings(sel.Withheld)
	sort.Strings(sel.Missing)
	return sel
}

// Err reports whether the selection is unsafe to act on. It returns a non-nil
// error when the manifest approves any slug the canon has not certified, naming
// each offending slug.
func (s Selection) Err() error {
	if len(s.Missing) == 0 {
		return nil
	}
	return fmt.Errorf("disclosure: %d approved glyph(s) are not certified and cannot be published: %s",
		len(s.Missing), strings.Join(s.Missing, ", "))
}

// ValidatePublicTree checks that the glyphs present in the public tree are
// exactly the approved set — no more, no fewer.
//
// A present-but-not-approved glyph is a disclosure leak: something reached the
// public tree without approval. An approved-but-absent glyph means promotion did
// not run or did not finish. Either is a failure, so the check is exact. The
// returned error names every discrepancy; a clean tree returns nil.
func ValidatePublicTree(present, approved []string) error {
	presentSet := toSet(present)
	approvedSet := toSet(approved)

	var leaked, absent []string
	for slug := range presentSet {
		if _, ok := approvedSet[slug]; !ok {
			leaked = append(leaked, slug)
		}
	}
	for slug := range approvedSet {
		if _, ok := presentSet[slug]; !ok {
			absent = append(absent, slug)
		}
	}
	sort.Strings(leaked)
	sort.Strings(absent)

	var problems []string
	if len(leaked) > 0 {
		problems = append(problems, fmt.Sprintf("LEAK — %d unapproved glyph(s) present in the public tree: %s",
			len(leaked), strings.Join(leaked, ", ")))
	}
	if len(absent) > 0 {
		problems = append(problems, fmt.Sprintf("%d approved glyph(s) missing from the public tree: %s",
			len(absent), strings.Join(absent, ", ")))
	}
	if len(problems) == 0 {
		return nil
	}
	return fmt.Errorf("disclosure: public tree does not match the approved manifest: %s",
		strings.Join(problems, "; "))
}

// toSet returns the unique, non-empty members of slugs as a set.
func toSet(slugs []string) map[string]struct{} {
	set := make(map[string]struct{}, len(slugs))
	for _, s := range slugs {
		if s == "" {
			continue
		}
		set[s] = struct{}{}
	}
	return set
}
