package testutil

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"corpos-lab/internal/model"
)

// WriteFiles writes each name -> content into dir (mode 0644), failing the
// test on an error.
func WriteFiles(t testing.TB, dir string, files map[string]string) {
	t.Helper()
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// WriteStudyDir lays out a container /in dir in a fresh temp dir: spec as
// study.json beside the materials. It returns the dir.
func WriteStudyDir(t testing.TB, spec any, materials map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	raw, err := json.Marshal(spec)
	if err != nil {
		t.Fatal(err)
	}
	WriteFiles(t, dir, map[string]string{"study.json": string(raw)})
	WriteFiles(t, dir, materials)
	return dir
}

// NeutralSampler returns a complete, neutral sampler chain as T (a sampler
// struct carrying the study's json tags, such as assay.Sampling): temp and
// seeds as given, max_tokens 512, min_p 0.05 the only live truncation stage,
// and every other stage at its disabled value.
func NeutralSampler[T any](temp float64, seeds []int) T {
	raw, err := json.Marshal(map[string]any{
		"temperature": temp, "seeds": seeds, "max_tokens": 512,
		"top_n_sigma": -1.0, "top_k": 0, "typical_p": 1.0, "top_p": 1.0, "min_p": 0.05,
		"repeat_penalty": 1.0, "repeat_last_n": 0, "presence_penalty": 0.0, "frequency_penalty": 0.0,
		"xtc_probability": 0.0, "dry_multiplier": 0.0,
	})
	var s T
	if err == nil {
		err = json.Unmarshal(raw, &s)
	}
	if err != nil {
		panic("testutil: NeutralSampler: " + err.Error())
	}
	return s
}

// FakeClient is a scripted model.Client. Generate records each prompt and
// GenParams, then returns Err if set, else *Resp if set, else a bare Text
// reply. Name is ModelName, or "fake-model" when that is empty; Props returns
// PropsErr if set, else ServerProps.
type FakeClient struct {
	Text        string
	Err         error
	ModelName   string
	Resp        *model.Response
	ServerProps model.ServerProps
	PropsErr    error

	Prompts []string
	Params  []model.GenParams
}

// Generate records the call and returns the scripted reply.
func (f *FakeClient) Generate(_ context.Context, prompt string, p model.GenParams) (model.Response, error) {
	f.Prompts = append(f.Prompts, prompt)
	f.Params = append(f.Params, p)
	if f.Err != nil {
		return model.Response{}, f.Err
	}
	if f.Resp != nil {
		return *f.Resp, nil
	}
	return model.Response{Text: f.Text}, nil
}

// Name reports ModelName, or "fake-model".
func (f *FakeClient) Name() string {
	if f.ModelName != "" {
		return f.ModelName
	}
	return "fake-model"
}

// Version reports a fixed version.
func (f *FakeClient) Version() string { return "0.0.0" }

// Props returns PropsErr, else ServerProps.
func (f *FakeClient) Props(context.Context) (model.ServerProps, error) {
	if f.PropsErr != nil {
		return model.ServerProps{}, f.PropsErr
	}
	return f.ServerProps, nil
}

// WriteFile writes body to path (mode 0600), first making its parent
// directories (mode 0750), failing the test on an error.
func WriteFile(t testing.TB, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}

// IdentityShuffle is a shuffle that leaves order untouched, so shuffled output
// is deterministic in tests.
func IdentityShuffle(int, func(i, j int)) {}

// ReverseShuffle reverses the n items, a deterministic non-identity reorder.
func ReverseShuffle(n int, swap func(i, j int)) {
	for i, j := 0, n-1; i < j; i, j = i+1, j-1 {
		swap(i, j)
	}
}

// ScriptClient is a scripted model.Client for loop tests. Each call returns
// the next of Replies (Fallback once they run out), marked truncated where the
// index-aligned Truncated says so, and carrying Reply's Model,
// SystemFingerprint, and Timings. It records each prompt. When Err is set,
// every call fails with it before recording anything. Its name is
// Reply.Model.
type ScriptClient struct {
	Replies   []string
	Truncated []bool
	Fallback  string
	Err       error
	Reply     model.Response

	Prompts []string
	turn    int
}

// Generate returns the next scripted turn.
func (c *ScriptClient) Generate(_ context.Context, prompt string, _ model.GenParams) (model.Response, error) {
	if c.Err != nil {
		return model.Response{}, c.Err
	}
	c.Prompts = append(c.Prompts, prompt)
	r := c.Reply
	r.Text = c.Fallback
	if c.turn < len(c.Replies) {
		r.Text = c.Replies[c.turn]
	}
	r.Truncated = c.turn < len(c.Truncated) && c.Truncated[c.turn]
	c.turn++
	return r, nil
}

// Name reports Reply.Model.
func (c *ScriptClient) Name() string { return c.Reply.Model }

// Version reports a fixed version.
func (c *ScriptClient) Version() string { return "v0" }

// Props reports an empty server self-report.
func (c *ScriptClient) Props(context.Context) (model.ServerProps, error) {
	return model.ServerProps{}, nil
}

// UnderFile returns a path named name beneath a regular file, so any attempt
// to create a directory there (MkdirAll) fails.
func UnderFile(t testing.TB, name string) string {
	t.Helper()
	blocker := filepath.Join(t.TempDir(), "blocker")
	WriteFiles(t, filepath.Dir(blocker), map[string]string{"blocker": "x"})
	return filepath.Join(blocker, name)
}
