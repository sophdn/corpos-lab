package taskcatalogue

import (
	"sort"
	"strings"
)

// Term maps a name that can appear in a prompt to the project it names: a
// project id, a retired alias, or the slug of a chain, bug, suggestion or
// proposal that belongs to the project.
type Term struct {
	Text    string `json:"text"`
	Project string `json:"project"`
}

// Lexicon is the set of terms used to attribute a prompt to a project.
type Lexicon struct {
	terms []Term // lowercased, deduplicated, longest first
}

// NewLexicon builds a lexicon. Terms are matched case-insensitively. A term
// that maps to two different projects is ambiguous and dropped, because a
// guess between them would be an assertion the data does not support. Terms
// named in ignore are dropped the same way. A
// hyphenated term also matches with spaces in place of the hyphens, so
// "campaign-settings" matches "campaign settings".
func NewLexicon(terms []Term, ignore ...string) Lexicon {
	byText := map[string]string{}
	ambiguous := map[string]bool{}
	for _, w := range ignore {
		ambiguous[strings.ToLower(strings.TrimSpace(w))] = true
	}
	add := func(text, project string) {
		text = strings.ToLower(strings.TrimSpace(text))
		if text == "" || project == "" {
			return
		}
		if prev, ok := byText[text]; ok && prev != project {
			ambiguous[text] = true
			return
		}
		byText[text] = project
	}
	for _, t := range terms {
		add(t.Text, t.Project)
		if strings.Contains(t.Text, "-") {
			add(strings.ReplaceAll(t.Text, "-", " "), t.Project)
		}
	}
	out := make([]Term, 0, len(byText))
	for text, project := range byText {
		if !ambiguous[text] {
			out = append(out, Term{Text: text, Project: project})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if len(out[i].Text) != len(out[j].Text) {
			return len(out[i].Text) > len(out[j].Text)
		}
		return out[i].Text < out[j].Text
	})
	return Lexicon{terms: out}
}

// Len reports how many terms the lexicon holds after deduplication.
func (l Lexicon) Len() int { return len(l.terms) }

// Mentions counts, per project, the terms found in text. A match must sit on
// word boundaries, where letters, digits, '-', '_' and non-ASCII bytes count as
// word characters, so "corpos" does not match inside "the private toolkit service". Longer terms
// are matched first and their span is consumed, so one name is not counted
// twice through a shorter term inside it.
func (l Lexicon) Mentions(text string) map[string]int {
	lower := []byte(strings.ToLower(text))
	used := make([]bool, len(lower))
	out := map[string]int{}
	for _, t := range l.terms {
		needle := []byte(t.Text)
		for from := 0; from+len(needle) <= len(lower); {
			idx := indexFrom(lower, needle, from)
			if idx < 0 {
				break
			}
			end := idx + len(needle)
			if onBoundary(lower, idx, end) && !anyUsed(used, idx, end) {
				out[t.Project]++
				for k := idx; k < end; k++ {
					used[k] = true
				}
			}
			from = idx + 1
		}
	}
	return out
}

func indexFrom(hay, needle []byte, from int) int {
	i := strings.Index(string(hay[from:]), string(needle))
	if i < 0 {
		return -1
	}
	return from + i
}

func onBoundary(s []byte, start, end int) bool {
	return (start == 0 || !isWordByte(s[start-1])) && (end == len(s) || !isWordByte(s[end]))
}

// isWordByte reports whether c continues a word. '.' does not, so "corpos." at
// a sentence end still matches; a term that contains a '.' ("self.compile")
// matches as a whole because the needle carries it.
func isWordByte(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '-' || c == '_' || c >= 0x80
}

func anyUsed(used []bool, start, end int) bool {
	for k := start; k < end; k++ {
		if used[k] {
			return true
		}
	}
	return false
}

// Attribution says which project a prompt belongs to and how that was decided.
type Attribution struct {
	Project string `json:"project"`
	// Source is one of the Source* constants.
	Source string `json:"source"`
}

// Attribution sources, in the order they are tried.
const (
	// SourceMention: the prompt names the project.
	SourceMention = "mention"
	// SourceSessionToolkit: the prompt's session made toolkit calls scoped to
	// the project (the observed record of where the session's work went).
	SourceSessionToolkit = "session-toolkit"
	// SourceSessionMention: other prompts in the same session name the project.
	SourceSessionMention = "session-mention"
	// SourceNone: nothing ties the prompt to a project.
	SourceNone = "none"
)

// Attribute assigns each prompt a project. sessionProjects maps a session id
// to the project its toolkit calls named most often. Ties between projects
// break alphabetically so the result is deterministic.
func Attribute(prompts []Prompt, lex Lexicon, sessionProjects map[string]string) map[int]Attribution {
	perPrompt := make(map[int]map[string]int, len(prompts))
	perSession := map[string]map[string]int{}
	for _, p := range prompts {
		m := lex.Mentions(p.Text)
		perPrompt[p.ID] = m
		if perSession[p.Session] == nil {
			perSession[p.Session] = map[string]int{}
		}
		for proj, n := range m {
			perSession[p.Session][proj] += n
		}
	}
	out := make(map[int]Attribution, len(prompts))
	for _, p := range prompts {
		switch {
		case len(perPrompt[p.ID]) > 0:
			out[p.ID] = Attribution{Project: top(perPrompt[p.ID]), Source: SourceMention}
		case sessionProjects[p.Session] != "":
			out[p.ID] = Attribution{Project: sessionProjects[p.Session], Source: SourceSessionToolkit}
		case len(perSession[p.Session]) > 0:
			out[p.ID] = Attribution{Project: top(perSession[p.Session]), Source: SourceSessionMention}
		default:
			out[p.ID] = Attribution{Source: SourceNone}
		}
	}
	return out
}

// top returns the key with the highest count, ties broken alphabetically.
func top(counts map[string]int) string {
	best, bestN := "", -1
	for k, n := range counts {
		if n > bestN || n == bestN && k < best {
			best, bestN = k, n
		}
	}
	return best
}
