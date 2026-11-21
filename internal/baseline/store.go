package baseline

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// LoadCaptures reads every *.json capture record under dir, sorted by path, into
// a slice. Each file is one model's capture written by the capture half. A file
// that does not parse fails the load rather than being skipped silently — a
// half-written capture is a gap the operator should see, not swallow.
func LoadCaptures(dir string) ([]Capture, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return nil, fmt.Errorf("baseline: list captures in %s: %w", dir, err)
	}
	sort.Strings(files)
	captures := make([]Capture, 0, len(files))
	for _, f := range files {
		data, err := os.ReadFile(f) //nolint:gosec // a capture path under the run's own dir
		if err != nil {
			return nil, fmt.Errorf("baseline: read capture %s: %w", f, err)
		}
		var c Capture
		if err := json.Unmarshal(data, &c); err != nil {
			return nil, fmt.Errorf("baseline: parse capture %s: %w", f, err)
		}
		captures = append(captures, c)
	}
	return captures, nil
}

// LoadRaterScores reads the rater results under dir into one {id: code} map per
// rater family, plus the family names in the same order. Each immediate
// subdirectory of dir is one rater family (the layout `rate` writes:
// <out>/<rater-id>/<stem>.json); its *.json result files are merged into a
// single map. A .prov.json provenance sidecar is not a result and is skipped.
func LoadRaterScores(dir string) ([]map[string]string, []string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, nil, fmt.Errorf("baseline: list rater dirs in %s: %w", dir, err)
	}
	var (
		maps     []map[string]string
		raterIDs []string
	)
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		merged, err := mergeRaterDir(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, nil, err
		}
		maps = append(maps, merged)
		raterIDs = append(raterIDs, e.Name())
	}
	return maps, raterIDs, nil
}

// mergeRaterDir merges every result file in one rater family's directory into a
// single {id: code} map. It skips .prov.json sidecars, which match *.json but
// hold provenance, not codes.
func mergeRaterDir(dir string) (map[string]string, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return nil, fmt.Errorf("baseline: list rater results in %s: %w", dir, err)
	}
	out := map[string]string{}
	for _, f := range files {
		if strings.HasSuffix(f, ".prov.json") {
			continue
		}
		data, err := os.ReadFile(f) //nolint:gosec // a rater result under the run's own dir
		if err != nil {
			return nil, fmt.Errorf("baseline: read rater result %s: %w", f, err)
		}
		var m map[string]string
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, fmt.Errorf("baseline: parse rater result %s: %w", f, err)
		}
		for id, code := range m {
			out[id] = code
		}
	}
	return out, nil
}
