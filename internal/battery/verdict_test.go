package battery

import (
	"encoding/json"
	"testing"
)

// Characterization tests ported from lab-app-types/src/verdict.rs (18 tests);
// see docs/PORT_BATTERY_INVENTORY.md for the mapping.

func intp(n int) *int { return &n }

func TestRouteChainFailure(t *testing.T) {
	failure := func(item int, class FailureClassKind, detail string) Failure {
		return Failure{Item: item, Reason: "test failure", Class: FailureClass{Kind: class, Detail: detail}}
	}
	cases := []struct {
		name     string
		failures []Failure
		want     FailureRoute
	}{
		{"definition gap stops", []Failure{failure(3, ClassDefinitionGap, "")}, RouteStopAndStudyDefinition},
		{"executor correctable requeues", []Failure{failure(5, ClassExecutorCorrectable, "")}, RouteRequeueToResearcher},
		{"mixed failures definition gap takes precedence", []Failure{
			failure(2, ClassExecutorCorrectable, ""),
			failure(7, ClassDefinitionGap, ""),
		}, RouteStopAndStudyDefinition},
		{"unknown classification treated as definition gap", []Failure{failure(4, ClassUnknown, "no data")}, RouteStopAndStudyDefinition},
		{"empty failures requeue", nil, RouteRequeueToResearcher},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := RouteChainFailure(tc.failures); got != tc.want {
				t.Fatalf("RouteChainFailure = %s, want %s", got, tc.want)
			}
		})
	}
}

func TestVerdictDisplay(t *testing.T) {
	cases := []struct {
		name string
		v    Verdict
		want string
	}{
		{"pass", Pass(), "PASS"},
		{"fail with item", FailItem(3, "missing axis"), "FAIL at item 3: missing axis"},
		{"fail without item", Fail("model timeout"), "FAIL: model timeout"},
		{"pass with condition", PassWithCondition("pending review"), "PASS* (pending review)"},
		{"flag", Flag("ambiguous"), "FLAG: ambiguous"},
		{"deferred", Deferred("gather more data"), "DEFERRED: gather more data"},
		{"not applicable", NotApplicable("item 4 only applies to TABOO_REGISTRY entries"), "N/A: item 4 only applies to TABOO_REGISTRY entries"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.v.String(); got != tc.want {
				t.Fatalf("String() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestPassedHelperCoversPassingVariants(t *testing.T) {
	if !Pass().Passed() {
		t.Error("Pass should count as passed")
	}
	if !PassWithCondition("c").Passed() {
		t.Error("PassWithCondition should count as passed")
	}
	if !NotApplicable("n/a").Passed() {
		t.Error("NotApplicable should count as passed")
	}
	if Fail("r").Passed() {
		t.Error("Fail should not count as passed")
	}
	if Flag("r").Passed() {
		t.Error("Flag should not count as passed")
	}
	if Deferred("p").Passed() {
		t.Error("Deferred should not count as passed")
	}
}

func TestVerdictRoundTripsThroughJSON(t *testing.T) {
	cases := []Verdict{
		Pass(),
		PassWithCondition("needs reviewer ack"),
		Flag("ambiguous"),
		Deferred("gather more data"),
		FailItem(7, "missing fallout profile"),
		Fail("generic failure"),
		NotApplicable("n/a for glyph entries"),
	}
	for _, v := range cases {
		raw, err := json.Marshal(v)
		if err != nil {
			t.Fatalf("marshal %v: %v", v, err)
		}
		var back Verdict
		if err := json.Unmarshal(raw, &back); err != nil {
			t.Fatalf("unmarshal %s: %v", raw, err)
		}
		if v.Kind != back.Kind || v.Reason != back.Reason || v.Pending != back.Pending ||
			v.Condition != back.Condition || v.Note != back.Note {
			t.Fatalf("round trip for %s changed fields: %+v -> %+v", raw, v, back)
		}
		if (v.Item == nil) != (back.Item == nil) || (v.Item != nil && *v.Item != *back.Item) {
			t.Fatalf("round trip for %s changed item: %v -> %v", raw, v.Item, back.Item)
		}
	}
}

func TestVerdictJSONShapeIsKindTagged(t *testing.T) {
	raw, err := json.Marshal(FailItem(3, "missing axis"))
	if err != nil {
		t.Fatal(err)
	}
	want := `{"kind":"fail","reason":"missing axis","item":3}`
	if string(raw) != want {
		t.Fatalf("JSON shape = %s, want %s", raw, want)
	}
	raw, err = json.Marshal(Pass())
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != `{"kind":"pass"}` {
		t.Fatalf("pass JSON = %s, want {\"kind\":\"pass\"}", raw)
	}
}

func TestComposeAllPassIsPromote(t *testing.T) {
	items := []Verdict{Pass(), PassWithCondition("reviewer ack"), NotApplicable("n/a")}
	got := ComposeRunVerdict(items)
	if got.Kind != RunPromote || !got.IsPromote() {
		t.Fatalf("expected Promote, got %+v", got)
	}
}

func TestComposeWithFailIsBlock(t *testing.T) {
	items := []Verdict{Pass(), FailItem(3, "missing axis"), Deferred("corpus access")}
	got := ComposeRunVerdict(items)
	if got.Kind != RunBlock {
		t.Fatalf("expected Block, got %+v", got)
	}
	if len(got.Failures) != 1 || got.Failures[0].Item != 3 {
		t.Fatalf("expected 1 failure at item 3, got %+v", got.Failures)
	}
}

func TestComposeWithOnlyDeferredIsDefer(t *testing.T) {
	items := []Verdict{Pass(), Deferred("corpus"), Deferred("dep graph")}
	got := ComposeRunVerdict(items)
	if got.Kind != RunDefer || len(got.Pending) != 2 {
		t.Fatalf("expected Defer with 2 pending, got %+v", got)
	}
}

func TestComposeFlagCountsAsBlockSoftFail(t *testing.T) {
	items := []Verdict{Pass(), Flag("ambiguous")}
	got := ComposeRunVerdict(items)
	if got.Kind != RunBlock {
		t.Fatalf("expected Block, got %+v", got)
	}
	if got.Failures[0].Class.Kind != ClassExecutorCorrectable {
		t.Fatalf("flag soft-fail should be executor_correctable, got %+v", got.Failures[0].Class)
	}
	if got.Failures[0].Reason != "FLAG: ambiguous" {
		t.Fatalf("flag reason should carry FLAG: prefix, got %q", got.Failures[0].Reason)
	}
}

func TestComposeFailWithoutItemMapsToItemZeroUnknownClass(t *testing.T) {
	got := ComposeRunVerdict([]Verdict{Fail("no item")})
	if got.Kind != RunBlock || got.Failures[0].Item != 0 {
		t.Fatalf("expected Block with item 0, got %+v", got)
	}
	if got.Failures[0].Class.Kind != ClassUnknown {
		t.Fatalf("expected unknown class, got %+v", got.Failures[0].Class)
	}
}

func TestRunVerdictDisplay(t *testing.T) {
	promote := RunVerdict{Kind: RunPromote}
	if got := promote.String(); got != "PROMOTE" {
		t.Fatalf("got %q", got)
	}
	one := RunVerdict{Kind: RunBlock, Failures: []Failure{{Item: 3, Reason: "r", Class: FailureClass{Kind: ClassDefinitionGap}}}}
	if got := one.String(); got != "BLOCK (1 failure)" {
		t.Fatalf("got %q", got)
	}
	two := RunVerdict{Kind: RunBlock, Failures: []Failure{{}, {}}}
	if got := two.String(); got != "BLOCK (2 failures)" {
		t.Fatalf("got %q", got)
	}
	d := RunVerdict{Kind: RunDefer, Pending: []string{"a", "b"}}
	if got := d.String(); got != "DEFER (2 pending)" {
		t.Fatalf("got %q", got)
	}
}

func TestRunVerdictRoundTripsThroughJSON(t *testing.T) {
	cases := []RunVerdict{
		{Kind: RunPromote},
		{Kind: RunDefer, Pending: []string{"p1"}},
	}
	for _, v := range cases {
		raw, err := json.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		var back RunVerdict
		if err := json.Unmarshal(raw, &back); err != nil {
			t.Fatal(err)
		}
		if back.Kind != v.Kind || len(back.Pending) != len(v.Pending) {
			t.Fatalf("round trip changed %+v -> %+v", v, back)
		}
	}
}

func TestUnknownKindsFallBackSafely(t *testing.T) {
	// A Verdict with an unrecognized kind (e.g. from a newer writer) must
	// not pass, and both String() impls must render the raw kind rather
	// than panic — the Go analogue of Rust's #[non_exhaustive] guarantee.
	bogus := Verdict{Kind: VerdictKind("bogus")}
	if bogus.Passed() {
		t.Error("unknown kind must not count as passed")
	}
	if got := bogus.String(); got != "bogus" {
		t.Fatalf("unknown kind String() = %q, want raw kind", got)
	}
	bogusRun := RunVerdict{Kind: RunVerdictKind("bogus")}
	if got := bogusRun.String(); got != "bogus" {
		t.Fatalf("unknown run kind String() = %q, want raw kind", got)
	}
}

func TestFailItemStoresItemNumber(t *testing.T) {
	v := FailItem(9, "r")
	if v.Item == nil || *v.Item != 9 {
		t.Fatalf("expected item 9, got %v", v.Item)
	}
	if got := intp(9); *got != 9 {
		t.Fatal("helper sanity")
	}
}
