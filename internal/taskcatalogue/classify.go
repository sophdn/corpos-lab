package taskcatalogue

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/BurntSushi/toml"

	"corpos-lab/internal/model"
)

// Label is one entry in a closed list: a role or a task kind.
type Label struct {
	Name        string `toml:"name" json:"name"`
	Description string `toml:"description" json:"description"`
}

// Role is a Label that also says whether a prompt in it opens new work. Only
// openers are counted in the catalogue; the other roles are the follow-up and
// control traffic a multi-turn case replays.
type Role struct {
	Label
	Opener bool `toml:"opener" json:"opener"`
}

// Taxonomy is the closed vocabulary the labeller must choose from, plus extra
// project aliases the ledger does not carry. It is loaded from a file the
// caller supplies (kept with the private data), so no description of anyone's
// work is compiled into this package.
type Taxonomy struct {
	Roles   []Role  `toml:"role"`
	Kinds   []Label `toml:"kind"`
	Aliases []Term  `toml:"alias"`
	// Ignore lists names that are project ids but, in prompts, are mostly an
	// ordinary word (a project called "vault" versus "file it to the vault").
	// They never attribute a prompt.
	Ignore []string `toml:"ignore"`
}

// KindNone is the kind for a prompt that asks for no work (a harness command,
// thanks). It is always allowed and never part of the catalogue.
const KindNone = "none"

// ParseTaxonomy decodes and validates a taxonomy TOML document.
func ParseTaxonomy(src string) (Taxonomy, error) {
	var t Taxonomy
	if _, err := toml.Decode(src, &t); err != nil {
		return Taxonomy{}, fmt.Errorf("taxonomy: %w", err)
	}
	if len(t.Roles) == 0 || len(t.Kinds) == 0 {
		return Taxonomy{}, fmt.Errorf("taxonomy: needs at least one role and one kind")
	}
	seen := map[string]bool{KindNone: true}
	openers := 0
	for _, r := range t.Roles {
		if r.Opener {
			openers++
		}
		if err := checkName("role", r.Name, seen); err != nil {
			return Taxonomy{}, err
		}
	}
	if openers == 0 {
		return Taxonomy{}, fmt.Errorf("taxonomy: no role is marked opener, so nothing would be catalogued")
	}
	for _, k := range t.Kinds {
		if err := checkName("kind", k.Name, seen); err != nil {
			return Taxonomy{}, err
		}
	}
	return t, nil
}

func checkName(what, name string, seen map[string]bool) error {
	if strings.TrimSpace(name) == "" || strings.ContainsAny(name, " \"") {
		return fmt.Errorf("taxonomy: %s name %q must be a non-empty token without spaces or quotes", what, name)
	}
	if seen[name] {
		return fmt.Errorf("taxonomy: name %q is used twice (or is the reserved %q)", name, KindNone)
	}
	seen[name] = true
	return nil
}

// IsOpener reports whether role opens new work.
func (t Taxonomy) IsOpener(role string) bool {
	for _, r := range t.Roles {
		if r.Name == role {
			return r.Opener
		}
	}
	return false
}

func (t Taxonomy) hasRole(name string) bool {
	for _, r := range t.Roles {
		if r.Name == name {
			return true
		}
	}
	return false
}

func (t Taxonomy) hasKind(name string) bool {
	if name == KindNone {
		return true
	}
	for _, k := range t.Kinds {
		if k.Name == name {
			return true
		}
	}
	return false
}

// KindDescription returns a kind's description, or "" for an unknown kind.
func (t Taxonomy) KindDescription(name string) string {
	for _, k := range t.Kinds {
		if k.Name == name {
			return k.Description
		}
	}
	return ""
}

// BuildPrompt renders the labelling prompt. The fixed instructions and the
// vocabulary come first and the prompt-specific text last, so the server's
// prompt cache reuses the shared prefix across calls.
func BuildPrompt(t Taxonomy, p Prompt) string {
	var b strings.Builder
	b.WriteString("You label messages that a person typed to an AI coding agent in a terminal. ")
	b.WriteString("Choose exactly one role and exactly one kind from the lists below.\n\nRoles:\n")
	for _, r := range t.Roles {
		fmt.Fprintf(&b, "- %s: %s\n", r.Name, r.Description)
	}
	b.WriteString("\nKinds (what the message asks the agent to do; for a follow-up, what the follow-up itself asks for):\n")
	for _, k := range t.Kinds {
		fmt.Fprintf(&b, "- %s: %s\n", k.Name, k.Description)
	}
	fmt.Fprintf(&b, "- %s: the message asks for no work at all.\n", KindNone)
	b.WriteString("\nAnswer with one JSON object and nothing else:\n")
	b.WriteString(`{"role": "<role>", "kind": "<kind>", "ask": "<what the message asks for, at most 15 words>"}`)
	b.WriteString("\n\nPrevious message in the same session: ")
	if p.Prev == "" {
		b.WriteString("(none; this is the first message of the session)")
	} else {
		fmt.Fprintf(&b, "<<<%s>>>", p.Prev)
	}
	fmt.Fprintf(&b, "\nMessage to label: <<<%s>>>\n", p.Text)
	return b.String()
}

// retrySuffix is appended once when the first answer is not a valid label.
const retrySuffix = "\nYour previous answer was not one JSON object with a role and a kind from the lists. Answer again with only the JSON object.\n"

type rawLabel struct {
	Role string `json:"role"`
	Kind string `json:"kind"`
	Ask  string `json:"ask"`
}

// ParseLabel extracts the JSON object from a model answer and checks it
// against the taxonomy.
func ParseLabel(t Taxonomy, text string) (role, kind, ask string, err error) {
	start, end := strings.Index(text, "{"), strings.LastIndex(text, "}")
	if start < 0 || end < start {
		return "", "", "", fmt.Errorf("no JSON object in answer")
	}
	var r rawLabel
	if err := json.Unmarshal([]byte(text[start:end+1]), &r); err != nil {
		return "", "", "", fmt.Errorf("answer JSON: %w", err)
	}
	r.Role, r.Kind = strings.TrimSpace(r.Role), strings.TrimSpace(r.Kind)
	if !t.hasRole(r.Role) {
		return "", "", "", fmt.Errorf("role %q is not in the taxonomy", r.Role)
	}
	if !t.hasKind(r.Kind) {
		return "", "", "", fmt.Errorf("kind %q is not in the taxonomy", r.Kind)
	}
	return r.Role, r.Kind, capRunes(strings.TrimSpace(r.Ask), 200), nil
}

// Label statuses.
const (
	StatusOK      = "ok"
	StatusInvalid = "invalid" // the model answered, but not with a valid label
	StatusError   = "error"   // the call itself failed
)

// PromptLabel is the labeller's verdict on one prompt, with what the server
// says ran it (record what ran: per row, the served model, build and speed).
type PromptLabel struct {
	ID          int     `json:"id"`
	Role        string  `json:"role,omitempty"`
	Kind        string  `json:"kind,omitempty"`
	Ask         string  `json:"ask,omitempty"`
	Status      string  `json:"status"`
	Detail      string  `json:"detail,omitempty"`
	Attempts    int     `json:"attempts"`
	Model       string  `json:"model,omitempty"`
	Fingerprint string  `json:"fingerprint,omitempty"`
	TokPerSec   float64 `json:"tok_per_sec,omitempty"`
	Reasoning   bool    `json:"reasoning,omitempty"`
	Truncated   bool    `json:"truncated,omitempty"`
}

// LabelParams is the labeller's complete sampler chain: greedy decoding with
// every truncation and penalty stage pinned to its neutral value, so no stage
// inherits a server default (corpos-lab's record-what-ran rule). 160 tokens is
// well above a label's length; a truncated answer is recorded per row.
func LabelParams() model.GenParams {
	return model.GenParams{
		Temperature:      model.Float64(0),
		MaxTokens:        model.Int(160),
		Seed:             model.Int(0),
		TopK:             model.Int(0),
		TopP:             model.Float64(1),
		MinP:             model.Float64(0),
		TypicalP:         model.Float64(1),
		TopNSigma:        model.Float64(-1),
		RepeatPenalty:    model.Float64(1),
		RepeatLastN:      model.Int(0),
		PresencePenalty:  model.Float64(0),
		FrequencyPenalty: model.Float64(0),
		XTCProbability:   model.Float64(0),
		DryMultiplier:    model.Float64(0),
	}
}

// Classifier labels prompts with a model.
type Classifier struct {
	Client model.Client
	Tax    Taxonomy
	Params model.GenParams
}

// Classify labels p. A harness command is labelled without a model call: its
// role is the taxonomy's control role when one is named "control", and the
// model is never asked to judge what it cannot see. An answer that does not
// parse gets one retry; a second failure is recorded as invalid, not guessed.
func (c Classifier) Classify(ctx context.Context, p Prompt) PromptLabel {
	if p.Slash && c.Tax.hasRole(RoleControl) {
		return PromptLabel{ID: p.ID, Role: RoleControl, Kind: KindNone, Ask: firstWord(p.Text), Status: StatusOK}
	}
	prompt := BuildPrompt(c.Tax, p)
	out := PromptLabel{ID: p.ID}
	for attempt := 1; attempt <= 2; attempt++ {
		out.Attempts = attempt
		resp, err := c.Client.Generate(ctx, prompt, c.Params)
		if err != nil {
			out.Status, out.Detail = StatusError, err.Error()
			return out
		}
		out.Model, out.Fingerprint = resp.Model, resp.SystemFingerprint
		out.TokPerSec, out.Reasoning, out.Truncated = resp.Timings.PredictedPerSecond, resp.Reasoning != "", resp.Truncated
		role, kind, ask, err := ParseLabel(c.Tax, resp.Text)
		if err == nil {
			out.Role, out.Kind, out.Ask, out.Status, out.Detail = role, kind, ask, StatusOK, ""
			return out
		}
		out.Status, out.Detail = StatusInvalid, fmt.Sprintf("%v; answer: %s", err, capRunes(resp.Text, 200))
		prompt += retrySuffix
	}
	return out
}

// RoleControl is the role name the classifier assigns to harness commands
// without a model call, when the taxonomy defines it.
const RoleControl = "control"

func firstWord(s string) string {
	f := strings.Fields(s)
	if len(f) == 0 {
		return ""
	}
	return f[0]
}
