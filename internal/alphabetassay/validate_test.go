package alphabetassay

import (
	"strings"
	"testing"
)

func TestCheckCollectInputsParseMiss(t *testing.T) {
	err := CheckCollectInputs(4, 0, 0)
	if err == nil {
		t.Fatal("want error on dirs-found-but-none-parsed, got nil")
	}
	if !strings.Contains(err.Error(), "ParseRunDir") || !strings.Contains(err.Error(), "4") {
		t.Fatalf("error should name ParseRunDir and the dir count; got %q", err)
	}
}

func TestCheckCollectInputsEmpty(t *testing.T) {
	// No response dirs matched at all: zero collected -> error, not "done".
	err := CheckCollectInputs(0, 0, 0)
	if err == nil {
		t.Fatal("want error on zero collected responses, got nil")
	}
	if !strings.Contains(err.Error(), "0 responses") {
		t.Fatalf("error should name the empty result; got %q", err)
	}
}

func TestCheckCollectInputsValid(t *testing.T) {
	if err := CheckCollectInputs(2, 2, 96); err != nil {
		t.Fatalf("want nil on a valid collect, got %q", err)
	}
}
