package disclosure

import (
	"strings"
	"testing"
)

// TestSelectFixtureAlphabet is the disclosure invariant Sophi named: a test
// alphabet {A,B,C,D,E} with an approved subset {A,C,D} publishes exactly A, C,
// and D and withholds B and E.
func TestSelectFixtureAlphabet(t *testing.T) {
	certified := []string{"A", "B", "C", "D", "E"}
	approved := []string{"A", "C", "D"}

	sel := Select(certified, approved)

	if err := sel.Err(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got, want := strings.Join(sel.Publish, ","), "A,C,D"; got != want {
		t.Errorf("Publish = %q, want %q", got, want)
	}
	if got, want := strings.Join(sel.Withheld, ","), "B,E"; got != want {
		t.Errorf("Withheld = %q, want %q", got, want)
	}
	if len(sel.Missing) != 0 {
		t.Errorf("Missing = %v, want empty", sel.Missing)
	}
}

func TestSelectApprovedButNotCertifiedIsError(t *testing.T) {
	certified := []string{"A", "B"}
	approved := []string{"A", "F"} // F is not certified

	sel := Select(certified, approved)

	if got, want := strings.Join(sel.Publish, ","), "A"; got != want {
		t.Errorf("Publish = %q, want %q", got, want)
	}
	if got, want := strings.Join(sel.Missing, ","), "F"; got != want {
		t.Errorf("Missing = %q, want %q", got, want)
	}
	err := sel.Err()
	if err == nil {
		t.Fatal("Err() = nil, want error naming the uncertified slug")
	}
	if !strings.Contains(err.Error(), "F") {
		t.Errorf("Err() = %q, want it to name F", err.Error())
	}
}

func TestSelectEmptyApprovedWithholdsEverything(t *testing.T) {
	sel := Select([]string{"A", "B"}, nil)

	if len(sel.Publish) != 0 {
		t.Errorf("Publish = %v, want empty (default closed)", sel.Publish)
	}
	if got, want := strings.Join(sel.Withheld, ","), "A,B"; got != want {
		t.Errorf("Withheld = %q, want %q", got, want)
	}
	if err := sel.Err(); err != nil {
		t.Errorf("Err() = %v, want nil", err)
	}
}

func TestSelectDeduplicatesAndIgnoresEmpty(t *testing.T) {
	sel := Select([]string{"A", "A", "", "B"}, []string{"A", "A", ""})

	if got, want := strings.Join(sel.Publish, ","), "A"; got != want {
		t.Errorf("Publish = %q, want %q", got, want)
	}
	if got, want := strings.Join(sel.Withheld, ","), "B"; got != want {
		t.Errorf("Withheld = %q, want %q", got, want)
	}
	if len(sel.Missing) != 0 {
		t.Errorf("Missing = %v, want empty", sel.Missing)
	}
}

func TestValidatePublicTreeExactMatch(t *testing.T) {
	if err := ValidatePublicTree([]string{"A", "C", "D"}, []string{"A", "C", "D"}); err != nil {
		t.Errorf("ValidatePublicTree() = %v, want nil for an exact match", err)
	}
}

func TestValidatePublicTreeLeak(t *testing.T) {
	// B is present but not approved — a disclosure leak.
	err := ValidatePublicTree([]string{"A", "B", "C"}, []string{"A", "C"})
	if err == nil {
		t.Fatal("ValidatePublicTree() = nil, want a leak error")
	}
	if !strings.Contains(err.Error(), "LEAK") || !strings.Contains(err.Error(), "B") {
		t.Errorf("error = %q, want it to flag LEAK and name B", err.Error())
	}
}

func TestValidatePublicTreeMissing(t *testing.T) {
	// D is approved but absent — promotion did not run.
	err := ValidatePublicTree([]string{"A"}, []string{"A", "D"})
	if err == nil {
		t.Fatal("ValidatePublicTree() = nil, want a missing error")
	}
	if !strings.Contains(err.Error(), "missing") || !strings.Contains(err.Error(), "D") {
		t.Errorf("error = %q, want it to flag missing and name D", err.Error())
	}
}

func TestValidatePublicTreeReportsBothLeakAndMissing(t *testing.T) {
	err := ValidatePublicTree([]string{"A", "B"}, []string{"A", "D"})
	if err == nil {
		t.Fatal("ValidatePublicTree() = nil, want an error")
	}
	if !strings.Contains(err.Error(), "B") || !strings.Contains(err.Error(), "D") {
		t.Errorf("error = %q, want it to name both B (leak) and D (missing)", err.Error())
	}
}
