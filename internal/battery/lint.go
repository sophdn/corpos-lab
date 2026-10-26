package battery

// BuildStructuralLint assembles the deterministic structural subset of the
// ALPHABET battery: the items that judge a candidate's shape from its text
// alone, with no model call — item 2 (intent-language scan), item 3 (duplicate
// identity), item 10 (three-axis presence), and item 13 (contamination radius).
//
// It is the fast, model-free pre-battery structural check: a candidate that
// fails it is malformed before any model-judged item is worth spending. Run it
// with Input.ContinueOnFail so every item is measured, not just the first
// failure. Run it against the extracted entry (ExtractEntry), not the raw
// candidate doc, so the AC-3 specimen's project-scoped text does not trip the
// intent scan.
//
// Item 3 reads Input.Registry; the other three read only Input.Content. This
// sequence is the intended host for the typed-schema enforcement of suggestion
// 166 (typed-enforced-glyph-file-schema-cut-bloat): a schema-validation step is
// added here when that lands, and every consumer of the lint picks it up with
// no change.
func BuildStructuralLint() *Sequence {
	seq := NewSequence("structural-lint")
	seq.AddStep("item2-intent-language-scan", Item2IntentLanguageScan)
	seq.AddStep("item3-duplicate-check", Item3DuplicateCheck)
	seq.AddStep("item10-axis-presence", Item10AxisPresence)
	seq.AddStep("item13-contamination-radius", Item13ContaminationRadius)
	return seq
}
