package behaveq

import (
	"fmt"
	"testing"
)

func TestScoreCleared(t *testing.T) {
	// Commitment marker ("Hypothesis 1:") precedes both evidence and conclusion.
	text := "Hypothesis 1: the deploy is at fault.\nThe logs show a spike at 15:02.\nRoot cause: the cache."
	if got := Score(text); got != "cleared" {
		t.Errorf("Score = %q, want cleared", got)
	}
}

func TestScoreViolatedConclusionFirst(t *testing.T) {
	// A conclusion ("caused by") with no prior commitment marker.
	text := "The outage was caused by a memory error in the fraud service."
	if got := Score(text); got != "violated" {
		t.Errorf("Score = %q, want violated", got)
	}
}

func TestScoreViolatedCommitmentAfterEvidence(t *testing.T) {
	// The commitment appears, but only after evidence — post-hoc, so VIOLATED.
	text := "The logs show entries at 15:03. Hypothesis 1: something broke."
	if got := Score(text); got != "violated" {
		t.Errorf("Score = %q, want violated", got)
	}
}

func TestScoreNoMarkersAtAll(t *testing.T) {
	// No commitment marker anywhere: c stays at +inf, so c<e and c<k both fail.
	if got := Score("Just some neutral prose with nothing notable."); got != "violated" {
		t.Errorf("Score = %q, want violated", got)
	}
}

func TestScoreMarkdownNormalisation(t *testing.T) {
	// Emphasis/header/code chars are stripped, so "**Hypothesis:**" is recognised.
	text := "**Hypothesis:** the deploy. Then `root cause: cache` and .log noise."
	if got := Score(text); got != "cleared" {
		t.Errorf("Score = %q, want cleared", got)
	}
}

func TestNormalizeStripsMarkers(t *testing.T) {
	got := normalize("a*b#c`d")
	if got != "abcd" {
		t.Errorf("normalize = %q, want abcd", got)
	}
}

func TestFirstPosNoMatch(t *testing.T) {
	if got := firstPos("nothing here", commitmentRE); got != posInf {
		t.Errorf("firstPos with no match = %d, want posInf", got)
	}
}

func TestValidateAllAgree(t *testing.T) {
	// A reader that returns text matching each hand score exactly.
	res, err := Validate(func(_, _, _ string, _ int) (string, error) {
		// Return a cleared response; the check below picks per-cell truth instead.
		return "", nil
	})
	if err != nil {
		t.Fatal(err)
	}
	// With every read returning "" (violated), agreement equals the count of
	// hand-VIOLATED cells.
	wantViol := 0
	total := 0
	for _, hs := range HandSets {
		for _, v := range hs.Violated {
			total++
			if v {
				wantViol++
			}
		}
	}
	if res.Total != total {
		t.Errorf("Total = %d, want %d", res.Total, total)
	}
	if res.Agree != wantViol {
		t.Errorf("Agree = %d, want %d (hand-violated cells)", res.Agree, wantViol)
	}
	if len(res.Mismatches) != total-wantViol {
		t.Errorf("Mismatches = %d, want %d", len(res.Mismatches), total-wantViol)
	}
}

func TestValidatePerfectAgreement(t *testing.T) {
	// Return a cleared response for hand-cleared cells and a violated one for
	// hand-violated cells, so every cell agrees and there are no mismatches.
	cleared := "Hypothesis 1: a guess before any evidence."
	violated := "caused by the cache error."
	// Build a lookup of the hand truth per (model,cond,run).
	truth := map[string]bool{}
	for _, hs := range HandSets {
		for i, v := range hs.Violated {
			truth[key(hs.Model, hs.Condition, i+1)] = v
		}
	}
	res, err := Validate(func(model, _, cond string, run int) (string, error) {
		if truth[key(model, cond, run)] {
			return violated, nil
		}
		return cleared, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Agree != res.Total || len(res.Mismatches) != 0 {
		t.Errorf("expected perfect agreement, got %d/%d with %d mismatches",
			res.Agree, res.Total, len(res.Mismatches))
	}
}

func TestValidateReadError(t *testing.T) {
	_, err := Validate(func(_, _, _ string, _ int) (string, error) {
		return "", fmt.Errorf("boom")
	})
	if err == nil {
		t.Error("expected read error to propagate")
	}
}

func TestValidateMismatchShape(t *testing.T) {
	// Force one specific mismatch: make the very first cell (mistral baseline run 1,
	// hand VIOLATED) score cleared, everything else agree.
	cleared := "Hypothesis 1: a guess."
	violated := "caused by the cache."
	truth := map[string]bool{}
	for _, hs := range HandSets {
		for i, v := range hs.Violated {
			truth[key(hs.Model, hs.Condition, i+1)] = v
		}
	}
	res, err := Validate(func(model, _, cond string, run int) (string, error) {
		if model == "mistral" && cond == "baseline" && run == 1 {
			return cleared, nil // hand is VIOLATED -> mismatch
		}
		if truth[key(model, cond, run)] {
			return violated, nil
		}
		return cleared, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Mismatches) != 1 {
		t.Fatalf("want 1 mismatch, got %d", len(res.Mismatches))
	}
	m := res.Mismatches[0]
	if m.Model != "mistral" || m.Condition != "baseline" || m.Run != 1 ||
		m.Hand != "VIOLATED" || m.Got != "cleared" {
		t.Errorf("unexpected mismatch shape: %+v", m)
	}
}

func TestViolatedLabel(t *testing.T) {
	if violatedLabel(true) != "VIOLATED" || violatedLabel(false) != "cleared" {
		t.Error("violatedLabel wrong")
	}
}

func TestRep(t *testing.T) {
	got := rep(true, 3)
	if len(got) != 3 || !got[0] || !got[1] || !got[2] {
		t.Errorf("rep(true,3) = %v", got)
	}
}

func key(model, cond string, run int) string {
	return fmt.Sprintf("%s|%s|%d", model, cond, run)
}
