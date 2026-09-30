package pubshared

import (
	"math/rand"
	"regexp"
	"strings"
)

// prefixRe matches a line's structural prefix: an optional blockquote "> ", an
// optional bullet "- ", and an optional bold label ending ":** ". Ported from
// scramble.py PREFIX_RE.
var prefixRe = regexp.MustCompile(`^(\s*>\s*)?(-\s+)?(\*\*[^*]+?:\*\*\s*)?`)

// titleRe matches a markdown title line "# X". Ported from scramble.py TITLE_RE.
var titleRe = regexp.MustCompile(`^#\s+\S`)

// loremWords is the fixed neutral filler pool for vocab-swap mode, ported
// verbatim from scramble.py LOREM.
var loremWords = strings.Fields(
	"lorem ipsum dolor sit amet consectetur adipiscing elit sed do eiusmod tempor " +
		"incididunt ut labore et dolore magna aliqua enim ad minim veniam quis nostrud " +
		"exercitation ullamco laboris nisi aliquip ex ea commodo consequat duis aute " +
		"irure reprehenderit voluptate velit esse cillum fugiat nulla pariatur excepteur " +
		"sint occaecat cupidatat non proident sunt culpa qui officia deserunt mollit anim")

// ShuffleFunc has math/rand's Shuffle signature. It is injected so the shuffle
// mode is deterministic in a test and reproducible in production. This does NOT
// reproduce scramble.py's exact word order — Go's RNG is not byte-compatible with
// python's Mersenne Twister — but the structure (prefixes, blank/rule/title
// lines, word count per line, the class's own vocabulary) is preserved exactly.
type ShuffleFunc func(n int, swap func(i, j int))

// transformLine ports scramble.py _transform_line. In vocab-swap mode each word
// is replaced by the next lorem token (cursor shared across the whole document);
// in shuffle mode a line's words are shuffled when there is more than one.
func transformLine(line, mode string, shuffle ShuffleFunc, loremCursor *int) string {
	prefix := prefixRe.FindString(line)
	words := strings.Fields(line[len(prefix):])
	switch {
	case mode == "vocab-swap":
		swapped := make([]string, len(words))
		for i := range words {
			swapped[i] = loremWords[*loremCursor%len(loremWords)]
			*loremCursor++
		}
		words = swapped
	case len(words) > 1:
		shuffle(len(words), func(i, j int) { words[i], words[j] = words[j], words[i] })
	}
	return prefix + strings.Join(words, " ")
}

// gradedLine scrambles one content line at a graded strength. Each word is
// vocab-swapped to a lorem token with probability `strength`; the words left in
// place are shuffled among their own positions, so the shape, length, and prefix
// match transformLine. The lorem cursor advances across the document, matching
// vocab-swap mode. strength is assumed already clamped to [0, 1].
func gradedLine(line string, strength float64, rng *rand.Rand, loremCursor *int) string {
	prefix := prefixRe.FindString(line)
	words := strings.Fields(line[len(prefix):])
	if len(words) == 0 {
		return line
	}
	swap := make([]bool, len(words))
	kept := make([]string, 0, len(words))
	for i := range words {
		if rng.Float64() < strength {
			swap[i] = true
		} else {
			kept = append(kept, words[i])
		}
	}
	rng.Shuffle(len(kept), func(i, j int) { kept[i], kept[j] = kept[j], kept[i] })
	out := make([]string, len(words))
	keptIdx := 0
	for i := range words {
		if swap[i] {
			out[i] = loremWords[*loremCursor%len(loremWords)]
			*loremCursor++
		} else {
			out[i] = kept[keptIdx]
			keptIdx++
		}
	}
	return prefix + strings.Join(out, " ")
}

// TransformGraded scrambles a glyph at a graded strength between the weak shuffle
// (strength 0: all vocabulary kept and reordered) and the strong vocab-swap
// (strength 1: no vocabulary kept). Each content word is vocab-swapped to a lorem
// token with probability `strength`, and the kept words are shuffled among their
// positions, so shape, length, prefixes, and pass-through lines match Transform.
// strength is clamped to [0, 1]; rng makes the selection and shuffle reproducible.
// This is a new capability — Transform and its published shuffle/vocab-swap modes
// are unchanged, so committed scrambles still reproduce byte-for-byte.
func TransformGraded(srcText string, strength float64, neutralizeTitle bool, rng *rand.Rand) string {
	if strength < 0 {
		strength = 0
	}
	if strength > 1 {
		strength = 1
	}
	loremCursor := 0
	var out []string
	for _, line := range strings.Split(srcText, "\n") {
		stripped := pyStrip(line)
		lstripped := pyLStrip(line)
		switch {
		case stripped == "" || stripped == "---":
			out = append(out, line)
		case titleRe.MatchString(lstripped) || strings.HasPrefix(lstripped, "#"):
			if neutralizeTitle && titleRe.MatchString(lstripped) {
				out = append(out, "# glyph")
			} else {
				out = append(out, line)
			}
		default:
			out = append(out, gradedLine(line, strength, rng, &loremCursor))
		}
	}
	return strings.Join(out, "\n")
}

// Transform ports scramble.py transform. mode is "shuffle" (default) or
// "vocab-swap"; neutralizeTitle replaces a "# <slug>" title line with "# glyph".
// Blank lines, "---" rules, and non-title "#"-prefixed lines pass through
// verbatim. shuffle is used only in shuffle mode.
func Transform(srcText, mode string, neutralizeTitle bool, shuffle ShuffleFunc) string {
	loremCursor := 0
	var out []string
	for _, line := range strings.Split(srcText, "\n") {
		stripped := pyStrip(line)
		lstripped := pyLStrip(line)
		switch {
		case stripped == "" || stripped == "---":
			out = append(out, line)
		case titleRe.MatchString(lstripped) || strings.HasPrefix(lstripped, "#"):
			if neutralizeTitle && titleRe.MatchString(lstripped) {
				out = append(out, "# glyph")
			} else {
				out = append(out, line)
			}
		default:
			out = append(out, transformLine(line, mode, shuffle, &loremCursor))
		}
	}
	return strings.Join(out, "\n")
}
