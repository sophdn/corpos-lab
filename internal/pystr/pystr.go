// Package pystr holds the python string idioms the report ports share. The
// report strings are ASCII, so byte lengths and byte slices match python's
// character lengths and slices.
package pystr

import "strings"

// Ljust left-justifies s to width n with spaces, like python's str.ljust.
func Ljust(s string, n int) string {
	if len(s) >= n {
		return s
	}
	return s + strings.Repeat(" ", n-len(s))
}

// Rjust right-justifies s to width n with spaces, like python's str.rjust.
func Rjust(s string, n int) string {
	if len(s) >= n {
		return s
	}
	return strings.Repeat(" ", n-len(s)) + s
}

// Trunc returns s[:n] like a python slice, safe when s is shorter than n.
func Trunc(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s
}
