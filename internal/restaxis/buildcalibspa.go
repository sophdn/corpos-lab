package restaxis

// This file ports build_calib_spa.py: it injects the calibration items into the
// blind-scoring single-page app template. The transform is pure and sans-IO —
// the caller reads calib_map.json, the CORRECT_COMPLETIONS.md files, the
// scenario materials, and the response texts, then hands the assembled items and
// the template string in here.
//
// The item JSON must be byte-identical to python's
// json.dumps(items, ensure_ascii=False).replace("</", "<\\/"), so it is built by
// a hand-rolled serializer that reproduces CPython's separators (", " and ": ")
// and its ensure_ascii=False string escaping, rather than by encoding/json
// (which uses no separators and escapes different characters).

import (
	"regexp"
	"strconv"
	"strings"

	"corpos-lab/internal/pubshared"
)

// CalibItem is one blind scoring item embedded into the app: only the fields the
// operator is allowed to see (id, task, ground truth, response) — never the arm
// or the judges' labels. The JSON key order (id, task, gt, resp) matches the
// python dict insertion order.
type CalibItem = pubshared.AnchorItem

// scenarioSplitRe matches the "## Scenario N — ..." headings in a
// CORRECT_COMPLETIONS.md file, capturing the scenario number. It mirrors the
// python re.split pattern r'(?im)^#{1,3}\s*scenario[ _]?(\d+)\b.*$'.
var scenarioSplitRe = regexp.MustCompile(`(?im)^#{1,3}\s*scenario[ _]?(\d+)\b.*$`)

// ParseCompletions parses a CORRECT_COMPLETIONS.md body into a map from scenario
// number to the completion text that follows its heading, stripped. It
// reproduces python parse_completions: re.split on the scenario heading with a
// captured number, then pairing each number with the text up to the next
// heading. Go's regexp.Split drops capture groups, so the split is done by hand
// over the match offsets.
func ParseCompletions(txt string) map[int]string {
	out := map[int]string{}
	locs := scenarioSplitRe.FindAllStringSubmatchIndex(txt, -1)
	for i, loc := range locs {
		num, err := strconv.Atoi(txt[loc[2]:loc[3]])
		if err != nil {
			continue
		}
		start := loc[1]
		end := len(txt)
		if i+1 < len(locs) {
			end = locs[i+1][0]
		}
		out[num] = pyStrip(txt[start:end])
	}
	return out
}

// BuildCalibSPA fills the template with the serialized items and the answers key,
// mirroring python's
// TPL.read_text().replace("__ITEMS_JSON__", data).replace("__ANSWERS_KEY__", '"v1"').
// The answers key is embedded already quoted, e.g. "v1".
func BuildCalibSPA(template string, items []CalibItem, answersKey string) string {
	data := pubshared.RenderAnchorItemsJSON(items)
	html := strings.ReplaceAll(template, "__ITEMS_JSON__", data)
	html = strings.ReplaceAll(html, "__ANSWERS_KEY__", `"`+answersKey+`"`)
	return html
}
