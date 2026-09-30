package rater

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

// SliceLines reads a slice JSONL file into its lines. Each non-blank line is one
// {id, text, scenario?} object; blank lines are skipped.
func SliceLines(path string) ([]SliceLine, error) {
	data, err := os.ReadFile(path) //nolint:gosec // a study input path, chosen by the operator
	if err != nil {
		return nil, fmt.Errorf("read slice %s: %w", path, err)
	}
	var lines []SliceLine
	for i, raw := range strings.Split(string(data), "\n") {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		var l SliceLine
		if err := json.Unmarshal([]byte(raw), &l); err != nil {
			return nil, fmt.Errorf("slice %s line %d: %w", path, i+1, err)
		}
		lines = append(lines, l)
	}
	return lines, nil
}

// sliceIDs returns the ids a slice declares, in order.
func sliceIDs(lines []SliceLine) []string {
	ids := make([]string, len(lines))
	for i, l := range lines {
		ids[i] = l.ID
	}
	return ids
}

// IsComplete reports whether a canonical output file exists, parses, and its
// keys are exactly the expected ids. This is the content check that makes
// completion independent of whether a job process finished: a killed run
// resumes and re-runs only the slices without a valid result.
func IsComplete(path string, expected []string) bool {
	data, err := os.ReadFile(path) //nolint:gosec // a result path under the run's out dir
	if err != nil {
		return false
	}
	var got map[string]string
	if err := json.Unmarshal(data, &got); err != nil {
		return false
	}
	return sameIDSet(got, expected)
}

func sameIDSet(got map[string]string, expected []string) bool {
	want := make(map[string]struct{}, len(expected))
	for _, id := range expected {
		want[id] = struct{}{}
	}
	if len(got) != len(want) {
		return false
	}
	for id := range got {
		if _, ok := want[id]; !ok {
			return false
		}
	}
	return true
}

// Status is one slice's outcome.
type Status string

const (
	// StatusSkipComplete means a valid canonical result already existed.
	StatusSkipComplete Status = "skip-complete"
	// StatusScored means this run produced and promoted the canonical result.
	StatusScored Status = "scored"
	// StatusKeptExisting means a fresh canonical result appeared while this slice
	// scored (a straggler or a parallel run), so this result was dropped.
	StatusKeptExisting Status = "kept-existing"
)

// SliceOutcome pairs a slice's file name with its status.
type SliceOutcome struct {
	Slice  string
	Status Status
}

// Summary counts a run's outcomes.
type Summary struct {
	Scored          int
	AlreadyComplete int
	KeptExisting    int
	Outcomes        []SliceOutcome
}

// RunConfig drives a run. Rater and OutDir are required. Results land under
// OutDir/RaterID/<stem>.json. Jobs bounds concurrency (default 1, sequential).
// WriteProvenance writes a <result>.prov.json sidecar next to each result.
type RunConfig struct {
	Rater           *Rater
	OutDir          string
	Jobs            int
	WriteProvenance bool
}

// DiscoverSlices returns the slice files to rate: a single file when slice is
// set, otherwise every *.jsonl under slicesDir, sorted.
func DiscoverSlices(slicesDir, slice string) ([]string, error) {
	if slice != "" {
		return []string{slice}, nil
	}
	matches, err := filepath.Glob(filepath.Join(slicesDir, "*.jsonl"))
	if err != nil {
		return nil, fmt.Errorf("discover slices in %s: %w", slicesDir, err)
	}
	sort.Strings(matches)
	return matches, nil
}

// RunSlices rates each slice, skipping any with a valid canonical result, with
// at most cfg.Jobs slices in flight at once. It is resumable: a re-run only
// scores the slices still missing a result.
func RunSlices(ctx context.Context, cfg RunConfig, slicePaths []string) (Summary, error) {
	jobs := cfg.Jobs
	if jobs < 1 {
		jobs = 1
	}

	var (
		mu       sync.Mutex
		summary  Summary
		firstErr error
		wg       sync.WaitGroup
	)
	sem := make(chan struct{}, jobs)

	for _, path := range slicePaths {
		mu.Lock()
		stop := firstErr != nil
		mu.Unlock()
		if stop {
			break
		}
		wg.Add(1)
		sem <- struct{}{}
		go func(path string) {
			defer wg.Done()
			defer func() { <-sem }()
			outcome, err := runFile(ctx, cfg, path)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				if firstErr == nil {
					firstErr = err
				}
				return
			}
			switch outcome.Status {
			case StatusScored:
				summary.Scored++
			case StatusSkipComplete:
				summary.AlreadyComplete++
			case StatusKeptExisting:
				summary.KeptExisting++
			}
			summary.Outcomes = append(summary.Outcomes, outcome)
		}(path)
	}
	wg.Wait()
	if firstErr != nil {
		return Summary{}, firstErr
	}
	sort.Slice(summary.Outcomes, func(i, j int) bool {
		return summary.Outcomes[i].Slice < summary.Outcomes[j].Slice
	})
	return summary, nil
}

// runFile rates one slice and promotes its result create-only.
func runFile(ctx context.Context, cfg RunConfig, slicePath string) (SliceOutcome, error) {
	name := filepath.Base(slicePath)
	lines, err := SliceLines(slicePath)
	if err != nil {
		return SliceOutcome{Slice: name}, err
	}
	ids := sliceIDs(lines)
	stem := strings.TrimSuffix(name, filepath.Ext(name))
	canonical := filepath.Join(cfg.OutDir, cfg.Rater.raterID, stem+".json")

	if IsComplete(canonical, ids) {
		return SliceOutcome{Slice: name, Status: StatusSkipComplete}, nil
	}
	if err := os.MkdirAll(filepath.Dir(canonical), 0o750); err != nil {
		return SliceOutcome{Slice: name}, fmt.Errorf("make out dir: %w", err)
	}

	result, err := cfg.Rater.RateSlice(ctx, lines)
	if err != nil {
		return SliceOutcome{Slice: name}, err
	}

	encoded, err := json.MarshalIndent(result.Scores, "", "")
	if err != nil {
		return SliceOutcome{Slice: name}, fmt.Errorf("encode scores: %w", err)
	}
	promoted, err := writeCreateOnly(canonical, append(encoded, '\n'))
	if err != nil {
		return SliceOutcome{Slice: name}, err
	}
	if !promoted {
		return SliceOutcome{Slice: name, Status: StatusKeptExisting}, nil
	}
	if cfg.WriteProvenance && result.Provenance != nil {
		if err := writeProvenance(canonical+".prov.json", result.Provenance); err != nil {
			return SliceOutcome{Slice: name}, err
		}
	}
	return SliceOutcome{Slice: name, Status: StatusScored}, nil
}

// writeCreateOnly writes data to path only if path does not yet exist. It
// reports whether it created the file. An existing file is left untouched, so a
// late straggler can never overwrite a fresh canonical result.
func writeCreateOnly(path string, data []byte) (bool, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return false, nil
		}
		return false, fmt.Errorf("create %s: %w", path, err)
	}
	defer func() { _ = f.Close() }()
	if _, err := f.Write(data); err != nil {
		return false, fmt.Errorf("write %s: %w", path, err)
	}
	return true, nil
}

func writeProvenance(path string, prov *Provenance) error {
	encoded, err := json.MarshalIndent(prov, "", "  ")
	if err != nil {
		return fmt.Errorf("encode provenance: %w", err)
	}
	if err := os.WriteFile(path, append(encoded, '\n'), 0o600); err != nil {
		return fmt.Errorf("write provenance %s: %w", path, err)
	}
	return nil
}
