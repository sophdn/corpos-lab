package testutil

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"corpos-lab/internal/model"
)

func TestWriteStudyDir(t *testing.T) {
	dir := WriteStudyDir(t, map[string]int{"n": 1}, map[string]string{"a.md": "A"})
	for name, want := range map[string]string{"study.json": `{"n":1}`, "a.md": "A"} {
		b, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil || string(b) != want {
			t.Errorf("%s = %q, %v; want %q", name, b, err, want)
		}
	}
}

func TestNeutralSampler(t *testing.T) {
	type sampler struct {
		Temperature   float64 `json:"temperature"`
		Seeds         []int   `json:"seeds"`
		MinP          float64 `json:"min_p"`
		RepeatPenalty float64 `json:"repeat_penalty"`
		TopNSigma     float64 `json:"top_n_sigma"`
	}
	s := NeutralSampler[sampler](0.8, []int{1, 2})
	if s.Temperature != 0.8 || len(s.Seeds) != 2 || s.MinP != 0.05 || s.RepeatPenalty != 1.0 || s.TopNSigma != -1.0 {
		t.Errorf("NeutralSampler = %+v", s)
	}
	defer func() {
		if recover() == nil {
			t.Error("a type the chain cannot decode into should panic")
		}
	}()
	NeutralSampler[int](0, nil)
}

func TestFakeClient(t *testing.T) {
	ctx := context.Background()
	f := &FakeClient{Text: "t"}
	if r, err := f.Generate(ctx, "p", model.GenParams{}); err != nil || r.Text != "t" {
		t.Errorf("text reply = %+v, %v", r, err)
	}
	f.Resp = &model.Response{Text: "r"}
	if r, _ := f.Generate(ctx, "q", model.GenParams{}); r.Text != "r" {
		t.Errorf("resp reply = %+v", r)
	}
	f.Err = errors.New("boom")
	if _, err := f.Generate(ctx, "s", model.GenParams{}); err == nil {
		t.Error("Err not returned")
	}
	if len(f.Prompts) != 3 || len(f.Params) != 3 {
		t.Errorf("recorded %d prompts, %d params", len(f.Prompts), len(f.Params))
	}
	if f.Name() != "fake-model" || (&FakeClient{ModelName: "m"}).Name() != "m" || f.Version() != "0.0.0" {
		t.Error("identity")
	}
	f.ServerProps = model.ServerProps{BuildInfo: "b"}
	if p, err := f.Props(ctx); err != nil || p.BuildInfo != "b" {
		t.Errorf("Props = %+v, %v", p, err)
	}
	f.PropsErr = errors.New("down")
	if _, err := f.Props(ctx); err == nil {
		t.Error("PropsErr not returned")
	}
}

func TestWriteFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "a", "b", "c.txt")
	WriteFile(t, path, "body")
	if b, err := os.ReadFile(path); err != nil || string(b) != "body" {
		t.Errorf("WriteFile wrote %q, %v", b, err)
	}
}

func TestShuffles(t *testing.T) {
	xs := []int{1, 2, 3}
	swap := func(i, j int) { xs[i], xs[j] = xs[j], xs[i] }
	IdentityShuffle(len(xs), swap)
	ReverseShuffle(len(xs), swap)
	if xs[0] != 3 || xs[1] != 2 || xs[2] != 1 {
		t.Errorf("shuffled = %v, want [3 2 1]", xs)
	}
}

func TestScriptClient(t *testing.T) {
	ctx := context.Background()
	c := &ScriptClient{Replies: []string{"a", "b"}, Truncated: []bool{false, true}, Fallback: "z", Reply: model.Response{Model: "m"}}
	var got []string
	for range 3 {
		r, err := c.Generate(ctx, "p", model.GenParams{})
		if err != nil || r.Model != "m" {
			t.Fatalf("Generate = %+v, %v", r, err)
		}
		got = append(got, r.Text+map[bool]string{true: "!", false: ""}[r.Truncated])
	}
	if strings.Join(got, ",") != "a,b!,z" || len(c.Prompts) != 3 {
		t.Errorf("turns = %v, prompts %d", got, len(c.Prompts))
	}
	if c.Name() != "m" || c.Version() != "v0" {
		t.Error("identity")
	}
	if _, err := c.Props(ctx); err != nil {
		t.Error(err)
	}
	c.Err = errors.New("boom")
	if _, err := c.Generate(ctx, "p", model.GenParams{}); err == nil || len(c.Prompts) != 3 {
		t.Error("Err must fail the call before recording")
	}
}

func TestUnderFile(t *testing.T) {
	if err := os.MkdirAll(UnderFile(t, "work"), 0o750); err == nil {
		t.Error("MkdirAll under a file should fail")
	}
}
