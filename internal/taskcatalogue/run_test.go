package taskcatalogue

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
)

func TestReadLabelsLastWins(t *testing.T) {
	src := `{"id":1,"status":"error","attempts":1}

{"id":2,"status":"ok","role":"opener","kind":"other","attempts":1}
{"id":1,"status":"ok","role":"approve","kind":"none","attempts":1}
`
	got, err := ReadLabels(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[1].Status != StatusOK || got[1].Role != "approve" {
		t.Errorf("got %+v", got)
	}
	if _, err := ReadLabels(strings.NewReader("{bad\n")); err == nil {
		t.Error("malformed: want error")
	}
	if _, err := ReadLabels(strings.NewReader(strings.Repeat("a", 5*1024*1024))); err == nil {
		t.Error("over-long line: want error")
	}
}

func TestPending(t *testing.T) {
	ps := []Prompt{{ID: 1}, {ID: 2}, {ID: 3}}
	done := map[int]PromptLabel{1: {Status: StatusOK}, 2: {Status: StatusInvalid}}
	if got := Pending(ps, done, false); len(got) != 1 || got[0].ID != 3 {
		t.Errorf("no retry: %v", got)
	}
	if got := Pending(ps, done, true); len(got) != 2 || got[0].ID != 2 {
		t.Errorf("retry: %v", got)
	}
}

func TestLabelAll(t *testing.T) {
	tax := mustTax(t)
	ps := []Prompt{{ID: 1, Text: "a"}, {ID: 2, Text: "b"}, {ID: 3, Text: "c"}}
	var buf bytes.Buffer
	var seen []int
	n, err := LabelAll(context.Background(), Classifier{Client: &seqClient{}, Tax: tax}, ps, 0, &buf, func(d int) { seen = append(seen, d) })
	if err != nil || n != 3 || len(seen) != 3 {
		t.Fatalf("n=%d err=%v seen=%v", n, err, seen)
	}
	labels, err := ReadLabels(&buf)
	if err != nil || len(labels) != 3 {
		t.Fatalf("labels %v %v", labels, err)
	}
}

type failWriter struct{}

func (failWriter) Write([]byte) (int, error) { return 0, errors.New("disk full") }

func TestLabelAllWriteError(t *testing.T) {
	tax := mustTax(t)
	ps := []Prompt{{ID: 1}, {ID: 2}}
	n, err := LabelAll(context.Background(), Classifier{Client: &seqClient{}, Tax: tax}, ps, 2, failWriter{}, nil)
	if err == nil || n != 0 || !strings.Contains(err.Error(), "disk full") {
		t.Errorf("n=%d err=%v", n, err)
	}
}

func TestLabelAllCancelled(t *testing.T) {
	tax := mustTax(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var buf bytes.Buffer
	c := &seqClient{err: context.Canceled}
	n, err := LabelAll(ctx, Classifier{Client: c, Tax: tax}, []Prompt{{ID: 1}, {ID: 2}}, 1, &buf, nil)
	if !errors.Is(err, context.Canceled) || n != 0 || buf.Len() != 0 {
		t.Errorf("n=%d err=%v buf=%q", n, err, buf.String())
	}
}
