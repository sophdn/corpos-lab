// Package loopcells reads the per-cell outcomes of an agentic-loop run dir, so a
// reader of the run does not write a throwaway scorer for each battery. It
// reports what each cell did — which files the subject opened, which edits it
// claimed and whether each one ran, what changed in the sandbox, how it ended —
// and makes no fire/not-fire judgment: reading the cells is the method's job,
// not the tool's (corpus/private/glyph-model/specimen-method/METHOD.md).
//
// It reads the transcript the loop writes per cell (responses/<cell>.txt), so it
// works on runs recorded before the loop kept structured per-turn data. Where the
// results row carries the parse record (parse_outcomes, lost_calls) it uses that;
// otherwise it estimates lost calls from the transcript and says so.
package loopcells

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"

	"corpos-lab/internal/agentloop"
)

// Cell is one loop cell's outcome.
type Cell struct {
	Condition string `json:"condition"`
	Run       int    `json:"run"`
	// Seed is the sampler seed for this run, nil when the run fell outside the
	// study's seed list (an unseeded run).
	Seed     *int   `json:"seed"`
	Terminal string `json:"terminal"`
	Turns    int    `json:"turns"`
	// Opened lists the read_file paths that ran, in order.
	Opened []string `json:"opened"`
	// Edits lists every edit the subject claimed and whether it ran.
	Edits          []Edit `json:"edits"`
	TruncatedTurns int    `json:"truncated_turns"`
	// LostCalls counts calls the parse lost. LostSource is "results" when the
	// row recorded it, "transcript" when estimated from the transcript (older
	// runs), "" for a cell error.
	LostCalls    int    `json:"lost_calls"`
	LostSource   string `json:"lost_source"`
	IgnoredCalls *int   `json:"ignored_calls,omitempty"`
	// CollapsedCalls counts repeated identical calls within a turn that the
	// loop did not run again (from the transcript's [collapsed: …] notes).
	CollapsedCalls int      `json:"collapsed_calls"`
	ParseOutcomes  []string `json:"parse_outcomes,omitempty"`
	// Changed lists the sandbox files the cell added, modified or removed. It is
	// nil when the starting sandbox is not available.
	Changed []Change `json:"changed"`
	// Final is the subject's FINAL line, else the transcript's last line.
	Final string `json:"final"`
}

// Edit is one claimed edit_file call.
type Edit struct {
	Turn     int    `json:"turn"`
	Path     string `json:"path"`
	Executed bool   `json:"executed"`
	// Observation is what the harness returned for it, empty when it never ran.
	Observation string `json:"observation,omitempty"`
}

// Change is one sandbox file a cell changed.
type Change struct {
	Path string `json:"path"`
	Kind string `json:"kind"` // added, modified, removed
}

type resultsFile struct {
	Sampler struct {
		Seeds []int `json:"seeds"`
	} `json:"sampler"`
	Rows []struct {
		Condition string `json:"condition"`
		Run       int    `json:"run"`
		Rationale string `json:"rationale"`
		Observed  struct {
			ParseOutcomes []string `json:"parse_outcomes"`
			LostCalls     *int     `json:"lost_calls"`
			IgnoredCalls  *int     `json:"ignored_calls"`
		} `json:"observed"`
	} `json:"rows"`
}

// Load reads every cell of the run at dir. dir is the run dir (holding in/ and
// out/) or its out/ dir. A missing results.json is an error naming the path it
// expected.
func Load(dir string) ([]Cell, error) {
	outDir := filepath.Join(dir, "out")
	if _, err := os.Stat(filepath.Join(dir, "results.json")); err == nil {
		outDir = dir
	}
	resultsPath := filepath.Join(outDir, "results.json")
	raw, err := os.ReadFile(resultsPath)
	if err != nil {
		return nil, fmt.Errorf("loopcells: no results.json at %s (pass a loop run dir or its out/ dir): %w", resultsPath, err)
	}
	var res resultsFile
	if err := json.Unmarshal(raw, &res); err != nil {
		return nil, fmt.Errorf("loopcells: parse %s: %w", resultsPath, err)
	}
	start, err := startSandbox(filepath.Join(filepath.Dir(outDir), "in"))
	if err != nil {
		return nil, err
	}
	cells := make([]Cell, 0, len(res.Rows))
	for _, row := range res.Rows {
		name := fmt.Sprintf("%s_%d.txt", row.Condition, row.Run)
		tr, err := os.ReadFile(filepath.Join(outDir, "responses", name))
		if err != nil {
			return nil, fmt.Errorf("loopcells: transcript for %s/%d: %w", row.Condition, row.Run, err)
		}
		c := Cell{Condition: row.Condition, Run: row.Run, Opened: []string{}, Edits: []Edit{}}
		if row.Run >= 1 && row.Run <= len(res.Sampler.Seeds) {
			seed := res.Sampler.Seeds[row.Run-1]
			c.Seed = &seed
		}
		if strings.Contains(row.Rationale, ":cell-error:") {
			c.Terminal = "cell-error"
			c.Final = strings.TrimSpace(string(tr))
			cells = append(cells, c)
			continue
		}
		c.Terminal = field(row.Rationale, "terminal")
		c.Turns, _ = strconv.Atoi(field(row.Rationale, "turns"))
		readTranscript(&c, string(tr))
		if row.Observed.LostCalls != nil {
			c.LostCalls, c.LostSource = *row.Observed.LostCalls, "results"
		}
		c.IgnoredCalls = row.Observed.IgnoredCalls
		c.ParseOutcomes = row.Observed.ParseOutcomes
		if start != nil {
			end, err := readSandbox(filepath.Join(outDir, "sandboxes", name))
			if err != nil {
				return nil, err
			}
			c.Changed = diff(start, end)
		}
		cells = append(cells, c)
	}
	return cells, nil
}

// field returns the value of "key=value" in a colon-separated rationale.
func field(rationale, key string) string {
	for _, part := range strings.Split(rationale, ":") {
		if v, ok := strings.CutPrefix(part, key+"="); ok {
			return v
		}
	}
	return ""
}

// startSandbox reads the sandbox the cells started from, named by the loop
// section of in/study.json. A run dir with no in/ yields nil, not an error: the
// cells are still readable, only the change list is unknown.
func startSandbox(inDir string) (map[string]string, error) {
	raw, err := os.ReadFile(filepath.Join(inDir, "study.json"))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("loopcells: %w", err)
	}
	var spec struct {
		Loop struct {
			Sandbox string `json:"sandbox"`
		} `json:"loop"`
	}
	if err := json.Unmarshal(raw, &spec); err != nil || spec.Loop.Sandbox == "" {
		return nil, fmt.Errorf("loopcells: %s names no loop sandbox (is this a loop run?)", filepath.Join(inDir, "study.json"))
	}
	return readSandbox(filepath.Join(inDir, spec.Loop.Sandbox))
}

func readSandbox(path string) (map[string]string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("loopcells: %w", err)
	}
	var m map[string]string
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, fmt.Errorf("loopcells: parse sandbox %s: %w", path, err)
	}
	return m, nil
}

func diff(start, end map[string]string) []Change {
	changes := []Change{}
	for p, v := range end {
		if old, ok := start[p]; !ok {
			changes = append(changes, Change{Path: p, Kind: "added"})
		} else if old != v {
			changes = append(changes, Change{Path: p, Kind: "modified"})
		}
	}
	for p := range start {
		if _, ok := end[p]; !ok {
			changes = append(changes, Change{Path: p, Kind: "removed"})
		}
	}
	sort.Slice(changes, func(i, j int) bool { return changes[i].Path < changes[j].Path })
	return changes
}

// turn is one parsed transcript turn.
type turn struct {
	text  string
	calls []call
}

type call struct {
	label, observation string
}

// readTranscript fills the transcript-derived fields of c. The transcript is the
// loop's Transcript() rendering: "[turn N] text", then per executed call an
// "EXECUTED: label" line (a label can span lines in older runs) and an
// "OBSERVATION: …" block, then optional loop-end notes.
func readTranscript(c *Cell, tr string) {
	var turns []*turn
	var cur *turn
	var section *string // the text block lines are appended to
	lastLine := ""
	for _, line := range strings.Split(tr, "\n") {
		if strings.TrimSpace(line) != "" {
			lastLine = strings.TrimSpace(line)
		}
		switch {
		case strings.HasPrefix(line, "[turn ") && strings.Contains(line, " truncated at the call_tokens cap"):
			c.TruncatedTurns++
			section = nil
		case strings.HasPrefix(line, "[turn ") && strings.Contains(line, "] "):
			_, text, _ := strings.Cut(line, "] ")
			cur = &turn{text: text}
			turns = append(turns, cur)
			section = &cur.text
		case strings.HasPrefix(line, "[collapsed: "):
			n, _ := strconv.Atoi(strings.Fields(strings.TrimPrefix(line, "[collapsed: "))[0])
			c.CollapsedCalls += n
			section = nil
		case strings.HasPrefix(line, "[loop ended") || strings.HasPrefix(line, "[WARNING"):
			section = nil
		case cur != nil && strings.HasPrefix(line, "EXECUTED: "):
			cur.calls = append(cur.calls, call{label: strings.TrimPrefix(line, "EXECUTED: ")})
			section = &cur.calls[len(cur.calls)-1].label
		case cur != nil && len(cur.calls) > 0 && strings.HasPrefix(line, "OBSERVATION: "):
			k := &cur.calls[len(cur.calls)-1]
			k.observation = strings.TrimPrefix(line, "OBSERVATION: ")
			section = &k.observation
		case section != nil:
			*section += "\n" + line
		}
	}
	c.Final = lastLine
	lost := 0
	for i, t := range turns {
		for _, k := range t.calls {
			if path, ok := strings.CutPrefix(k.label, "read_file "); ok {
				c.Opened = append(c.Opened, strings.TrimSpace(path))
			}
			if lostCall(k) {
				lost++
			}
		}
		c.Edits = append(c.Edits, edits(i+1, t)...)
		if strings.HasPrefix(strings.TrimSpace(lastLineOf(t.text)), "FINAL") && i == len(turns)-1 && len(t.calls) == 0 {
			c.Final = strings.TrimSpace(lastLineOf(t.text))
		}
	}
	c.LostCalls, c.LostSource = lost, "transcript"
}

// edits pairs the turn's claimed edits with the edit calls that ran, by path in
// order. A claimed edit with no matching executed call never ran.
func edits(n int, t *turn) []Edit {
	var ran []call
	for _, k := range t.calls {
		if strings.HasPrefix(k.label, "edit_file") {
			ran = append(ran, k)
		}
	}
	used := make([]bool, len(ran))
	var out []Edit
	for _, path := range agentloop.ClaimedEdits(t.text) {
		e := Edit{Turn: n, Path: path}
		for j, k := range ran {
			if !used[j] && strings.TrimSpace(strings.TrimPrefix(k.label, "edit_file")) == path {
				used[j] = true
				e.Observation = k.observation
				e.Executed = strings.HasPrefix(k.observation, "wrote ")
				break
			}
		}
		out = append(out, e)
	}
	return out
}

// lostCall estimates, from a transcript line, whether a call was lost to the
// parse: an unknown tool, an argument still holding tag text, a read with no
// path, or an edit the sandbox refused.
func lostCall(k call) bool {
	label := strings.TrimSpace(k.label)
	switch {
	case strings.HasPrefix(label, "unknown:"):
		return true
	case strings.Contains(label, "<") || strings.Contains(label, "\n"):
		return true
	case label == "read_file" || label == "run_query" || label == "edit_file":
		return true
	case strings.HasPrefix(k.observation, "error: edit_file"), strings.Contains(k.observation, "(0 bytes)"):
		return true
	}
	return false
}

func lastLineOf(s string) string {
	s = strings.TrimRight(s, "\n ")
	if i := strings.LastIndexByte(s, '\n'); i >= 0 {
		return s[i+1:]
	}
	return s
}

// WriteJSON writes the cells as an indented JSON array.
func WriteJSON(w io.Writer, cells []Cell) error {
	b, err := json.MarshalIndent(cells, "", "  ")
	if err != nil {
		return err
	}
	_, err = w.Write(append(b, '\n'))
	return err
}

// WriteTable writes one line per cell for reading at a glance.
func WriteTable(w io.Writer, cells []Cell) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "CELL\tSEED\tTERMINAL\tTURNS\tLOST\tREPEATS\tEDITS RAN/CLAIMED\tTRUNC\tOPENED\tCHANGED\tFINAL")
	for _, c := range cells {
		seed := "-"
		if c.Seed != nil {
			seed = strconv.Itoa(*c.Seed)
		}
		ran := 0
		for _, e := range c.Edits {
			if e.Executed {
				ran++
			}
		}
		changed := "?"
		if c.Changed != nil {
			var parts []string
			for _, ch := range c.Changed {
				parts = append(parts, ch.Kind[:1]+":"+ch.Path)
			}
			changed = strings.Join(parts, ",")
		}
		fmt.Fprintf(tw, "%s_%d\t%s\t%s\t%d\t%d\t%d\t%d/%d\t%d\t%s\t%s\t%s\n",
			c.Condition, c.Run, seed, c.Terminal, c.Turns, c.LostCalls, c.CollapsedCalls, ran, len(c.Edits),
			c.TruncatedTurns, strings.Join(c.Opened, ","), changed, shorten(c.Final, 60))
	}
	return tw.Flush()
}

// shorten flattens s to one line of at most n runes.
func shorten(s string, n int) string {
	s = strings.Join(strings.Fields(s), " ")
	if r := []rune(s); len(r) > n {
		return string(r[:n-1]) + "…"
	}
	return s
}
