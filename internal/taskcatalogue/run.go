package taskcatalogue

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"
)

// ReadLabels reads a labels JSONL stream. When a prompt was labelled more than
// once (a re-run of failures), the last record wins: the file is append-only
// and later is newer.
func ReadLabels(r io.Reader) (map[int]PromptLabel, error) {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	out := map[int]PromptLabel{}
	for line := 1; sc.Scan(); line++ {
		if len(sc.Bytes()) == 0 {
			continue
		}
		var l PromptLabel
		if err := json.Unmarshal(sc.Bytes(), &l); err != nil {
			return nil, fmt.Errorf("labels line %d: %w", line, err)
		}
		out[l.ID] = l
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("read labels: %w", err)
	}
	return out, nil
}

// Pending returns the prompts that still need a label: those with no record,
// and, when retryFailed is set, those whose last record is not ok.
func Pending(prompts []Prompt, done map[int]PromptLabel, retryFailed bool) []Prompt {
	var out []Prompt
	for _, p := range prompts {
		l, ok := done[p.ID]
		if !ok || retryFailed && l.Status != StatusOK {
			out = append(out, p)
		}
	}
	return out
}

// LabelAll labels prompts with up to workers concurrent calls and appends each
// result to w as one JSON line, as soon as it is ready, so an interrupted run
// keeps every label it finished. progress, when non-nil, is called after each
// write with the number written so far. It stops early when ctx is cancelled;
// the labels already written stay valid and a re-run picks up the rest.
func LabelAll(ctx context.Context, c Classifier, prompts []Prompt, workers int, w io.Writer, progress func(done int)) (int, error) {
	if workers < 1 {
		workers = 1
	}
	jobs := make(chan Prompt)
	results := make(chan PromptLabel)
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for p := range jobs {
				results <- c.Classify(ctx, p)
			}
		}()
	}
	go func() {
		defer close(jobs)
		for _, p := range prompts {
			select {
			case <-ctx.Done():
				return
			case jobs <- p:
			}
		}
	}()
	go func() {
		wg.Wait()
		close(results)
	}()

	enc := json.NewEncoder(w)
	written := 0
	var writeErr error
	for l := range results {
		if writeErr != nil {
			continue // drain so the workers can exit
		}
		if ctx.Err() != nil && l.Status == StatusError {
			continue // a call cut off by cancellation is not a finding
		}
		if err := enc.Encode(l); err != nil {
			writeErr = fmt.Errorf("write label %d: %w", l.ID, err)
			continue
		}
		written++
		if progress != nil {
			progress(written)
		}
	}
	if writeErr != nil {
		return written, writeErr
	}
	return written, ctx.Err()
}
