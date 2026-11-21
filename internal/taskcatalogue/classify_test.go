package taskcatalogue

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"

	"corpos-lab/internal/model"
)

const taxFixture = `
ignore = ["vault"]

[[role]]
name = "opener"
opener = true
description = "starts new work"

[[role]]
name = "approve"
description = "says yes"

[[role]]
name = "control"
description = "a harness command"

[[kind]]
name = "build-thing"
description = "build a thing"

[[kind]]
name = "other"
description = "fits nothing"

[[alias]]
text = "nick"
project = "alpha"
`

func mustTax(t *testing.T) Taxonomy {
	t.Helper()
	tax, err := ParseTaxonomy(taxFixture)
	if err != nil {
		t.Fatal(err)
	}
	return tax
}

func TestParseTaxonomy(t *testing.T) {
	tax := mustTax(t)
	if !tax.IsOpener("opener") || tax.IsOpener("approve") || tax.IsOpener("missing") {
		t.Error("IsOpener wrong")
	}
	if tax.KindDescription("build-thing") != "build a thing" || tax.KindDescription("nope") != "" {
		t.Error("KindDescription wrong")
	}
	if len(tax.Ignore) != 1 || tax.Ignore[0] != "vault" {
		t.Errorf("ignore %v", tax.Ignore)
	}
	if len(tax.Aliases) != 1 || tax.Aliases[0].Project != "alpha" {
		t.Errorf("aliases %v", tax.Aliases)
	}
	bad := map[string]string{
		"syntax":        "[[role]\n",
		"empty":         "",
		"no opener":     "[[role]]\nname=\"a\"\n[[kind]]\nname=\"k\"\n",
		"dup":           "[[role]]\nname=\"a\"\nopener=true\n[[kind]]\nname=\"a\"\n",
		"reserved none": "[[role]]\nname=\"a\"\nopener=true\n[[kind]]\nname=\"none\"\n",
		"space":         "[[role]]\nname=\"a b\"\nopener=true\n[[kind]]\nname=\"k\"\n",
		"bad kind":      "[[role]]\nname=\"a\"\nopener=true\n[[kind]]\nname=\"\"\n",
	}
	for name, src := range bad {
		if _, err := ParseTaxonomy(src); err == nil {
			t.Errorf("%s: want error", name)
		}
	}
}

func TestBuildPrompt(t *testing.T) {
	tax := mustTax(t)
	first := BuildPrompt(tax, Prompt{Text: "make a thing"})
	if !strings.Contains(first, "- build-thing: build a thing") || !strings.Contains(first, "first message of the session") ||
		!strings.Contains(first, "<<<make a thing>>>") || !strings.Contains(first, "- none:") {
		t.Errorf("prompt missing parts:\n%s", first)
	}
	later := BuildPrompt(tax, Prompt{Text: "yes", Prev: "make a thing"})
	if !strings.Contains(later, "same session: <<<make a thing>>>") {
		t.Errorf("prev missing:\n%s", later)
	}
	if strings.Index(later, "<<<yes>>>") < strings.Index(later, "Roles:") {
		t.Error("the variable text must come after the shared prefix")
	}
}

func TestParseLabel(t *testing.T) {
	tax := mustTax(t)
	role, kind, ask, err := ParseLabel(tax, "sure:\n{\"role\": \" opener \", \"kind\": \"build-thing\", \"ask\": \" build it \"}\n")
	if err != nil || role != "opener" || kind != "build-thing" || ask != "build it" {
		t.Errorf("got %q %q %q %v", role, kind, ask, err)
	}
	if _, kind, _, err := ParseLabel(tax, `{"role":"approve","kind":"none"}`); err != nil || kind != KindNone {
		t.Errorf("none kind: %v %v", kind, err)
	}
	for _, bad := range []string{"no json", "} {", `{"role":`, `{"role":"boss","kind":"other"}`, `{"role":"opener","kind":"dance"}`} {
		if _, _, _, err := ParseLabel(tax, bad); err == nil {
			t.Errorf("%q: want error", bad)
		}
	}
}

// seqClient answers from a list, safely across goroutines.
type seqClient struct {
	mu      sync.Mutex
	answers []string
	err     error
	prompts []string
}

func (c *seqClient) Generate(_ context.Context, prompt string, _ model.GenParams) (model.Response, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.prompts = append(c.prompts, prompt)
	if c.err != nil {
		return model.Response{}, c.err
	}
	text := `{"role":"opener","kind":"build-thing","ask":"x"}`
	if len(c.answers) > 0 {
		text, c.answers = c.answers[0], c.answers[1:]
	}
	return model.Response{Text: text, Model: "m", SystemFingerprint: "b1", Reasoning: "", Timings: model.Timings{PredictedPerSecond: 42}}, nil
}
func (c *seqClient) Name() string    { return "m" }
func (c *seqClient) Version() string { return "v" }
func (c *seqClient) Props(context.Context) (model.ServerProps, error) {
	return model.ServerProps{ModelAlias: "m"}, nil
}

func TestClassify(t *testing.T) {
	tax := mustTax(t)
	ctx := context.Background()

	c := &seqClient{}
	got := Classifier{Client: c, Tax: tax}.Classify(ctx, Prompt{ID: 7, Text: "/mcp reconnect", Slash: true})
	if got.Role != RoleControl || got.Kind != KindNone || got.Ask != "/mcp" || len(c.prompts) != 0 {
		t.Errorf("slash: %+v (calls %d)", got, len(c.prompts))
	}

	got = Classifier{Client: &seqClient{}, Tax: tax}.Classify(ctx, Prompt{ID: 1, Text: "build"})
	if got.Status != StatusOK || got.Attempts != 1 || got.Model != "m" || got.TokPerSec != 42 {
		t.Errorf("ok: %+v", got)
	}

	c = &seqClient{answers: []string{"garbage"}}
	got = Classifier{Client: c, Tax: tax}.Classify(ctx, Prompt{ID: 2, Text: "build"})
	if got.Status != StatusOK || got.Attempts != 2 || !strings.HasSuffix(c.prompts[1], retrySuffix) {
		t.Errorf("retry: %+v", got)
	}

	got = Classifier{Client: &seqClient{answers: []string{"garbage", "worse"}}, Tax: tax}.Classify(ctx, Prompt{ID: 3, Text: "build"})
	if got.Status != StatusInvalid || got.Attempts != 2 || !strings.Contains(got.Detail, "worse") {
		t.Errorf("invalid: %+v", got)
	}

	got = Classifier{Client: &seqClient{err: errors.New("down")}, Tax: tax}.Classify(ctx, Prompt{ID: 4, Text: "build"})
	if got.Status != StatusError || got.Detail != "down" {
		t.Errorf("error: %+v", got)
	}
}

func TestLabelParamsPinsEveryStage(t *testing.T) {
	p := LabelParams()
	if p.Temperature == nil || p.MaxTokens == nil || p.Seed == nil || p.TopK == nil || p.TopP == nil ||
		p.MinP == nil || p.TypicalP == nil || p.TopNSigma == nil || p.RepeatPenalty == nil ||
		p.RepeatLastN == nil || p.PresencePenalty == nil || p.FrequencyPenalty == nil ||
		p.XTCProbability == nil || p.DryMultiplier == nil {
		t.Errorf("unpinned stage: %+v", p)
	}
}

func TestFirstWord(t *testing.T) {
	if firstWord("  ") != "" || firstWord("/a b") != "/a" {
		t.Error("firstWord wrong")
	}
}
