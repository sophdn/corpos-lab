package battery

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

// intentRejectPhrases are the intent-modeling constructions item 2 rejects:
// each makes the firing condition uncheckable from observable artifacts.
// Order matters — the first match is the reported phrase (source parity).
var intentRejectPhrases = []string{
	"when the agent decides",
	"when the agent believes",
	"when the agent thinks",
	"when the agent is confused",
	"when the agent is under pressure",
	"when the agent intends",
	"when the agent wants",
	"when the agent feels",
}

// Item2IntentLanguageScan — firing condition observability. Scans for
// intent-modeling language that makes the firing condition uncheckable
// from observable artifacts.
func Item2IntentLanguageScan(_ context.Context, st *State) StepOutcome {
	contentLower := strings.ToLower(st.Content)
	for _, phrase := range intentRejectPhrases {
		if strings.Contains(contentLower, phrase) {
			return FailItemOutcome(2,
				fmt.Sprintf("firing condition contains intent-modeling language: %q", phrase))
		}
	}
	return PassOutcome()
}

// DeferredPending is the single-source deferred-item registry: why each
// deferred item cannot yet run and what would unblock it. Kept as a switch
// so the list is grep-able by item number and edits are visible at review
// time (source parity with static_steps.rs).
//
// Only items 5, 7, 8, and 14 remain deferred: items 3 and 13 are now
// deterministic (Item3DuplicateCheck, Item13ContaminationRadius) and items 6,
// 11, and 12 are model-assessed verdict steps. A separate task decides 5, 7, 8,
// and 14.
func DeferredPending(item int) string {
	switch item {
	case 5:
		return "item 5: Y-not-fire positive terrain — complex multi-sub-check, decomposition cross-reference to GLYPH_DECOMPOSITION_PROCESS.md pending"
	case 7:
		return "item 7: sister/mirror check — corpus-wide scan pending (corpus access)"
	case 8:
		return "item 8: phenomenological grounding — multi-faceted judgment across Y-fire, Y-not-fire, and three axes; underspecified for single-agent step"
	case 14:
		return "item 14: globality demand — judgment-heavy, underspecified for single-agent step"
	default:
		return "item deferred pending integration"
	}
}

// DeferredStep builds the stub StepFn for a deferred item: it emits a
// Deferred verdict with the item's pending reason. Deferred is not a
// stopping condition, so the 15-item battery registers every item in
// stratum order while integration work is outstanding.
func DeferredStep(item int) StepFn {
	return func(_ context.Context, _ *State) StepOutcome {
		return VerdictOutcome(Deferred(DeferredPending(item)))
	}
}

// Item4NA — the registry-summary requirement applies to TABOO_REGISTRY
// entries only; for glyph entries item 4 is structurally inapplicable
// (NotApplicable, not "pass with a note").
func Item4NA(_ context.Context, _ *State) StepOutcome {
	return VerdictOutcome(NotApplicable("item 4 does not apply to glyph entries"))
}

// Item10AxisPresence — three-axis coverage or honest gap notation. Requires
// a Y marker section, then Marker/Aim/Rest each present as a heading
// (`**<axis>`), a mention (`<axis> axis`), or gap notation (`<axis>:`),
// case-insensitively.
func Item10AxisPresence(_ context.Context, st *State) StepOutcome {
	content := st.Content

	hasYMarker := strings.Contains(content, "**Y") ||
		strings.Contains(content, "Y —") ||
		strings.Contains(content, "Y marker")
	if !hasYMarker {
		return FailItemOutcome(10, "no Y marker section found")
	}

	contentLower := strings.ToLower(content)
	var missing []string
	for _, axis := range []string{"Marker", "Aim", "Rest"} {
		axisLower := strings.ToLower(axis)
		found := strings.Contains(contentLower, "**"+axisLower) ||
			strings.Contains(contentLower, axisLower+" axis") ||
			strings.Contains(contentLower, axisLower+":")
		if !found {
			missing = append(missing, axis)
		}
	}

	if len(missing) == 0 {
		return PassOutcome()
	}
	return FailItemOutcome(10,
		"missing axis coverage for: "+strings.Join(missing, ", "))
}

// falloutProfileMarker is the entry field Item 15 reads: the text following it
// on the same line is the reference to the fallout-profile document.
const falloutProfileMarker = "**Fallout profile:**"

// falloutDimensions are the five dimensions ALPHABET_ENTRY_BATTERY.md Item 15
// requires the referenced profile to address. A null finding on a dimension is
// a valid result, but the dimension must be explicitly present.
var falloutDimensions = []string{
	"attentional shift",
	"over-application",
	"meta-awareness",
	"scope creep",
	"suppression effects",
}

// Item15FalloutProfile — fallout characterization. The entry must carry a
// `**Fallout profile:**` field, the referenced document must resolve and be
// readable, and it must address all five fallout dimensions. This is a
// characterization gate: only presence and dimensional coverage are checked,
// never the correctness of the analysis.
//
// The step performs pure path math (resolving the reference relative to the
// entry file) and delegates the read to the injected ProfileReader, so the
// package stays sans-IO. It fails closed on a missing field, a missing reader,
// an unreadable referent, or any unaddressed dimension — a battery must never
// manufacture a pass it did not verify, which is exactly what the old
// substring-only check did for the whole window the referents were deleted.
func Item15FalloutProfile(_ context.Context, st *State) StepOutcome {
	ref, ok := falloutProfileRef(st.Content)
	if !ok {
		return FailItemOutcome(15, "no **Fallout profile:** field found in entry")
	}
	if st.Profiles == nil {
		return FailItemOutcome(15,
			fmt.Sprintf("no profile reader configured to resolve %q", ref))
	}

	resolved := ref
	if st.EntryPath != "" {
		resolved = filepath.Join(filepath.Dir(st.EntryPath), ref)
	}

	doc, err := st.Profiles.ReadProfile(resolved)
	if err != nil {
		return FailItemOutcome(15,
			fmt.Sprintf("fallout profile %q is missing or unreadable: %v", resolved, err))
	}

	if missing := missingFalloutDimensions(doc); len(missing) > 0 {
		return FailItemOutcome(15,
			"fallout profile does not address: "+strings.Join(missing, ", "))
	}
	return PassOutcome()
}

// falloutProfileRef extracts the reference following the `**Fallout profile:**`
// marker on its line, returning false when the field is absent or empty.
func falloutProfileRef(content string) (string, bool) {
	for _, line := range strings.Split(content, "\n") {
		idx := strings.Index(line, falloutProfileMarker)
		if idx < 0 {
			continue
		}
		ref := strings.TrimSpace(line[idx+len(falloutProfileMarker):])
		if ref == "" {
			continue
		}
		return ref, true
	}
	return "", false
}

// missingFalloutDimensions returns the dimensions the profile does not address,
// in canonical order; an empty result means all five are covered.
func missingFalloutDimensions(doc string) []string {
	var missing []string
	for _, dim := range falloutDimensions {
		if !dimensionAddressed(doc, dim) {
			missing = append(missing, dim)
		}
	}
	return missing
}

// glyphMarker and candidateHeadingMarker are the two forms an entry declares
// its identity in: the canonical `**Glyph:** ` + "`slug`" field of an ALPHABET
// entry, and the `# Glyph Candidate: <slug>` heading of a candidate file. The
// duplicate check reads either.
const (
	glyphMarker            = "**Glyph:**"
	candidateHeadingMarker = "# Glyph Candidate:"
)

// GlyphIdentity extracts the entry's identity slug for the duplicate check: the
// slug of the canonical `**Glyph:** ` + "`slug`" field if present, else the slug
// in a `# Glyph Candidate: <slug>` heading. Exported so a registry reader can
// parse promoted entries with the same rule the candidate is read by — the two
// identities are compared, so they must be extracted identically. Returns false
// when the entry declares no identity at all.
func GlyphIdentity(content string) (string, bool) {
	if id, ok := markerValue(content, glyphMarker, true); ok {
		return id, true
	}
	return markerValue(content, candidateHeadingMarker, false)
}

// markerValue returns the text following marker on its line, backtick-stripped
// and trimmed. When headingOnly is true the marker must begin the line (an ATX
// heading); otherwise it may appear anywhere on the line (a bold field).
func markerValue(content, marker string, headingOnly bool) (string, bool) {
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		idx := strings.Index(trimmed, marker)
		if idx < 0 || (headingOnly && idx != 0) {
			continue
		}
		val := strings.TrimSpace(trimmed[idx+len(marker):])
		val = strings.TrimSpace(strings.Trim(val, "`"))
		if val != "" {
			return val, true
		}
	}
	return "", false
}

func normalizeIdentity(s string) string { return strings.ToLower(strings.TrimSpace(s)) }

// RegistryIdentitiesFrom extracts every promoted-entry identity from a registry
// document — one per `**Glyph:** ` + "`slug`" field. A registry file lists many
// entries, so unlike GlyphIdentity (which returns the first identity of a single
// candidate) this returns all of them. Exported so a filesystem RegistryReader
// parses the registry with the same rule the candidate identity is read by.
func RegistryIdentitiesFrom(content string) []string {
	var ids []string
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		idx := strings.Index(trimmed, glyphMarker)
		if idx < 0 {
			continue
		}
		val := strings.TrimSpace(strings.Trim(strings.TrimSpace(trimmed[idx+len(glyphMarker):]), "`"))
		if val != "" {
			ids = append(ids, val)
		}
	}
	return ids
}

// Item3DuplicateCheck — duplicate check against the ALPHABET registry.
//
// DETERMINISTIC (no model call) — recorded per-item sub-decision: this is an
// exact structural identity comparison, not a semantic judgment, so a
// reproducible registry lookup is the right instrument and an LLM verdict would
// be non-reproducible here. It matches the archived structural-prober's
// prove_no_duplicate, which did an exact identity match against a registry
// slice. Semantic near-duplicate detection (a rephrased or narrower version of a
// promoted glyph — ALPHABET_ENTRY_BATTERY.md Item 3) is beyond an exact match
// and stays with the promotion-time assessor; this step catches the
// exact-identity collision the prober caught. ALPHABET is currently empty, so a
// clean candidate passes — but the check is wired to read the registry, so it
// starts failing the moment a colliding identity is promoted.
//
// Fails closed on a nil reader (parity with Item 15): the battery must not pass
// a duplicate check it could not run. The reader reports an ABSENT registry as
// an empty identity list with no error (a verified-empty registry has no
// duplicates); only a real read failure returns an error, which fails the item.
//
// Self-exclusion: a glyph is never a duplicate of itself. Re-certifying a
// candidate already promoted to the registry finds its OWN identity there, which
// must not false-fail as a duplicate. So the first registry occurrence of the
// candidate's identity is skipped as the candidate's own entry; the item fails
// only if a SECOND occurrence remains — a genuinely distinct entry sharing the
// identity.
func Item3DuplicateCheck(_ context.Context, st *State) StepOutcome {
	id, ok := GlyphIdentity(st.Content)
	if !ok {
		return FailItemOutcome(3,
			"no glyph identity (**Glyph:** field or # Glyph Candidate: heading) found to run the duplicate check")
	}
	if st.Registry == nil {
		return FailItemOutcome(3,
			fmt.Sprintf("no registry reader configured to check %q for duplicates", id))
	}
	identities, err := st.Registry.RegistryIdentities()
	if err != nil {
		return FailItemOutcome(3,
			fmt.Sprintf("registry unreadable, cannot verify %q is unique: %v", id, err))
	}
	want := normalizeIdentity(id)
	selfExcluded := false
	for _, existing := range identities {
		if normalizeIdentity(existing) != want {
			continue
		}
		if !selfExcluded {
			// The candidate's own promoted entry — not a duplicate of itself.
			selfExcluded = true
			continue
		}
		return FailItemOutcome(3,
			fmt.Sprintf("duplicate identity %q already present in the ALPHABET registry as a distinct entry", id))
	}
	return PassOutcome()
}

// highRadiusThreshold is the dependency-surface size above which contamination
// radius is judged high — parity with the structural-prober's
// HIGH_RADIUS_THRESHOLD.
const highRadiusThreshold = 5

// highRiskCategories are the decision-class categories ALPHABET_ENTRY_BATTERY.md
// Item 13 names as carrying high contamination radius by default. Their presence
// is advisory (a FLAG), matching the prover — the assessor adjudicates category
// severity; the deterministic step reports the signal reproducibly.
var highRiskCategories = []string{
	"corpus-design", "trust-class", "source-assessment", "authority-position",
}

// slugRefPattern matches a backtick-quoted lowercase-hyphen slug token
// (e.g. `casg-direct`). It deliberately excludes tokens with slashes or dots so
// backtick-quoted file paths (which Item 9 handles) are not counted as
// dependency references.
var slugRefPattern = regexp.MustCompile("`([a-z0-9]+(?:-[a-z0-9]+)*)`")

// Item13ContaminationRadius — contamination radius (dependency-surface count).
//
// DETERMINISTIC (no model call) — recorded per-item sub-decision: this is a
// count-and-threshold check over a structural surface, so a reproducible count
// is the right instrument and an LLM verdict would be non-reproducible. It
// matches the archived structural-prober's prove_contained_radius: count the
// candidate's own dependency surface against a threshold, then FLAG a high-risk
// category. Per the task's framing (item 13 = "a dependency-surface count /
// threshold check"), this reads the candidate SELF-CONTAINED — the outgoing
// surface the entry declares — not a registry scan.
//
// Dependency surface = the distinct OTHER glyph slugs the entry references in
// backticks (its own identity excluded). The corpus-wide INCOMING reading
// ("how many other promoted glyphs depend on this one", ALPHABET_ENTRY_BATTERY.md
// Item 13) needs the full cross-entry graph and is a future tightening once the
// registry is populated; flagged for the follow-up. Count > threshold → FAIL
// (high radius); a high-risk category present → FLAG (advisory); else PASS.
func Item13ContaminationRadius(_ context.Context, st *State) StepOutcome {
	own := ""
	if id, ok := GlyphIdentity(st.Content); ok {
		own = normalizeIdentity(id)
	}

	seen := map[string]struct{}{}
	for _, m := range slugRefPattern.FindAllStringSubmatch(st.Content, -1) {
		slug := normalizeIdentity(m[1])
		if slug == own {
			continue
		}
		seen[slug] = struct{}{}
	}
	if len(seen) > highRadiusThreshold {
		return FailItemOutcome(13,
			fmt.Sprintf("dependency surface has %d referenced glyphs (threshold %d): tighten category scoping",
				len(seen), highRadiusThreshold))
	}

	contentLower := strings.ToLower(st.Content)
	for _, cat := range highRiskCategories {
		if strings.Contains(contentLower, cat) {
			return VerdictOutcome(Flag(fmt.Sprintf(
				"decision class touches high-contamination-radius category %q; assessor must confirm scoping", cat)))
		}
	}
	return PassOutcome()
}

// dimensionAddressed reports whether the profile carries a section heading for
// the dimension in either form the restored profiles use: a bold run
// (`**Attentional shift.**`) or an ATX heading (`### Attentional shift`), at
// the start of a line. Matching the heading forms — not a bare substring — is
// deliberate: every profile opens with a five-item definition preamble that
// lists all five dimensions as `- **<dim>:**` list items, so a substring scan
// would report full coverage even for a profile whose analysis omits a
// dimension. List items begin with "-", so anchoring on a leading "###" or
// "**" excludes the preamble and matches only genuine section headings.
func dimensionAddressed(doc, dim string) bool {
	for _, line := range strings.Split(doc, "\n") {
		trimmed := strings.TrimSpace(line)
		var heading string
		switch {
		case strings.HasPrefix(trimmed, "###"):
			heading = strings.TrimLeft(trimmed, "#")
		case strings.HasPrefix(trimmed, "**"):
			heading = strings.TrimLeft(trimmed, "*")
		default:
			continue
		}
		heading = strings.ToLower(strings.TrimSpace(heading))
		if strings.HasPrefix(heading, dim) {
			return true
		}
	}
	return false
}
