package battery

import "testing"

func TestEvidenceStateValid(t *testing.T) {
	for _, s := range []EvidenceState{StateCandidate, StateCertified, StateDerivedView, StateAuditSnapshot} {
		if !s.Valid() {
			t.Errorf("%q should be valid", s)
		}
	}
	if EvidenceState("nonsense").Valid() {
		t.Error("an unknown state should not be valid")
	}
}

func TestPromoteOn(t *testing.T) {
	promote := RunVerdict{Kind: RunPromote}
	block := RunVerdict{Kind: RunBlock, Failures: []Failure{{Item: 3, Reason: "x"}}}
	defer_ := RunVerdict{Kind: RunDefer, Pending: []string{"item 5"}}

	cases := []struct {
		name    string
		from    EvidenceState
		v       RunVerdict
		want    EvidenceState
		wantErr bool
	}{
		{"candidate promotes to certified", StateCandidate, promote, StateCertified, false},
		{"candidate defer stays candidate", StateCandidate, defer_, StateCandidate, false},
		{"candidate block stays candidate", StateCandidate, block, StateCandidate, false},
		{"certified promote recertifies", StateCertified, promote, StateCertified, false},
		{"certified block demotes", StateCertified, block, StateCandidate, false},
		{"certified defer stays certified", StateCertified, defer_, StateCertified, false},
		{"derived view is not certifiable", StateDerivedView, promote, StateDerivedView, true},
		{"audit snapshot is not certifiable", StateAuditSnapshot, promote, StateAuditSnapshot, true},
		{"unknown state errors", EvidenceState("bogus"), promote, EvidenceState("bogus"), true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, note, err := PromoteOn(c.from, c.v)
			if (err != nil) != c.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, c.wantErr)
			}
			if got != c.want {
				t.Errorf("state = %q, want %q", got, c.want)
			}
			if !c.wantErr && note == "" {
				t.Error("a successful transition should carry a note")
			}
		})
	}
}
