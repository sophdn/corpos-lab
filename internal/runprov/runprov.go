// Package runprov summarises the provenance of every run under a study dir,
// one row per model: the llama.cpp builds, n_ctx, sampler chains, seeds and
// image digests the runs recorded, and how many rows were truncated or ran on a
// model other than the one declared. Paper readers used to open 60KB
// run-record.json files, or script over dozens of them, to fill a provenance
// table by hand.
package runprov

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"

	"corpos-lab/internal/assay"
	"corpos-lab/internal/control"
)

// Model is the provenance of every run of one model under a study dir.
type Model struct {
	Model     string   `json:"model"`
	Runs      int      `json:"runs"`
	Builds    []string `json:"builds"`
	NCtx      []int    `json:"n_ctx"`
	Samplers  []string `json:"samplers"`
	Seeds     []int    `json:"seeds"`
	Digests   []string `json:"image_digests"`
	Rows      int      `json:"rows"`
	Truncated int      `json:"truncated_rows"`
	// Mismatched counts runs whose server reported a model other than the one
	// the study declared.
	Mismatched int `json:"mismatched_runs"`
	// Unread counts runs whose /props readback failed, so their build and n_ctx
	// are unknown rather than absent.
	Unread int `json:"props_unread_runs"`
}

// Load walks dir for run-record.json files and groups them by model.
func Load(dir string) ([]Model, error) {
	byModel := map[string]*acc{}
	found := 0
	err := filepath.WalkDir(dir, func(path string, e fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if e.IsDir() || e.Name() != "run-record.json" {
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		var run control.StudyRun
		if err := json.Unmarshal(raw, &run); err != nil {
			return fmt.Errorf("runprov: parse %s: %w", path, err)
		}
		found++
		if run.Results == nil {
			return nil // a failed run records no results to summarise
		}
		name := run.Results.ModelID
		if a := byModel[name]; a == nil {
			byModel[name] = newAcc(name)
		}
		byModel[name].add(run)
		return nil
	})
	if err != nil {
		return nil, err
	}
	if found == 0 {
		return nil, fmt.Errorf("runprov: no run-record.json under %s", dir)
	}
	out := make([]Model, 0, len(byModel))
	for _, a := range byModel {
		out = append(out, a.model())
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Model < out[j].Model })
	return out, nil
}

type acc struct {
	m        Model
	builds   map[string]bool
	nctx     map[int]bool
	samplers map[string]bool
	seeds    map[int]bool
	digests  map[string]bool
}

func newAcc(name string) *acc {
	return &acc{m: Model{Model: name}, builds: map[string]bool{}, nctx: map[int]bool{},
		samplers: map[string]bool{}, seeds: map[int]bool{}, digests: map[string]bool{}}
}

func (a *acc) add(run control.StudyRun) {
	r := run.Results
	a.m.Runs++
	if r.ServerReadbackError != "" {
		a.m.Unread++
	} else {
		if r.Server.BuildInfo != "" {
			a.builds[r.Server.BuildInfo] = true
		}
		if r.Server.NCtx > 0 {
			a.nctx[r.Server.NCtx] = true
		}
	}
	if r.ModelMismatch != "" {
		a.m.Mismatched++
	}
	a.samplers[SamplerChain(r.Sampler)] = true
	for _, s := range r.Sampler.Seeds {
		a.seeds[s] = true
	}
	if run.ImageDigest != "" {
		a.digests[run.ImageDigest] = true
	}
	for _, row := range r.Rows {
		a.m.Rows++
		// Per-row builds count too: a server upgrade mid-study shows here.
		if row.Observed.BuildInfo != "" {
			a.builds[row.Observed.BuildInfo] = true
		}
		if row.Observed.Truncated {
			a.m.Truncated++
		}
	}
}

func (a *acc) model() Model {
	m := a.m
	m.Builds, m.Samplers, m.Digests = keys(a.builds), keys(a.samplers), keys(a.digests)
	m.NCtx, m.Seeds = intKeys(a.nctx), intKeys(a.seeds)
	return m
}

func keys(m map[string]bool) []string {
	out := []string{}
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func intKeys(m map[int]bool) []int {
	out := []int{}
	for k := range m {
		out = append(out, k)
	}
	sort.Ints(out)
	return out
}

// SamplerChain renders the full sampler chain a run sent, in chain order, as
// one comparable string. Seeds are reported separately.
func SamplerChain(s assay.Sampling) string {
	f := func(v float64) string { return strconv.FormatFloat(v, 'g', -1, 64) }
	return strings.Join([]string{
		"temp=" + f(s.Temperature), "max_tokens=" + strconv.Itoa(s.MaxTokens),
		"top_n_sigma=" + f(s.TopNSigma), "top_k=" + strconv.Itoa(s.TopK),
		"typical_p=" + f(s.TypicalP), "top_p=" + f(s.TopP), "min_p=" + f(s.MinP),
		"repeat_penalty=" + f(s.RepeatPenalty), "repeat_last_n=" + strconv.Itoa(s.RepeatLastN),
		"presence=" + f(s.PresencePenalty), "frequency=" + f(s.FrequencyPenalty),
		"xtc=" + f(s.XTCProbability), "dry=" + f(s.DryMultiplier),
	}, " ")
}

// WriteJSON writes the models as an indented JSON array.
func WriteJSON(w io.Writer, models []Model) error {
	b, err := json.MarshalIndent(models, "", "  ")
	if err != nil {
		return err
	}
	_, err = w.Write(append(b, '\n'))
	return err
}

// WriteTable writes one block per model.
func WriteTable(w io.Writer, models []Model) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	for _, m := range models {
		fmt.Fprintf(tw, "model\t%s\n", m.Model)
		fmt.Fprintf(tw, "runs / rows\t%d / %d\n", m.Runs, m.Rows)
		fmt.Fprintf(tw, "build\t%s\n", orNone(m.Builds))
		fmt.Fprintf(tw, "n_ctx\t%s\n", orNone(ints(m.NCtx)))
		fmt.Fprintf(tw, "sampler\t%s\n", strings.Join(m.Samplers, "\n\t"))
		fmt.Fprintf(tw, "seeds\t%s\n", orNone(ints(m.Seeds)))
		fmt.Fprintf(tw, "image\t%s\n", orNone(m.Digests))
		fmt.Fprintf(tw, "truncated rows\t%d\n", m.Truncated)
		fmt.Fprintf(tw, "model mismatch runs\t%d\n", m.Mismatched)
		fmt.Fprintf(tw, "props unread runs\t%d\n\n", m.Unread)
	}
	return tw.Flush()
}

func ints(v []int) []string {
	out := make([]string, len(v))
	for i, n := range v {
		out[i] = strconv.Itoa(n)
	}
	return out
}

func orNone(v []string) string {
	if len(v) == 0 {
		return "(none recorded)"
	}
	return strings.Join(v, ", ")
}
