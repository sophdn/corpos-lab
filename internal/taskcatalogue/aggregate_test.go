package taskcatalogue

import (
	"strings"
	"testing"
	"time"
)

func rec(id int, day string, project string) PromptRecord {
	ts, _ := time.Parse("2006-01-02", day)
	return PromptRecord{Prompt: Prompt{ID: id, Time: ts}, Attribution: Attribution{Project: project}}
}

func TestAggregate(t *testing.T) {
	tax := mustTax(t)
	prompts := []PromptRecord{
		rec(1, "2026-01-02", "alpha"), rec(2, "2026-01-01", "beta"), rec(3, "2026-01-05", "alpha"),
		rec(4, "2026-01-03", ""), rec(5, "2026-01-04", "alpha"), rec(6, "2026-01-04", "alpha"),
		rec(7, "2026-01-04", "alpha"), rec(8, "2026-01-04", "alpha"), rec(9, "2026-01-04", "alpha"),
	}
	labels := map[int]PromptLabel{
		1: {Status: StatusOK, Role: "opener", Kind: "build-thing"},
		2: {Status: StatusOK, Role: "opener", Kind: "build-thing"},
		3: {Status: StatusOK, Role: "opener", Kind: "build-thing"},
		4: {Status: StatusOK, Role: "opener", Kind: "other"},
		5: {Status: StatusOK, Role: "approve", Kind: "build-thing"},
		6: {Status: StatusOK, Role: "control", Kind: KindNone},
		7: {Status: StatusInvalid},
		8: {Status: StatusError},
		// 9 unlabelled
	}
	s, err := Aggregate(prompts, labels, tax, 0.8)
	if err != nil {
		t.Fatal(err)
	}
	if s.Total != 9 || s.Openers != 4 || s.Invalid != 1 || s.Errored != 1 || s.Unlabelled != 1 {
		t.Errorf("totals %+v", s)
	}
	if s.ByRole["opener"] != 4 || s.ByRole["approve"] != 1 || s.ByRole["control"] != 1 {
		t.Errorf("roles %v", s.ByRole)
	}
	if len(s.Kinds) != 2 || s.Kinds[0].Kind != "build-thing" || s.Kinds[0].Count != 3 {
		t.Fatalf("kinds %+v", s.Kinds)
	}
	top := s.Kinds[0]
	if !top.Catalogue || top.Share != 0.75 || top.First != "2026-01-01" || top.Last != "2026-01-05" ||
		top.Description != "build a thing" || len(top.Members) != 3 || top.Members[0] != 1 {
		t.Errorf("top row %+v", top)
	}
	if top.Projects[0] != (ProjectCount{Project: "alpha", Count: 2}) {
		t.Errorf("projects %+v", top.Projects)
	}
	other := s.Kinds[1]
	if other.Catalogue || other.Projects[0].Project != "(unattributed)" || other.Cumulative != 1 {
		t.Errorf("other must never be catalogued: %+v", other)
	}
	// "other" is skipped, so coverage stops at 75% even though the threshold is 80%.
	if s.Covered != 0.75 {
		t.Errorf("covered %v", s.Covered)
	}
	if len(s.FollowupKinds) != 2 {
		t.Errorf("followups %+v", s.FollowupKinds)
	}
}

func TestAggregateTiesAndThreshold(t *testing.T) {
	tax, err := ParseTaxonomy(taxFixture + "\n[[kind]]\nname = \"a-kind\"\ndescription = \"a\"\n")
	if err != nil {
		t.Fatal(err)
	}
	prompts := []PromptRecord{rec(1, "2026-01-01", "p"), rec(2, "2026-01-01", "q"), rec(3, "2026-01-01", "p")}
	labels := map[int]PromptLabel{
		1: {Status: StatusOK, Role: "opener", Kind: "build-thing"},
		2: {Status: StatusOK, Role: "opener", Kind: "a-kind"},
		3: {Status: StatusOK, Role: "opener", Kind: "a-kind"},
	}
	s, err := Aggregate(prompts, labels, tax, 0.5)
	if err != nil {
		t.Fatal(err)
	}
	if s.Kinds[0].Kind != "a-kind" || !s.Kinds[0].Catalogue || s.Kinds[1].Catalogue {
		t.Errorf("threshold should stop after the first kind: %+v", s.Kinds)
	}
	if s.Kinds[0].Projects[0].Project != "p" || s.Kinds[0].Projects[1].Project != "q" {
		t.Errorf("project tie order: %+v", s.Kinds[0].Projects)
	}
	// Equal counts sort by name.
	labels[1] = PromptLabel{Status: StatusOK, Role: "opener", Kind: "build-thing"}
	labels[3] = PromptLabel{Status: StatusOK, Role: "approve", Kind: "build-thing"}
	s, _ = Aggregate(prompts, labels, tax, 1)
	if s.Kinds[0].Kind != "a-kind" || s.Kinds[1].Kind != "build-thing" {
		t.Errorf("tie order %+v", s.Kinds)
	}
}

func TestAggregateEmpty(t *testing.T) {
	s, err := Aggregate(nil, nil, mustTax(t), 0.8)
	if err != nil || s.Total != 0 || len(s.Kinds) != 0 {
		t.Errorf("%+v %v", s, err)
	}
}

func TestWriteTable(t *testing.T) {
	s := Summary{Total: 3, Openers: 2, Threshold: 0.8, Covered: 1, ByRole: map[string]int{"opener": 2, "approve": 1},
		Kinds: []KindRow{{Kind: "build-thing", Count: 2, Share: 1, Cumulative: 1, Catalogue: true,
			Projects: []ProjectCount{{"a", 1}, {"b", 1}, {"c", 1}, {"d", 1}}}}}
	var b strings.Builder
	if err := WriteTable(&b, s); err != nil {
		t.Fatal(err)
	}
	out := b.String()
	for _, want := range []string{"*   build-thing", "a 1, b 1, c 1", "roles: approve=1 opener=2", "covers 100.0%"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
	if strings.Contains(out, "d 1") {
		t.Error("only the top three projects are listed")
	}
	if err := WriteTable(failWriter{}, s); err == nil {
		t.Error("write failure: want error")
	}
}
