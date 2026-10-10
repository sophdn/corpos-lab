package matchedcontent

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// labelsRe matches the `const LABELS = [...]` array in the template, mirroring
// python re.sub(r'const LABELS = \[.*?\];', ..., flags=re.S).
var labelsRe = regexp.MustCompile(`(?s)const LABELS = \[.*?\];`)

// RenderItemsJSON serializes the items exactly as python's
// json.dumps(items, ensure_ascii=False).replace("</", "<\\/") does: keys in
// insertion order (id, task, gt, resp), ", " / ": " separators, then the
// <script>-safe </ rewrite.
func RenderItemsJSON(items []Item) string {
	var b strings.Builder
	b.WriteByte('[')
	for i, it := range items {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString("{")
		b.WriteString(`"id": ` + pyJSONString(it.ID, false))
		b.WriteString(`, "task": ` + pyJSONString(it.Task, false))
		b.WriteString(`, "gt": ` + pyJSONString(it.GT, false))
		b.WriteString(`, "resp": ` + pyJSONString(it.Resp, false))
		b.WriteString("}")
	}
	b.WriteByte(']')
	return strings.ReplaceAll(b.String(), "</", `<\/`)
}

// RenderCmapJSON serializes the operator-held map exactly as python's
// json.dumps(cmap, indent=1) does: one-space indent per level, "," item
// separator, ": " key separator, keys in insertion order (order), nested values
// in field order (model, condition, run, claude).
func RenderCmapJSON(order []string, cmap map[string]CmapEntry) string {
	if len(order) == 0 {
		return "{}"
	}
	var b strings.Builder
	b.WriteString("{\n")
	for i, hid := range order {
		e := cmap[hid]
		b.WriteString(" " + pyJSONString(hid, true) + ": {\n")
		b.WriteString("  " + pyJSONString("model", true) + ": " + pyJSONString(e.Model, true) + ",\n")
		b.WriteString("  " + pyJSONString("condition", true) + ": " + pyJSONString(e.Condition, true) + ",\n")
		b.WriteString(fmt.Sprintf("  %s: %d,\n", pyJSONString("run", true), e.Run))
		b.WriteString("  " + pyJSONString("claude", true) + ": " + pyJSONString(e.Claude, true) + "\n")
		b.WriteString(" }")
		if i < len(order)-1 {
			b.WriteString(",")
		}
		b.WriteString("\n")
	}
	b.WriteString("}")
	return b.String()
}

// FillTemplate ports the template-substitution chain: replace the LABELS array,
// the four title/label string swaps, then the items JSON and the answers key.
func FillTemplate(tpl, itemsJSON string) string {
	out := labelsRe.ReplaceAllLiteralString(tpl, LabelsBlock)
	out = strings.ReplaceAll(out, "Rest-axis calibration", "casg-direct scoring calibration")
	out = strings.ReplaceAll(out, "restaxis_calib_", "casg_human_anchor_")
	out = strings.ReplaceAll(out,
		"Ground truth &mdash; correct completion &amp; the over-fire to watch for",
		"Scoring rubric &mdash; what each code means")
	out = strings.ReplaceAll(out,
		"Ground truth — correct completion &amp; the over-fire to watch for",
		"Scoring rubric — what each code means")
	out = strings.ReplaceAll(out, "__ITEMS_JSON__", itemsJSON)
	out = strings.ReplaceAll(out, "__ANSWERS_KEY__", `"v1"`)
	return out
}

// StrataCounterRepr reproduces repr(Counter((model, condition) for cmap.values()))
// over the ids in build order: Counter({('m', 'c'): n, ...}) in most-common
// order (count descending, ties in first-seen order).
func StrataCounterRepr(order []string, cmap map[string]CmapEntry) string {
	keys := make([]string, 0, len(order))
	for _, hid := range order {
		e := cmap[hid]
		keys = append(keys, fmt.Sprintf("('%s', '%s')", e.Model, e.Condition))
	}
	return counterRepr(keys)
}

// ClaudeCounterRepr reproduces repr(Counter(m['claude'] for cmap.values())).
func ClaudeCounterRepr(order []string, cmap map[string]CmapEntry) string {
	keys := make([]string, 0, len(order))
	for _, hid := range order {
		keys = append(keys, fmt.Sprintf("'%s'", cmap[hid].Claude))
	}
	return counterRepr(keys)
}

// counterRepr renders repr(Counter(seq)): Counter({k: v, ...}) with keys in
// most_common order — count descending, ties broken by first-seen order.
func counterRepr(seq []string) string {
	var order []string
	count := map[string]int{}
	firstSeen := map[string]int{}
	for i, k := range seq {
		if _, ok := count[k]; !ok {
			order = append(order, k)
			firstSeen[k] = i
		}
		count[k]++
	}
	sort.SliceStable(order, func(i, j int) bool {
		if count[order[i]] != count[order[j]] {
			return count[order[i]] > count[order[j]]
		}
		return firstSeen[order[i]] < firstSeen[order[j]]
	})
	parts := make([]string, len(order))
	for i, k := range order {
		parts[i] = fmt.Sprintf("%s: %d", k, count[k])
	}
	return "Counter({" + strings.Join(parts, ", ") + "})"
}

// pyJSONString escapes a string as CPython's json encoder does. ensureAscii=false
// matches ensure_ascii=False (non-ASCII passed through); ensureAscii=true emits
// \uXXXX for non-ASCII (surrogate pair above U+FFFF), matching json.dumps' default.
func pyJSONString(s string, ensureAscii bool) string {
	var b strings.Builder
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '\\':
			b.WriteString(`\\`)
		case '"':
			b.WriteString(`\"`)
		case '\b':
			b.WriteString(`\b`)
		case '\f':
			b.WriteString(`\f`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		default:
			switch {
			case r < 0x20:
				fmt.Fprintf(&b, `\u%04x`, r)
			case ensureAscii && r > 0x7f:
				if r > 0xffff {
					rr := r - 0x10000
					fmt.Fprintf(&b, `\u%04x\u%04x`, 0xd800+(rr>>10), 0xdc00+(rr&0x3ff))
				} else {
					fmt.Fprintf(&b, `\u%04x`, r)
				}
			default:
				b.WriteRune(r)
			}
		}
	}
	b.WriteByte('"')
	return b.String()
}
