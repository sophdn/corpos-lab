package main

import (
	"fmt"
	"os"

	"corpos-lab/internal/restaxis"
)

// raDropAxis drops exactly one axis block (default "Rest") from a glyph markdown
// file, structure-preserving, and writes the result to <out.md>. It errors if the
// axis section is not found, so the caller learns the glyph delimits its axes in a
// shape the pattern misses. Ported from drop_axis.py main.
//
// Usage: <glyph.md> <out.md> [--axis Rest]
func raDropAxis(args []string) error {
	axis, rest := flagValue(args, "--axis")
	if axis == "" {
		axis = "Rest"
	}
	if len(rest) != 2 {
		return fmt.Errorf("usage: corpos-lab rest-axis drop-axis <glyph.md> <out.md> [--axis Rest]")
	}
	in, out := rest[0], rest[1]

	b, err := os.ReadFile(in) //nolint:gosec // a glyph markdown path
	if err != nil {
		return err
	}
	text, found := restaxis.DropAxis(string(b), axis)
	if !found {
		return fmt.Errorf("axis %q section not found in %s", axis, in)
	}
	if err := os.WriteFile(out, []byte(text), 0o600); err != nil {
		return err
	}
	return nil
}

func init() {
	registerRestAxisMode("drop-axis", raDropAxis)
}
