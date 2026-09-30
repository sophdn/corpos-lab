package actionconflict

import "testing"

func TestIngestBlindMapsVocabulary(t *testing.T) {
	in := []BlindVerdict{
		{ID: "c1::baseline::1", Verdict: "correct"},
		{ID: "c1::canon_conflict::1", Verdict: "harmful"},
		{ID: "c1::scrambled_canon::1", Verdict: "neither"},
		{ID: "c1::off_target_canon::1", Verdict: "unscoreable"},
	}
	got, err := IngestBlind(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := map[string]string{
		"c1::baseline::1":         "A_local",
		"c1::canon_conflict::1":   "A_canon",
		"c1::scrambled_canon::1":  "neither",
		"c1::off_target_canon::1": "unscoreable",
	}
	if len(got) != len(want) {
		t.Fatalf("got %d rows, want %d", len(got), len(want))
	}
	for id, w := range want {
		if got[id] != w {
			t.Errorf("id %s: got %q, want %q", id, got[id], w)
		}
	}
}

func TestAnonIDHidesConditionAndIsStable(t *testing.T) {
	rid := "aia-secfix-qwen38::canon_conflict::1"
	a := AnonID(rid)
	if a == AnonID("aia-secfix-qwen38::baseline::1") {
		t.Fatalf("distinct rids must anonymize distinctly")
	}
	if a != AnonID(rid) {
		t.Fatalf("AnonID must be stable for the same rid")
	}
	for _, leak := range []string{"canon_conflict", "baseline", "canon_aligned", "scrambled", "off_target", "::"} {
		if containsSub(a, leak) {
			t.Fatalf("opaque id %q leaks %q", a, leak)
		}
	}
}

func containsSub(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func TestIngestBlindRejectsBadInput(t *testing.T) {
	cases := []struct {
		name string
		in   []BlindVerdict
	}{
		{"unknown verdict", []BlindVerdict{{ID: "a", Verdict: "maybe"}}},
		{"empty id", []BlindVerdict{{ID: "", Verdict: "correct"}}},
		{"duplicate id", []BlindVerdict{{ID: "a", Verdict: "correct"}, {ID: "a", Verdict: "harmful"}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := IngestBlind(tc.in); err == nil {
				t.Fatalf("expected an error, got nil")
			}
		})
	}
}
