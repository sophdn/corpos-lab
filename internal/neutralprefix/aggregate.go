package neutralprefix

import (
	"fmt"
	"math"
	"strings"
)

// aggregateConds and aggregateClasses are aggregate.py's CONDS and CLASSES, in
// report order.
var (
	aggregateConds   = []string{"baseline", "neutral_prefix", "glyph_only", "imperative_only"}
	aggregateClasses = []string{"casg-direct", "formal-step-context-bypass",
		"parent-state-check-bypass", "conditional-gate-uniform-default"}
	aggregateModels = []string{"mistral", "phi4", "qwen38"}
	aggregateScens  = []string{"1", "2"}
)

// crate returns (C-count, n, rate) for a list of codes; rate is NaN when n==0,
// matching aggregate.py's crate.
func crate(items []string) (int, int, float64) {
	n := len(items)
	c := 0
	for _, x := range items {
		if x == "C" {
			c++
		}
	}
	if n == 0 {
		return c, n, math.NaN()
	}
	return c, n, float64(c) / float64(n)
}

// Aggregate reproduces aggregate.py's stdout. key supplies id metadata; a is
// rater A (rid -> code); b is rater B (rid -> code) when hasB is true, in which
// case the reported code per id is the strict consensus C ("C" only when both
// raters code C, else "x") and A/B agreement lines are printed. With hasB false,
// the code is rater A's own (or "?" when absent). It returns the full report text.
func Aggregate(key *KeyFile, a, b map[string]string, hasB bool) string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "rater-a: %d ids", len(a))
	if hasB {
		fmt.Fprintf(&sb, " | rater-b: %d ids", len(b))
	}
	sb.WriteByte('\n')

	if hasB {
		writeAgreement(&sb, key, a, b)
	}

	codeFor := func(i string) string {
		if hasB {
			if a[i] == "C" && b[i] == "C" {
				return "C"
			}
			return "x"
		}
		if v, ok := a[i]; ok {
			return v
		}
		return "?"
	}

	sb.WriteString("\n== pooled C-rate per class x condition ==")
	if hasB {
		sb.WriteString("  [strict-consensus]")
	} else {
		sb.WriteString("  [rater-a]")
	}
	sb.WriteByte('\n')
	writePooled(&sb, key, codeFor)
	writeCells(&sb, key, codeFor)
	return sb.String()
}

// writeAgreement writes the two A/B agreement lines over the ids both raters
// coded, in key order. With no common id the ratios are NaN, as in the python.
func writeAgreement(sb *strings.Builder, key *KeyFile, a, b map[string]string) {
	agree, cagree, nc := 0, 0, 0
	for _, i := range key.Order {
		_, okA := a[i]
		_, okB := b[i]
		if !okA || !okB {
			continue
		}
		nc++
		if a[i] == b[i] {
			agree++
		}
		if (a[i] == "C") == (b[i] == "C") {
			cagree++
		}
	}
	fmt.Fprintf(sb, "agreement (exact code): %d/%d = %.3f\n", agree, nc, float64(agree)/float64(nc))
	fmt.Fprintf(sb, "agreement (C vs not-C): %d/%d = %.3f\n", cagree, nc, float64(cagree)/float64(nc))
}

// writePooled writes the pooled C-rate header and one row per class.
func writePooled(sb *strings.Builder, key *KeyFile, codeFor func(string) string) {
	hdr := ljust("class", 34)
	for _, c := range aggregateConds {
		hdr += ljust(truncate(c, 9), 11)
	}
	sb.WriteString(hdr + "\n")

	pooled := map[string][]string{}
	for _, i := range key.Order {
		m := key.Meta[i]
		pooled[m.Cls+"\x00"+m.Condition] = append(pooled[m.Cls+"\x00"+m.Condition], codeFor(i))
	}
	for _, cls := range aggregateClasses {
		row := ljust(cls, 34)
		for _, cond := range aggregateConds {
			c, n, r := crate(pooled[cls+"\x00"+cond])
			cell := fmt.Sprintf("%2d/%s=%s", c, ljust(fmt.Sprintf("%d", n), 3), f42(r))
			row += ljust(cell, 11)
		}
		sb.WriteString(row + "\n")
	}
}

// writeCells writes the C-count per class x model x scenario x condition lines.
func writeCells(sb *strings.Builder, key *KeyFile, codeFor func(string) string) {
	sb.WriteString("\n== C-count/16 per class x model x scenario x condition ==\n")
	cells := map[string][]string{}
	for _, i := range key.Order {
		m := key.Meta[i]
		k := m.Cls + "\x00" + m.Model + "\x00" + m.Scenario + "\x00" + m.Condition
		cells[k] = append(cells[k], codeFor(i))
	}
	for _, cls := range aggregateClasses {
		for _, model := range aggregateModels {
			for _, sc := range aggregateScens {
				var parts []string
				for _, cond := range aggregateConds {
					c, n, _ := crate(cells[cls+"\x00"+model+"\x00"+sc+"\x00"+cond])
					parts = append(parts, fmt.Sprintf("%s=%2d/%d", truncate(cond, 4), c, n))
				}
				fmt.Fprintf(sb, "  %s %s s%s  %s\n",
					ljust(truncate(cls, 22), 22), ljust(model, 8), sc, strings.Join(parts, "  "))
			}
		}
	}
}

// f42 formats a float like Python's "{:4.2f}": width 4, two decimals, and " nan"
// (lowercase, width 4) for NaN.
func f42(x float64) string {
	if math.IsNaN(x) {
		return " nan"
	}
	return fmt.Sprintf("%4.2f", x)
}

// ljust left-justifies s to width w with spaces, byte-for-byte like Python's
// str.ljust for the ASCII strings used here.
func ljust(s string, w int) string {
	if len(s) >= w {
		return s
	}
	return s + strings.Repeat(" ", w-len(s))
}

// truncate returns s[:n] like a Python slice, safe when s is shorter than n. The
// aggregate strings are ASCII, so a byte slice matches Python's character slice.
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
