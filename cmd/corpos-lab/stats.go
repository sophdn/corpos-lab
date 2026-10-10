package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"strconv"

	"corpos-lab/internal/behaveq"
	"corpos-lab/internal/cartographer"
	"corpos-lab/internal/setupcompletion"
)

const statsUsage = "usage: corpos-lab stats fisher A B C D | wilson K N | newcombe [-conf 95|90] K1 N1 K2 N2 | kappa RATER_A.json RATER_B.json"

// errStatsUsage marks a missing or unknown stats mode.
var errStatsUsage = errors.New(statsUsage)

// runStats exposes the lab's Go statistics so paper verification needs no
// hand-written Python (there is no scipy here, and a hand-rolled Newcombe once
// had its pairing reversed). Every mode calls the one existing implementation:
// cartographer.FisherTwoSided, cartographer.Wilson, behaveq.NewcombeDiffCI and
// setupcompletion.Kappa.
func runStats(args []string) int {
	err := statsCmd(os.Stdout, args)
	switch {
	case errors.Is(err, errStatsUsage):
		fmt.Fprintln(os.Stderr, statsUsage)
		return 2
	case err != nil:
		fmt.Fprintf(os.Stderr, "corpos-lab stats: %v\n", err)
		return 1
	}
	return 0
}

func statsCmd(w io.Writer, args []string) error {
	if len(args) == 0 {
		return errStatsUsage
	}
	mode, rest := args[0], args[1:]
	switch mode {
	case "fisher":
		return statsFisher(w, rest)
	case "wilson":
		return statsWilson(w, rest)
	case "newcombe":
		return statsNewcombe(w, rest)
	case "kappa":
		return statsKappa(w, rest)
	default:
		return errStatsUsage
	}
}

// statsFisher prints the two-sided Fisher exact p for a 2x2 table A B C D.
func statsFisher(w io.Writer, rest []string) error {
	n, err := counts(rest, 4)
	if err != nil {
		return err
	}
	p := cartographer.FisherTwoSided(n[0], n[1], n[2], n[3])
	fmt.Fprintf(w, "fisher two-sided [[%d,%d],[%d,%d]] p=%.1e (%.6g)\n", n[0], n[1], n[2], n[3], p, p)
	return nil
}

// statsWilson prints the 95% Wilson interval for K of N.
func statsWilson(w io.Writer, rest []string) error {
	n, err := counts(rest, 2)
	if err != nil {
		return err
	}
	if err := proportion(n[0], n[1]); err != nil {
		return err
	}
	lo, hi := cartographer.Wilson(n[0], n[1])
	fmt.Fprintf(w, "wilson 95%% %d/%d p=%.2f [%.2f, %.2f]\n", n[0], n[1], float64(n[0])/float64(n[1]), lo, hi)
	return nil
}

// statsNewcombe prints the Newcombe interval for the difference K1/N1 - K2/N2 at
// 95% (default) or 90% confidence.
func statsNewcombe(w io.Writer, rest []string) error {
	fs := flag.NewFlagSet("newcombe", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	conf := fs.Int("conf", 95, "confidence level: 95 or 90")
	if err := fs.Parse(rest); err != nil {
		return err
	}
	z, ok := map[int]float64{95: cartographer.Z95, 90: 1.6448536269514722}[*conf]
	if !ok {
		return fmt.Errorf("-conf must be 95 or 90, got %d", *conf)
	}
	n, err := counts(fs.Args(), 4)
	if err != nil {
		return err
	}
	if err := proportion(n[0], n[1]); err != nil {
		return err
	}
	if err := proportion(n[2], n[3]); err != nil {
		return err
	}
	lo, hi := behaveq.NewcombeDiffCI(n[0], n[1], n[2], n[3], z)
	diff := float64(n[0])/float64(n[1]) - float64(n[2])/float64(n[3])
	fmt.Fprintf(w, "newcombe %d%% (%d/%d) - (%d/%d) diff=%+.2f [%+.2f, %+.2f]\n", *conf, n[0], n[1], n[2], n[3], diff, lo, hi)
	return nil
}

// statsKappa prints Cohen's kappa over the ids two rater files share. An
// undefined kappa is printed and then reported as an error.
func statsKappa(w io.Writer, rest []string) error {
	if len(rest) != 2 {
		return fmt.Errorf("kappa needs two rater output files (JSON id -> code maps)")
	}
	a, err := readCodes(rest[0])
	if err != nil {
		return err
	}
	b, err := readCodes(rest[1])
	if err != nil {
		return err
	}
	var pairs [][2]string
	ids := make([]string, 0, len(a))
	for id := range a {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	agree := 0
	for _, id := range ids {
		if cb, ok := b[id]; ok {
			pairs = append(pairs, [2]string{a[id], cb})
			if a[id] == cb {
				agree++
			}
		}
	}
	if len(pairs) == 0 {
		return fmt.Errorf("the two files share no ids")
	}
	unpaired := len(a) + len(b) - 2*len(pairs)
	k := setupcompletion.Kappa(pairs)
	fmt.Fprintf(w, "cohen's kappa n=%d agree=%.2f kappa=%.2f unpaired=%d\n", len(pairs), float64(agree)/float64(len(pairs)), k, unpaired)
	if math.IsNaN(k) {
		return fmt.Errorf("kappa undefined")
	}
	return nil
}

func counts(args []string, n int) ([]int, error) {
	if len(args) != n {
		return nil, fmt.Errorf("want %d counts, got %d (%s)", n, len(args), statsUsage)
	}
	out := make([]int, n)
	for i, a := range args {
		v, err := strconv.Atoi(a)
		if err != nil || v < 0 {
			return nil, fmt.Errorf("count %q is not a non-negative integer", a)
		}
		out[i] = v
	}
	return out, nil
}

func proportion(k, n int) error {
	if n == 0 || k > n {
		return fmt.Errorf("%d/%d is not a proportion (want 0 <= k <= n, n > 0)", k, n)
	}
	return nil
}

func readCodes(path string) (map[string]string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m map[string]string
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, fmt.Errorf("%s: want a JSON object of id -> code: %w", path, err)
	}
	return m, nil
}
