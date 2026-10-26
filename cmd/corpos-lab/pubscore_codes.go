package main

import (
	"os"
	"path/filepath"

	"corpos-lab/internal/neutralprefix"
)

// keyRaterFlags takes --key, --rater-a and --rater-b out of args, the flags
// every key-plus-raters aggregate reads, and returns them with the rest.
func keyRaterFlags(args []string) (keyPath, raterA, raterB string, rest []string) {
	keyPath, args = flagValue(args, "--key")
	raterA, args = flagValue(args, "--rater-a")
	raterB, args = flagValue(args, "--rater-b")
	return keyPath, raterA, raterB, args
}

// mergeCodes reads each file as an id -> code map and merges them in order, a
// later file's code winning for an id both carry.
func mergeCodes(files []string) (map[string]string, error) {
	out := map[string]string{}
	for _, f := range files {
		var codes map[string]string
		if err := readJSON(f, &codes); err != nil {
			return nil, err
		}
		for id, code := range codes {
			out[id] = code
		}
	}
	return out, nil
}

// readNPKey reads and parses a neutral-prefix scoring/key.json.
func readNPKey(path string) (*neutralprefix.KeyFile, error) {
	b, err := os.ReadFile(path) //nolint:gosec // study path
	if err != nil {
		return nil, err
	}
	return neutralprefix.ParseKeyFile(b)
}

// writeFileAll writes data to path (mode 0600), first making its parent
// directories (mode 0750).
func writeFileAll(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}
