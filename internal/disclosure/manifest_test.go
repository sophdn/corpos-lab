package disclosure

import (
	"strings"
	"testing"
)

func TestParseApprovedReadsSlugsInOrder(t *testing.T) {
	text := "# approved public glyphs\n" +
		"casg-direct\n" +
		"\n" +
		"  formal-step-context-bypass  \n" +
		"# a trailing comment\n" +
		"parent-state-check-bypass\n"

	got, err := ParseApproved(text)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "casg-direct,formal-step-context-bypass,parent-state-check-bypass"
	if strings.Join(got, ",") != want {
		t.Errorf("ParseApproved() = %v, want %s", got, want)
	}
}

func TestParseApprovedEmptyIsNoSlugs(t *testing.T) {
	got, err := ParseApproved("# only comments\n\n   \n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("ParseApproved() = %v, want empty", got)
	}
}

func TestParseApprovedRejectsNonBareSlug(t *testing.T) {
	_, err := ParseApproved("casg-direct and friends\n")
	if err == nil {
		t.Fatal("ParseApproved() = nil error, want rejection of a non-bare-slug line")
	}
	if !strings.Contains(err.Error(), "line 1") {
		t.Errorf("error = %q, want it to name line 1", err.Error())
	}
}
