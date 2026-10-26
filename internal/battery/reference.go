package battery

import "strings"

// ReferenceMaterial is the corpus reference text the model-assessed items
// deliver INLINE to the judge, so a coherence check traces each structural
// field against the actual definition rules rather than against the judge's
// own memory of them.
//
// It exists because the item-6 judge failed a coherent glyph by claiming a
// definition rule was missing when the rule was in fact present in the
// definition (bug qwen-model-assessed-battery-items-unreliable-item6-hallucination,
// 1334): the item-6 prompt asked the judge to trace every field back to "the
// glyph definition" but never handed it the definition, so the judge answered
// from memory and hallucinated a gap. Handing it the definition text — and
// telling it to trace against THAT text — closes the hallucination path
// (suggestion deliver-reference-material-inline-to-model-assessed-battery-items,
// 175).
//
// The zero value means no reference was wired: item 6 then falls back to its
// prior definition-from-memory prompt rather than failing a run. Nothing here
// may fail a run — an absent reference is a recorded degradation, not a refusal.
type ReferenceMaterial struct {
	// GlyphDefinition is the glyph-definition narrative (GLYPH_DEFINITION.md):
	// the rules item 6 traces every structural field back to. Delivered inline
	// to the item-6 judge whenever it is non-empty.
	GlyphDefinition string
	// ProvenanceTypes is the provenance-type taxonomy (GLYPH_PROVENANCE_TYPES.md),
	// a document SEPARATE from the definition. It is delivered inline only when
	// the entry cites a provenance type, so the judge is handed only the segment
	// the entry actually depends on (condition-appropriate delivery) rather than
	// the whole taxonomy on every entry.
	ProvenanceTypes string
}

// provenanceTypeNames are the five provenance types defined in
// GLYPH_PROVENANCE_TYPES.md. An entry "cites a provenance type" when its content
// names one of them; that is the condition under which the provenance taxonomy
// is relevant to a coherence trace and is delivered inline.
var provenanceTypeNames = []string{
	"Existence",
	"Recency",
	"Authorization",
	"Propagation",
	"Anachronicity",
}

// citesProvenanceType reports whether the entry names one of the five
// provenance types. It matches the capitalized type name as it appears in the
// taxonomy and in a glyph's Y-fire provenance line (e.g. "provenance type:
// Anachronicity"). The check is intentionally biased toward delivery: a false
// positive over-delivers a short taxonomy the judge can ignore, whereas a false
// negative reintroduces the out-of-scope-reference failure the inline delivery
// exists to prevent (suggestion 175, glyph #3).
func citesProvenanceType(content string) bool {
	for _, name := range provenanceTypeNames {
		if strings.Contains(content, name) {
			return true
		}
	}
	return false
}
