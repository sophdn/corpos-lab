package taskcatalogue

import (
	"fmt"
	"sort"
)

// PromptRecord is a prompt with its project attribution: one line of
// prompts.jsonl.
type PromptRecord struct {
	Prompt
	Attribution
}

// ProjectCount is one project's share of a kind.
type ProjectCount struct {
	Project string `json:"project"`
	Count   int    `json:"count"`
}

// KindRow is one task kind in the catalogue.
type KindRow struct {
	Kind        string `json:"kind"`
	Description string `json:"description"`
	Count       int    `json:"count"`
	// Share is Count over all openers; Cumulative is the running share down
	// the table, which is sorted by Count.
	Share      float64        `json:"share"`
	Cumulative float64        `json:"cumulative"`
	Projects   []ProjectCount `json:"projects"`
	First      string         `json:"first"`
	Last       string         `json:"last"`
	Members    []int          `json:"members"`
	// Catalogue marks the head of the table that together covers the
	// threshold share of openers: the kinds the replay sweep builds cases for.
	Catalogue bool `json:"catalogue"`
}

// Summary is the kind table and the role totals it was built from.
type Summary struct {
	Total int `json:"total"`
	// ByRole counts every validly labelled prompt by role. Invalid and Errored
	// count the prompts whose last label is not ok, and Unlabelled those with
	// no label at all. The four always sum to Total.
	ByRole     map[string]int `json:"by_role"`
	Invalid    int            `json:"invalid"`
	Errored    int            `json:"errored"`
	Unlabelled int            `json:"unlabelled"`
	Openers    int            `json:"openers"`
	Threshold  float64        `json:"threshold"`
	// Covered is the opener share the catalogue set actually reaches. It falls
	// short of Threshold only when the remaining openers are all in "other"
	// or "none", which are never catalogued.
	Covered float64 `json:"covered"`
	// Kinds is the opener table; FollowupKinds tabulates what non-opener
	// prompts ask for, for the multi-turn cases.
	Kinds         []KindRow `json:"kinds"`
	FollowupKinds []KindRow `json:"followup_kinds"`
}

// KindOther is never catalogued: it is the labeller's "fits nothing" bucket.
const KindOther = "other"

// Aggregate builds the catalogue. Every prompt lands in exactly one place (a
// role bucket, invalid, errored or unlabelled), and every opener in exactly
// one kind; the function checks both sums and returns an error if they do not
// hold, because a count that does not add up cannot weight a sweep.
func Aggregate(prompts []PromptRecord, labels map[int]PromptLabel, tax Taxonomy, threshold float64) (Summary, error) {
	s := Summary{Total: len(prompts), ByRole: map[string]int{}, Threshold: threshold}
	openers := map[string]*KindRow{}
	follow := map[string]*KindRow{}
	projects := map[*KindRow]map[string]int{}
	for _, p := range prompts {
		l, ok := labels[p.ID]
		switch {
		case !ok:
			s.Unlabelled++
			continue
		case l.Status == StatusInvalid:
			s.Invalid++
			continue
		case l.Status != StatusOK:
			s.Errored++
			continue
		}
		s.ByRole[l.Role]++
		table := follow
		if tax.IsOpener(l.Role) {
			table = openers
			s.Openers++
		}
		row := table[l.Kind]
		if row == nil {
			row = &KindRow{Kind: l.Kind, Description: tax.KindDescription(l.Kind)}
			table[l.Kind] = row
			projects[row] = map[string]int{}
		}
		row.Count++
		row.Members = append(row.Members, p.ID)
		proj := p.Project
		if proj == "" {
			proj = "(unattributed)"
		}
		projects[row][proj]++
		day := p.Time.Format("2006-01-02")
		if row.First == "" || day < row.First {
			row.First = day
		}
		if day > row.Last {
			row.Last = day
		}
	}
	roleSum := 0
	for _, n := range s.ByRole {
		roleSum += n
	}
	if got := roleSum + s.Invalid + s.Errored + s.Unlabelled; got != s.Total {
		return Summary{}, fmt.Errorf("aggregate: buckets sum to %d, want %d", got, s.Total)
	}
	s.Kinds = finish(openers, projects, s.Openers)
	s.FollowupKinds = finish(follow, projects, s.Total-s.Openers-s.Invalid-s.Errored-s.Unlabelled)
	kindSum := 0
	for _, r := range s.Kinds {
		kindSum += r.Count
	}
	if kindSum != s.Openers {
		return Summary{}, fmt.Errorf("aggregate: opener kinds sum to %d, want %d", kindSum, s.Openers)
	}
	for i := range s.Kinds {
		r := &s.Kinds[i]
		if s.Covered >= threshold {
			break
		}
		if r.Kind == KindOther || r.Kind == KindNone {
			continue
		}
		r.Catalogue = true
		s.Covered += r.Share
	}
	return s, nil
}

// finish sorts a kind table by count (ties by name) and fills the shares and
// the per-project breakdown.
func finish(table map[string]*KindRow, projects map[*KindRow]map[string]int, denom int) []KindRow {
	out := make([]KindRow, 0, len(table))
	for _, r := range table {
		for proj, n := range projects[r] {
			r.Projects = append(r.Projects, ProjectCount{Project: proj, Count: n})
		}
		sort.Slice(r.Projects, func(i, j int) bool {
			if r.Projects[i].Count != r.Projects[j].Count {
				return r.Projects[i].Count > r.Projects[j].Count
			}
			return r.Projects[i].Project < r.Projects[j].Project
		})
		sort.Ints(r.Members)
		out = append(out, *r)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count != out[j].Count {
			return out[i].Count > out[j].Count
		}
		return out[i].Kind < out[j].Kind
	})
	cum := 0.0
	for i := range out {
		if denom > 0 {
			out[i].Share = float64(out[i].Count) / float64(denom)
		}
		cum += out[i].Share
		out[i].Cumulative = cum
	}
	return out
}
