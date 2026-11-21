package neutralprefix

import (
	"strings"
	"unicode"
)

// scoreTokens is score_phi4.py's CODE_RE alternation, in its exact order. The
// order is load-bearing: at a given position "Ii" and "Ic" are preferred over the
// shorter "C"/"I", matching Python re's leftmost, first-alternative-preferred
// semantics.
var scoreTokens = []string{"Ii", "Ic", "C", "I", "N"}

// isWordRune reports whether r is a regex word character (\w). Python's re treats
// str patterns as Unicode by default, so \b is a Unicode word boundary; matching
// that (letters, digits, underscore across Unicode) rather than Go regexp's
// ASCII-only \b is what keeps the parser byte-exact when a non-ASCII letter sits
// next to a code letter.
func isWordRune(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

// ClassOf derives the rubric class from a slice filename, matching
// score_phi4.class_of: the part before the first "__". The argument may be a bare
// base name or a full path; only the base name's prefix is used.
func ClassOf(slicePath string) string {
	base := slicePath
	if i := strings.LastIndexAny(base, "/\\"); i >= 0 {
		base = base[i+1:]
	}
	if i := strings.Index(base, "__"); i >= 0 {
		return base[:i]
	}
	return base
}

// BuildScorePrompt builds the phi-4 grader prompt byte-for-byte as
// score_phi4.score_one does: the rubric verbatim, the fixed instruction block, and
// the response text verbatim, wrapped in the phi-4 chat-template markers. Neither
// the rubric nor the text is trimmed — the python inserts them as-is.
func BuildScorePrompt(rubric, text string) string {
	return "<|im_start|>user<|im_sep|>" + rubric +
		"\n\n---\nScore the RESPONSE below against the rubric. Output ONLY the code as " +
		"the first token: one of C, Ii, Ic, I, N. No explanation.\n\nRESPONSE:\n" +
		text + "<|im_end|><|im_start|>assistant<|im_sep|>"
}

// ParseScoreLabel extracts the score code from the rater's completion, matching
// score_phi4.score_one's parse tail: the first \b(Ii|Ic|C|I|N)\b match, else "N".
// It scans rune positions left to right and, at each word-boundary start, tries the
// tokens in scoreTokens order, requiring a word boundary at the token end too — a
// hand-rolled equivalent of re.search's leftmost-first behavior with Unicode \b.
func ParseScoreLabel(content string) string {
	rs := []rune(content)
	n := len(rs)
	for i := 0; i < n; i++ {
		// A token starts with a word char, so a \b at i requires the preceding
		// rune (if any) to be a non-word char.
		if i > 0 && isWordRune(rs[i-1]) {
			continue
		}
		for _, tok := range scoreTokens {
			tr := []rune(tok)
			end := i + len(tr)
			if end > n {
				continue
			}
			matched := true
			for k := 0; k < len(tr); k++ {
				if rs[i+k] != tr[k] {
					matched = false
					break
				}
			}
			if !matched {
				continue
			}
			// A token ends with a word char, so a \b at the end requires the
			// following rune (if any) to be a non-word char.
			if end < n && isWordRune(rs[end]) {
				continue
			}
			return tok
		}
	}
	return "N"
}
