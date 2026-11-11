package pubshared

import (
	"fmt"
	"sort"
	"strings"
)

// Allowed key/kind sets, ported from build_app.py.
var (
	allowedItemKeys  = map[string]bool{"id": true, "blocks": true}
	allowedBlockKeys = map[string]bool{"heading": true, "text": true, "kind": true}
	allowedKinds     = map[string]bool{"context": true, "ground": true, "response": true}
	allowedLabelKeys = map[string]bool{"code": true, "desc": true, "key": true, "criteria": true}
)

// BlindnessError marks a validation failure where an item, block, or label
// carried a field the human must not see — the load-bearing blindness guard.
type BlindnessError struct{ msg string }

func (e *BlindnessError) Error() string { return e.msg }

// forbidden formats the sorted forbidden-field slice like python's sorted(extra).
func forbidden(extra []string) string {
	sort.Strings(extra)
	quoted := make([]string, len(extra))
	for i, k := range extra {
		quoted[i] = "'" + k + "'"
	}
	return "[" + strings.Join(quoted, ", ") + "]"
}

func sortedKeyList(keys map[string]bool) string {
	ks := make([]string, 0, len(keys))
	for k := range keys {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	for i, k := range ks {
		ks[i] = "'" + k + "'"
	}
	return "[" + strings.Join(ks, ", ") + "]"
}

// extraKeys returns the keys of obj not in allowed.
func extraKeys(obj oobject, allowed map[string]bool) []string {
	var extra []string
	for _, k := range obj.keys() {
		if !allowed[k] {
			extra = append(extra, k)
		}
	}
	return extra
}

// ValidateItems ports build_app.validate_items: it enforces blindness structurally
// by rejecting any field beyond {id, blocks} on an item and {heading, text, kind}
// on a block, and validates ids and block kinds.
func ValidateItems(items any) error {
	arr, ok := items.([]any)
	if !ok || len(arr) == 0 {
		return fmt.Errorf("items must be a non-empty JSON array")
	}
	seen := map[string]bool{}
	for _, raw := range arr {
		if err := validateItem(raw, seen); err != nil {
			return err
		}
	}
	return nil
}

// validateItem checks one item and records its id in seen.
func validateItem(raw any, seen map[string]bool) error {
	it, ok := raw.(oobject)
	if !ok {
		return fmt.Errorf("each item must be an object")
	}
	if extra := extraKeys(it, allowedItemKeys); len(extra) > 0 {
		return &BlindnessError{fmt.Sprintf(
			"item '%s' carries forbidden field(s) %s — an item may hold only %s; the arm and any machine "+
				"labels stay in a held key file, never in the app", stringField(it, "id", "?"), forbidden(extra), sortedKeyList(allowedItemKeys))}
	}
	idv, ok := it.get("id")
	ids, isStr := idv.(string)
	if !ok || !isStr || ids == "" {
		return fmt.Errorf("item is missing a non-empty string id")
	}
	if seen[ids] {
		return fmt.Errorf("duplicate item id '%s'", ids)
	}
	seen[ids] = true
	blocksV, _ := it.get("blocks")
	blocks, ok := blocksV.([]any)
	if !ok || len(blocks) == 0 {
		return fmt.Errorf("item '%s' must have a non-empty blocks array", ids)
	}
	for _, braw := range blocks {
		if err := validateBlock(ids, braw); err != nil {
			return err
		}
	}
	return nil
}

// validateBlock checks one block of item ids. A missing or non-string kind is
// read as "context".
func validateBlock(ids string, braw any) error {
	blk, ok := braw.(oobject)
	if !ok {
		return fmt.Errorf("item '%s': each block must be an object", ids)
	}
	if extra := extraKeys(blk, allowedBlockKeys); len(extra) > 0 {
		return &BlindnessError{fmt.Sprintf(
			"item '%s': block carries forbidden field(s) %s — a block may hold only %s",
			ids, forbidden(extra), sortedKeyList(allowedBlockKeys))}
	}
	kind := stringField(blk, "kind", "context")
	if !allowedKinds[kind] {
		return fmt.Errorf("item '%s': block kind '%s' not one of %s", ids, kind, sortedKeyList(allowedKinds))
	}
	return nil
}

// ValidateLabels ports build_app.validate_labels.
func ValidateLabels(labels any) error {
	arr, ok := labels.([]any)
	if !ok || len(arr) == 0 {
		return fmt.Errorf("labels must be a non-empty JSON array")
	}
	for _, raw := range arr {
		if err := validateLabel(raw); err != nil {
			return err
		}
	}
	return nil
}

// validateLabel checks one label: code and desc present, no forbidden field, and
// criteria, when present, a non-empty array of non-blank strings.
func validateLabel(raw any) error {
	lab, ok := raw.(oobject)
	if !ok {
		return fmt.Errorf("each label needs at least code and desc")
	}
	_, hasCode := lab.get("code")
	_, hasDesc := lab.get("desc")
	if !hasCode || !hasDesc {
		return fmt.Errorf("each label needs at least code and desc")
	}
	if extra := extraKeys(lab, allowedLabelKeys); len(extra) > 0 {
		return &BlindnessError{fmt.Sprintf(
			"label '%s' carries forbidden field(s) %s — a label may hold only %s",
			labelCode(lab), forbidden(extra), sortedKeyList(allowedLabelKeys))}
	}
	critV, ok := lab.get("criteria")
	if !ok {
		return nil
	}
	crit, ok := critV.([]any)
	code := labelCode(lab)
	if !ok || len(crit) == 0 {
		return fmt.Errorf("label '%s': criteria must be a non-empty array of strings", code)
	}
	for _, c := range crit {
		cs, ok := c.(string)
		if !ok || pyStrip(cs) == "" {
			return fmt.Errorf("label '%s': every criterion must be a non-empty string", code)
		}
	}
	return nil
}

func labelCode(lab oobject) string {
	return stringField(lab, "code", "?")
}

// stringField returns obj[key] when it is a string, else def.
func stringField(obj oobject, key, def string) string {
	if v, ok := obj.get(key); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return def
}

// inlineJSON reproduces build_app.build_app.inline_json: DumpsPy(ensure_ascii=False)
// then the <script>-safe </ rewrite.
func inlineJSON(v any) string {
	return escapeScriptClose(DumpsPy(v, false))
}

// BuildApp ports build_app.build_app: it validates the items and labels, then
// fills the template placeholders. rubric is nil for no rubric panel. It returns
// the assembled HTML.
func BuildApp(template string, items, labels any, title, storeKey string, rubric *string) (string, error) {
	if err := ValidateItems(items); err != nil {
		return "", err
	}
	if err := ValidateLabels(labels); err != nil {
		return "", err
	}
	var rubricVal any
	if rubric != nil {
		rubricVal = *rubric
	}
	html := template
	html = strings.ReplaceAll(html, "__TITLE__", title)
	html = strings.ReplaceAll(html, "__ITEMS_JSON__", inlineJSON(items))
	html = strings.ReplaceAll(html, "__LABELS_JSON__", inlineJSON(labels))
	html = strings.ReplaceAll(html, "__RUBRIC_JSON__", inlineJSON(rubricVal))
	html = strings.ReplaceAll(html, "__STORE_KEY__", DumpsPy(storeKey, true))
	return html, nil
}

// MissingCriteria ports build_app.missing_criteria: true when a judgment-style
// label set (more than two codes) ships with no per-code criteria and no rubric
// panel. Advisory only.
func MissingCriteria(labels any, rubric *string) bool {
	if rubric != nil && *rubric != "" {
		return false
	}
	arr, ok := labels.([]any)
	if !ok || len(arr) <= 2 {
		return false
	}
	for _, raw := range arr {
		lab, ok := raw.(oobject)
		if !ok {
			continue
		}
		if critV, ok := lab.get("criteria"); ok {
			if crit, ok := critV.([]any); ok && len(crit) > 0 {
				return false
			}
		}
	}
	return true
}
