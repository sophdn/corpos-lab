// Package taskcatalogue builds a frequency-weighted catalogue of the task kinds
// a principal actually asks a coding agent for, from the agent harness's prompt
// log. It parses the log, attributes each prompt to a project, labels each
// prompt with a role (does it open new work, or continue work already under
// way?) and a task kind from a closed taxonomy, and aggregates the labels into
// a kind table whose head covers a stated share of the principal's openers.
//
// The catalogue is the input to a replay sweep that drives the same task mix
// through corpos to find its gaps (corpos-lab chain popular-task-catalogue).
// The prompts are private: this package only ever reads and writes the paths it
// is given, and the committed tests run on synthetic fixtures alone.
package taskcatalogue

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"
)

// Text caps. A prompt's own text keeps more than its pastes, because the ask
// lives in what was typed; a paste is usually a log or a document the ask is
// about. The previous prompt is context for the role judgement only.
const (
	MaxTextRunes  = 1200
	MaxPasteRunes = 600
	MaxPrevRunes  = 300
)

// Prompt is one typed prompt from the log, with the context the labeller needs.
type Prompt struct {
	// ID is the prompt's zero-based line index in the log: stable for an
	// append-only log, so labels survive a re-run on a longer log.
	ID      int       `json:"id"`
	Session string    `json:"session"`
	Time    time.Time `json:"time"`
	// CWD is the directory the harness ran in when the prompt was typed.
	CWD string `json:"cwd"`
	// Text is the typed text with pasted content appended, capped.
	Text string `json:"text"`
	// Prev is the previous prompt in the same session, capped; empty for a
	// session's first prompt.
	Prev string `json:"prev,omitempty"`
	// Slash marks a harness command ("/mcp reconnect", "/self-reflect").
	Slash bool `json:"slash,omitempty"`
}

type historyLine struct {
	Display        string                  `json:"display"`
	PastedContents map[string]pastedRecord `json:"pastedContents"`
	Timestamp      int64                   `json:"timestamp"`
	Project        string                  `json:"project"`
	SessionID      string                  `json:"sessionId"`
}

type pastedRecord struct {
	ID          int    `json:"id"`
	Type        string `json:"type"`
	Content     string `json:"content"`
	ContentHash string `json:"contentHash"`
}

// PasteLookup resolves a paste stored by reference (contentHash) to its text.
// The harness keeps those in a cache it prunes, so a miss is normal.
type PasteLookup func(hash string) (string, bool)

// pasteGone stands in for a paste whose text the harness no longer holds, so
// the labeller knows something was pasted rather than guessing at a bare
// placeholder.
const pasteGone = "(pasted content not retained)"

// ParseHistory reads the harness prompt log: one JSON object per line with
// display, pastedContents, timestamp (Unix ms), project (the cwd) and
// sessionId. lookup, when non-nil, resolves pastes stored by hash. Blank lines are skipped but still consume an ID, so IDs stay equal
// to line indexes. A malformed line is an error naming its line: the log is the
// population, and silently dropping a line would bias every count after it.
func ParseHistory(r io.Reader, lookup PasteLookup) ([]Prompt, error) {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)
	out := []Prompt{}
	lastInSession := map[string]string{}
	for line := 0; sc.Scan(); line++ {
		raw := strings.TrimSpace(sc.Text())
		if raw == "" {
			continue
		}
		var h historyLine
		if err := json.Unmarshal([]byte(raw), &h); err != nil {
			return nil, fmt.Errorf("history line %d: %w", line, err)
		}
		text := composeText(h, lookup)
		p := Prompt{
			ID:      line,
			Session: h.SessionID,
			Time:    time.UnixMilli(h.Timestamp).UTC(),
			CWD:     h.Project,
			Text:    text,
			Prev:    capRunes(lastInSession[h.SessionID], MaxPrevRunes),
			Slash:   strings.HasPrefix(strings.TrimSpace(h.Display), "/"),
		}
		lastInSession[h.SessionID] = text
		out = append(out, p)
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("read history: %w", err)
	}
	return out, nil
}

// composeText joins the typed display with each pasted text block, in paste-id
// order so the result is deterministic.
func composeText(h historyLine, lookup PasteLookup) string {
	var b strings.Builder
	b.WriteString(capRunes(h.Display, MaxTextRunes))
	keys := make([]string, 0, len(h.PastedContents))
	for k := range h.PastedContents {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		pc := h.PastedContents[k]
		if pc.Type != "" && pc.Type != "text" {
			continue
		}
		content := pc.Content
		if content == "" && pc.ContentHash != "" {
			content = pasteGone
			if lookup != nil {
				if c, ok := lookup(pc.ContentHash); ok {
					content = c
				}
			}
		}
		if strings.TrimSpace(content) == "" {
			continue
		}
		fmt.Fprintf(&b, "\n[pasted #%s] %s", k, capRunes(content, MaxPasteRunes))
	}
	return b.String()
}

// capRunes truncates s to n runes, marking a cut with an ellipsis.
func capRunes(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}
