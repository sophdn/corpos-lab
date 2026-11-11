package taskcatalogue

import (
	"strings"
	"testing"
)

const historyFixture = `{"display":"please fix the widget","pastedContents":{},"timestamp":1700000000000,"project":"/home/u/dev","sessionId":"s1"}

{"display":"yes please","pastedContents":{},"timestamp":1700000001000,"project":"/home/u/dev","sessionId":"s1"}
{"display":"/mcp reconnect","pastedContents":{},"timestamp":1700000002000,"project":"/home/u/dev","sessionId":"s2"}
{"display":"look at this [Pasted text #1]","pastedContents":{"2":{"id":2,"type":"text","contentHash":"abc"},"1":{"id":1,"type":"text","content":"inline log"},"3":{"id":3,"type":"image"},"4":{"id":4,"type":"text","content":"   "},"5":{"id":5,"type":"text","contentHash":"gone"}},"timestamp":1700000003000,"project":"/home/u/dev","sessionId":"s1"}
`

func TestParseHistory(t *testing.T) {
	lookup := func(h string) (string, bool) {
		if h == "abc" {
			return "cached paste", true
		}
		return "", false
	}
	ps, err := ParseHistory(strings.NewReader(historyFixture), lookup)
	if err != nil {
		t.Fatal(err)
	}
	if len(ps) != 4 {
		t.Fatalf("got %d prompts, want 4", len(ps))
	}
	if ps[1].ID != 2 || ps[2].ID != 3 || ps[3].ID != 4 {
		t.Errorf("IDs must equal line indexes (blank line consumes one): %d %d %d", ps[1].ID, ps[2].ID, ps[3].ID)
	}
	if ps[0].Prev != "" || ps[1].Prev != "please fix the widget" {
		t.Errorf("prev wrong: %q %q", ps[0].Prev, ps[1].Prev)
	}
	if ps[3].Prev != "yes please" {
		t.Errorf("prev must stay within the session: %q", ps[3].Prev)
	}
	if !ps[2].Slash || ps[0].Slash {
		t.Error("slash detection wrong")
	}
	want := "look at this [Pasted text #1]\n[pasted #1] inline log\n[pasted #2] cached paste\n[pasted #5] " + pasteGone
	if ps[3].Text != want {
		t.Errorf("text:\n%q\nwant\n%q", ps[3].Text, want)
	}
	if ps[0].Time.Unix() != 1700000000 || ps[0].CWD != "/home/u/dev" || ps[0].Session != "s1" {
		t.Errorf("fields: %+v", ps[0])
	}
}

func TestParseHistoryWithoutLookupMarksHashedPasteGone(t *testing.T) {
	ps, err := ParseHistory(strings.NewReader(`{"display":"x","pastedContents":{"1":{"type":"text","contentHash":"abc"}},"sessionId":"s"}`), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(ps) != 1 || !strings.HasSuffix(ps[0].Text, pasteGone) {
		t.Fatalf("got %+v", ps)
	}
}

func TestParseHistoryMalformedLineNamesIt(t *testing.T) {
	_, err := ParseHistory(strings.NewReader("{\"display\":\"a\"}\n{nope\n"), nil)
	if err == nil || !strings.Contains(err.Error(), "line 1") {
		t.Fatalf("want error naming line 1, got %v", err)
	}
}

func TestParseHistoryLineTooLong(t *testing.T) {
	long := `{"display":"` + strings.Repeat("a", 17*1024*1024) + `"}`
	if _, err := ParseHistory(strings.NewReader(long), nil); err == nil {
		t.Fatal("want a read error for a line over the scanner cap")
	}
}

func TestCapRunes(t *testing.T) {
	if got := capRunes("héllo", 10); got != "héllo" {
		t.Errorf("short: %q", got)
	}
	if got := capRunes("héllo", 2); got != "hé…" {
		t.Errorf("cut: %q", got)
	}
}
