package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

// parseDirFlags parses fs over args and returns the one directory argument,
// which may come before or after the flags. It returns false after a flag
// error (fs has already reported it) or, having printed usage, when no
// directory was given.
func parseDirFlags(fs *flag.FlagSet, args []string, usage string) (string, bool) {
	var dir string
	if len(args) > 0 && args[0] != "" && args[0][0] != '-' {
		dir, args = args[0], args[1:]
	}
	if err := fs.Parse(args); err != nil {
		return "", false
	}
	rest := fs.Args()
	if dir == "" && len(rest) > 0 {
		dir, rest = rest[0], rest[1:]
	}
	// One dir and nothing else: a stray extra argument used to be dropped
	// silently when the dir came first (bug 1454).
	if dir == "" || len(rest) > 0 {
		fmt.Fprintln(os.Stderr, usage)
		return "", false
	}
	return dir, true
}

// splitValueFlags reads args in any order: a flag in switches sets its bool, a
// flag in targets takes its value (the next arg, or the text after '=' in the
// -flag=value form), and everything else is returned in order. It prints the
// problem and returns false when a value flag is the last arg.
func splitValueFlags(args []string, targets map[string]*string, switches map[string]*bool) ([]string, bool) {
	setters := make(map[string]func(string) error, len(targets))
	for name, target := range targets {
		setters[name] = func(v string) error { *target = v; return nil }
	}
	rest, err := scanFlags(args, setters, switches)
	if err != nil {
		fmt.Fprintf(os.Stderr, "corpos-lab: %v\n", err)
		return nil, false
	}
	return rest, true
}

// scanFlags is the one walk behind splitValueFlags and run-study: a flag in
// switches sets its bool, a flag in setters is handed its value (the next arg,
// or the text after the first '=' in -flag=value), and everything else is
// returned in order. It stops at the first setter error or at a value flag
// with no value.
func scanFlags(args []string, setters map[string]func(string) error, switches map[string]*bool) ([]string, error) {
	var rest []string
	for i := 0; i < len(args); i++ {
		if sw, ok := switches[args[i]]; ok {
			*sw = true
			continue
		}
		name, val, inline := strings.Cut(args[i], "=")
		set, ok := setters[name]
		if !ok {
			rest = append(rest, args[i])
			continue
		}
		if !inline {
			if i+1 >= len(args) {
				return nil, fmt.Errorf("%s needs a value", name)
			}
			i++
			val = args[i]
		}
		if err := set(val); err != nil {
			return nil, err
		}
	}
	return rest, nil
}
