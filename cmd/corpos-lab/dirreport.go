package main

import (
	"flag"
	"fmt"
	"io"
	"os"
)

// runDirReport is the shared body of the read-only report commands (cells,
// provenance): parse a directory and -json, load it, and print a table or JSON.
// The directory may come before or after the flags. It returns 2 on a usage or
// flag error and 1 when the directory cannot be loaded or printed.
func runDirReport[T any](name, usage string, args []string, load func(string) (T, error), table, asJSON func(io.Writer, T) error) int {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	jsonOut := fs.Bool("json", false, "print JSON instead of a table")
	// Accept the dir before or after the flags.
	var dir string
	if len(args) > 0 && args[0] != "" && args[0][0] != '-' {
		dir, args = args[0], args[1:]
	}
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if dir == "" && fs.NArg() == 1 {
		dir = fs.Arg(0)
	}
	if dir == "" {
		fmt.Fprintln(os.Stderr, usage)
		return 2
	}
	data, err := load(dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	write := table
	if *jsonOut {
		write = asJSON
	}
	if err := write(os.Stdout, data); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}
