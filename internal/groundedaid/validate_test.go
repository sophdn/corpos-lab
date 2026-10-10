package groundedaid

import (
	"strings"
	"testing"
)

func TestCheckBuildInputsGlobMiss(t *testing.T) {
	// No run dirs found by the glob (the scramble study's scb-* vs gnp-*): zero
	// rows built -> error naming the empty result, not a success-shaped summary.
	err := CheckBuildInputs(0, 0, nil)
	if err == nil {
		t.Fatal("want error on 0 dirs / 0 rows, got nil")
	}
	if !strings.Contains(err.Error(), "0 items") {
		t.Fatalf("error should name the zero item count; got %q", err)
	}
}

func TestCheckBuildInputsParseMiss(t *testing.T) {
	// Glob matched run dirs but ParseRunDir accepted none of them.
	err := CheckBuildInputs(3, 0, nil)
	if err == nil {
		t.Fatal("want error on dirs-found-but-none-parsed, got nil")
	}
	if !strings.Contains(err.Error(), "ParseRunDir") || !strings.Contains(err.Error(), "3") {
		t.Fatalf("error should name ParseRunDir and the dir count; got %q", err)
	}
}

func TestCheckBuildInputsUnknownClass(t *testing.T) {
	// Rows parsed, but a class outside the reported Classes set would be silently
	// dropped by the per-class slice loop.
	rows := []RawRow{
		{Cls: Classes[0], Cond: "baseline"},
		{Cls: "made-up-class", Cond: "baseline"},
		{Cls: "another-unknown", Cond: "baseline"},
	}
	err := CheckBuildInputs(2, 2, rows)
	if err == nil {
		t.Fatal("want error on an unknown class, got nil")
	}
	if !strings.Contains(err.Error(), "made-up-class") || !strings.Contains(err.Error(), "another-unknown") {
		t.Fatalf("error should name the unknown classes; got %q", err)
	}
}

func TestCheckBuildInputsValid(t *testing.T) {
	// Every row's class in Classes, at least one row: no error, behavior unchanged.
	rows := []RawRow{{Cls: Classes[0], Cond: "baseline"}}
	if err := CheckBuildInputs(1, 1, rows); err != nil {
		t.Fatalf("want nil on a valid build, got %q", err)
	}
}

func TestCheckAggregateKeyEmpty(t *testing.T) {
	if err := CheckAggregateKey(map[string]KeyEntry{}); err == nil {
		t.Fatal("want error on an empty key, got nil")
	}
}

func TestCheckAggregateKeyBadCondition(t *testing.T) {
	key := map[string]KeyEntry{
		"r0": {Cls: Classes[0], Condition: Conds[0]},
		"r1": {Cls: Classes[0], Condition: "scrambled_glyph"},
	}
	err := CheckAggregateKey(key)
	if err == nil {
		t.Fatal("want error on a condition absent from Conds, got nil")
	}
	if !strings.Contains(err.Error(), "scrambled_glyph") {
		t.Fatalf("error should name the dropped condition; got %q", err)
	}
}

func TestCheckAggregateKeyBadClass(t *testing.T) {
	key := map[string]KeyEntry{
		"r0": {Cls: "not-a-real-class", Condition: Conds[0]},
	}
	err := CheckAggregateKey(key)
	if err == nil {
		t.Fatal("want error on a class absent from Classes, got nil")
	}
	if !strings.Contains(err.Error(), "not-a-real-class") {
		t.Fatalf("error should name the dropped class; got %q", err)
	}
}

func TestCheckAggregateKeyValid(t *testing.T) {
	key := map[string]KeyEntry{
		"r0": {Cls: Classes[0], Condition: Conds[0]},
		"r1": {Cls: Classes[1], Condition: Conds[1]},
	}
	if err := CheckAggregateKey(key); err != nil {
		t.Fatalf("want nil on a key with only reported classes/conds, got %q", err)
	}
}
