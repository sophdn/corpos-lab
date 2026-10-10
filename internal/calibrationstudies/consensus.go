package calibrationstudies

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"corpos-lab/internal/lengthdistraction"
)

// KeyEntry is one row of a study's key.json: the cell coordinates the consensus
// table groups on. Other key fields (scenario, run) are ignored.
type KeyEntry struct {
	Cls       string `json:"cls"`
	Model     string `json:"model"`
	Condition string `json:"condition"`
}

// consRow accumulates one (class, condition, model) cell.
type consRow struct {
	cls, cond, model string
	n                int // ids in the cell
	cMaj             int // ids whose majority-of-three consensus is C
	cCd              int // ids where claude and deepseek both scored C
}

// ConsensusTable renders the per-cell correct-action table for a calibration
// study, reproducing consensus_table.sh byte-for-byte. It joins the key with the
// three merged rater maps (claude, deepseek, devstral) on id, and reports two
// reads per (class, condition, model) cell:
//
//   - C_maj — the count where the majority-of-three consensus is C. A code wins a
//     cell id when at least two of the three raters agree, else the id is a split
//     and is not counted. This reuses lengthdistraction.Majority, the one
//     measure-of-record consensus rule.
//   - C_cd — the count where the validated pair (claude and deepseek) both scored
//     C. Devstral is advisory on these codes, so this is the robustness check.
//
// The output is tab-separated: a header row then one row per cell, sorted by
// class, then condition, then model. Rates are floored to two decimals.
func ConsensusTable(key map[string]KeyEntry, claude, deepseek, devstral map[string]string) string {
	cells := map[string]*consRow{}
	order := []string{}
	for id, k := range key {
		cellKey := k.Cls + "\x00" + k.Condition + "\x00" + k.Model
		row, ok := cells[cellKey]
		if !ok {
			row = &consRow{cls: k.Cls, cond: k.Condition, model: k.Model}
			cells[cellKey] = row
			order = append(order, cellKey)
		}
		row.n++
		if lengthdistraction.Majority(id, claude, deepseek, devstral) == lengthdistraction.CodeC {
			row.cMaj++
		}
		if claude[id] == lengthdistraction.CodeC && deepseek[id] == lengthdistraction.CodeC {
			row.cCd++
		}
	}

	rows := make([]*consRow, 0, len(order))
	for _, k := range order {
		rows = append(rows, cells[k])
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].cls != rows[j].cls {
			return rows[i].cls < rows[j].cls
		}
		if rows[i].cond != rows[j].cond {
			return rows[i].cond < rows[j].cond
		}
		return rows[i].model < rows[j].model
	})

	var sb strings.Builder
	sb.WriteString("cls\tcond\tmodel\tn\tC_maj\trate_maj\tC_cd\trate_cd\n")
	for _, r := range rows {
		fmt.Fprintf(&sb, "%s\t%s\t%s\t%d\t%d\t%s\t%d\t%s\n",
			r.cls, r.cond, r.model, r.n, r.cMaj, floorRate(r.cMaj, r.n), r.cCd, floorRate(r.cCd, r.n))
	}
	return sb.String()
}

// floorRate formats num/denom floored to two decimals, matching the jq
// expression (num/denom*100|floor)/100 and jq's trailing-zero-trimmed number
// output: 0 for zero, 1 for a full cell, 0.5 / 0.87 for the rest.
func floorRate(num, denom int) string {
	if denom == 0 {
		return "0"
	}
	v := math.Floor(float64(num)/float64(denom)*100) / 100
	return strconv.FormatFloat(v, 'f', -1, 64)
}
