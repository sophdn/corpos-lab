package cartographer

import (
	"fmt"
	"strings"
)

// condShort renders a condition name the way analyze's per-taboo header does:
// the text before the first underscore, truncated to nine runes. Both
// cartographer_* conditions collapse to "cartograp", exactly as the python
// produces two identical columns.
func condShort(cond string) string {
	head := cond
	if i := strings.IndexByte(cond, '_'); i >= 0 {
		head = cond[:i]
	}
	r := []rune(head)
	if len(r) > 9 {
		r = r[:9]
	}
	return string(r)
}

// Analyze produces the byte-for-byte text of analyze.py: per-condition,
// per-class coverage rates with Wilson intervals; the annotated-minus-cartographer
// interaction gap with a pooled non-first-principles Fisher test; per-non-FP-taboo
// scan recovery; and the full per-taboo coverage table. Ported from analyze.main.
func Analyze(taboos []Taboo, grid []GridEntry) string {
	class := make(map[string]string, len(taboos))
	scanTarget := make(map[string]bool, len(taboos))
	for _, t := range taboos {
		class[t.Slug] = t.Class
		scanTarget[t.Slug] = t.ScanTarget
	}
	fp := fpSlugs(taboos)
	nonfp := nonFPSlugs(taboos)

	var b strings.Builder
	line := func(s string) { b.WriteString(s); b.WriteByte('\n') }

	line("== coverage rate per condition per class (k/n, Wilson 95%) ==")
	line(fmt.Sprintf("%-32s %22s %24s", "condition", "first-principles", "non-first-principles"))
	for _, cond := range Conds {
		fk, fn := rate(grid, cond, fp)
		nk, nn := rate(grid, cond, nonfp)
		fl, fh := Wilson(fk, fn)
		nl, nh := Wilson(nk, nn)
		fpS := fmt.Sprintf("%d/%d [%.2f,%.2f]", fk, fn, fl, fh)
		nfS := fmt.Sprintf("%d/%d [%.2f,%.2f]", nk, nn, nl, nh)
		line(fmt.Sprintf("%-32s %22s %24s", cond, fpS, nfS))
	}

	line("")
	line("== primary interaction: annotated - cartographer coverage gap ==")
	const annot, cart = "annotated_instrument", "cartographer_instrument"
	for _, lab := range []struct {
		name  string
		slugs []string
	}{{"first-principles", fp}, {"non-first-principles", nonfp}} {
		ak, an := rate(grid, annot, lab.slugs)
		ck, cn := rate(grid, cart, lab.slugs)
		ap := ratio(ak, an)
		cp := ratio(ck, cn)
		line(fmt.Sprintf("  %-22s annotated %.2f - cartographer %.2f = gap %+.2f", lab.name, ap, cp, ap-cp))
	}
	ak, an := rate(grid, annot, nonfp)
	ck, cn := rate(grid, cart, nonfp)
	pPooled := FisherTwoSided(ak, an-ak, ck, cn-ck)
	line(fmt.Sprintf("  Fisher (annotated vs cartographer, non-first-principles pooled): p=%s", formatG(pPooled, 3)))

	line("")
	line("== scan recovery: cartographer vs cartographer+scan, per non-FP taboo ==")
	const scan = "cartographer_scan_instrument"
	for _, s := range nonfp {
		ck, cn := rate(grid, cart, []string{s})
		sk, sn := rate(grid, scan, []string{s})
		tgt := "not-targeted"
		if scanTarget[s] {
			tgt = "scan-target"
		}
		var p float64
		if sn != 0 && cn != 0 {
			p = FisherTwoSided(sk, sn-sk, ck, cn-ck)
		} else {
			p = nan()
		}
		line(fmt.Sprintf("  %-42s (%-16s, %s) cart %d/%d -> scan %d/%d  p=%s",
			s, class[s], tgt, ck, cn, sk, sn, formatG(p, 3)))
	}

	line("")
	line("== per-taboo coverage rate (k/n) by condition ==")
	var hdr strings.Builder
	hdr.WriteString(fmt.Sprintf("%-44s%-16s", "taboo", "class"))
	for _, cond := range Conds {
		hdr.WriteString(fmt.Sprintf("%10s", condShort(cond)))
	}
	line(hdr.String())
	for _, t := range taboos {
		var cells strings.Builder
		for _, cond := range Conds {
			k, n := rate(grid, cond, []string{t.Slug})
			cells.WriteString(fmt.Sprintf("%10s", fmt.Sprintf("%d/%d", k, n)))
		}
		line(fmt.Sprintf("%-44s%-16s%s", t.Slug, t.Class, cells.String()))
	}

	return b.String()
}

// ratio returns k/n as a float, or 0 when n==0, matching python's `k/n if n else 0`.
func ratio(k, n int) float64 {
	if n == 0 {
		return 0
	}
	return float64(k) / float64(n)
}
