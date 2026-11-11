package taskcatalogue

import (
	"reflect"
	"testing"
)

func testLexicon() Lexicon {
	return NewLexicon([]Term{
		{Text: "alpha", Project: "alpha"},
		{Text: "alpha-tools", Project: "alpha-tools"},
		{Text: "Shared-Name", Project: "p1"},
		{Text: "shared-name", Project: "p2"}, // ambiguous: dropped
		{Text: "", Project: "x"},             // empty: dropped
		{Text: "orphan", Project: ""},        // no project: dropped
		{Text: "fix-the-big-thing", Project: "beta"},
		{Text: "self.compile", Project: "selfc"},
		{Text: "alpha", Project: "alpha"}, // duplicate, same project: kept once
	})
}

func TestNewLexiconDropsAmbiguousAndEmpty(t *testing.T) {
	lex := testLexicon()
	// alpha, alpha-tools, "alpha tools", fix-the-big-thing, "fix the big thing", self.compile
	if lex.Len() != 6 {
		t.Fatalf("Len = %d, want 6: %+v", lex.Len(), lex.terms)
	}
	if len(lex.Mentions("shared-name and shared name")) != 0 {
		t.Error("an ambiguous term (and its spaced variant) must not match")
	}
}

func TestNewLexiconIgnore(t *testing.T) {
	lex := NewLexicon([]Term{{Text: "vault", Project: "vault"}, {Text: "alpha", Project: "alpha"}}, " Vault ")
	if lex.Len() != 1 || len(lex.Mentions("file it to the vault")) != 0 {
		t.Errorf("ignored term still matches: %+v", lex.terms)
	}
}

func TestMentions(t *testing.T) {
	lex := testLexicon()
	cases := map[string]map[string]int{
		"work on alpha-tools please":      {"alpha-tools": 1},
		"Alpha. then ALPHA tools":         {"alpha": 1, "alpha-tools": 1},
		"alphabet soup":                   {},
		"see fix-the-big-thing in alpha":  {"beta": 1, "alpha": 1},
		"the self.compile repo":           {"selfc": 1},
		"x-alpha and alpha_2 and alphaé":  {},
		"alpha":                           {"alpha": 1},
		"fix the big thing, then alpha  ": {"beta": 1, "alpha": 1},
	}
	for text, want := range cases {
		got := lex.Mentions(text)
		if len(want) == 0 && len(got) == 0 {
			continue
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("%q: got %v, want %v", text, got, want)
		}
	}
}

func TestAttribute(t *testing.T) {
	lex := testLexicon()
	prompts := []Prompt{
		{ID: 0, Session: "a", Text: "alpha needs work"},
		{ID: 1, Session: "a", Text: "yes please"},
		{ID: 2, Session: "b", Text: "yes please"},
		{ID: 3, Session: "c", Text: "hello"},
		{ID: 4, Session: "d", Text: "alpha-tools and alpha"}, // tie: alphabetical
	}
	got := Attribute(prompts, lex, map[string]string{"b": "gamma", "a": "zeta"})
	want := map[int]Attribution{
		0: {Project: "alpha", Source: SourceMention},
		1: {Project: "zeta", Source: SourceSessionToolkit},
		2: {Project: "gamma", Source: SourceSessionToolkit},
		3: {Source: SourceNone},
		4: {Project: "alpha", Source: SourceMention},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v\nwant %v", got, want)
	}
	// Without a toolkit scope, a session's other mentions decide.
	got = Attribute(prompts[:2], lex, nil)
	if got[1] != (Attribution{Project: "alpha", Source: SourceSessionMention}) {
		t.Errorf("session mention: %v", got[1])
	}
}
