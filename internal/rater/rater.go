package rater

import (
	"context"
	"fmt"
	"strings"
	"time"

	"corpos-lab/internal/model"
)

// SliceLine is one item to rate: an id, the response text under judgment, and
// (for the action rater) the scenario that selects the concrete actions.
type SliceLine struct {
	ID       string `json:"id"`
	Text     string `json:"text"`
	Scenario string `json:"scenario,omitempty"`
}

// chatSeam is the inference a rater needs: one raw chat completion, unsplit.
// It is defined at the consumer and satisfied by *model.OpenAI.RawChat.
type chatSeam interface {
	RawChat(ctx context.Context, prompt string, params model.GenParams) (model.RawChatResult, error)
}

// judgment is the per-kind strategy: how to build a line's prompt and how to
// read a verdict out of a reply.
type judgment interface {
	prompt(line SliceLine) string
	// score reads the verdict from a reply's raw content and reasoning. The
	// grounded kind returns "" for an unparseable reply, so the caller can count
	// it and substitute a default; the action kind always returns a canonical
	// verdict.
	score(content, reasoning string) string
	// unparsedCode is what an empty score() becomes.
	unparsedCode() string
}

// groundedJudgment scores the grounded-glyph probe codes against a class rubric.
type groundedJudgment struct{ rubric string }

func (g groundedJudgment) prompt(l SliceLine) string      { return GroundedPrompt(g.rubric, l.Text) }
func (g groundedJudgment) score(content, r string) string { return GroundedCode(content, r) }
func (g groundedJudgment) unparsedCode() string           { return "N" }

// actionJudgment scores which action a model committed to. It reads content, or
// falls back to reasoning when content is blank (a reasoning model that carried
// the answer in reasoning_content).
type actionJudgment struct{}

func (actionJudgment) prompt(l SliceLine) string { return ActionPrompt(l.Scenario, l.Text) }

func (actionJudgment) score(content, reasoning string) string {
	reply := content
	if strings.TrimSpace(reply) == "" {
		reply = reasoning
	}
	return actionVerdict(reply)
}

func (actionJudgment) unparsedCode() string { return "unscoreable" }

// Provenance is the record a rater with no content digest leaves beside its
// scores: what answered, over which endpoint, when, and at what token cost. It
// is captured from the first reply and totalled over the slice.
type Provenance struct {
	ModelReported     string      `json:"model_reported"`
	Endpoint          string      `json:"endpoint"`
	RequestedModel    string      `json:"requested_model"`
	ResponseID        string      `json:"response_id"`
	SystemFingerprint string      `json:"system_fingerprint"`
	RunDate           string      `json:"run_date"`
	Calls             int         `json:"calls"`
	Usage             model.Usage `json:"usage"`
}

// SliceScores is one slice's {id: code} result plus the provenance gathered
// while scoring and the count of replies that could not be parsed.
type SliceScores struct {
	Scores     map[string]string
	Provenance *Provenance
	Unparsed   int
}

// Config configures a rater. Seam and RaterID are required. Endpoint and
// RequestedModel are recorded in provenance. MaxTokens overrides the per-rating
// budget; nil takes the hosted-or-local default. Hosted marks a hosted endpoint
// (a generous default budget, and no seed on the wire).
type Config struct {
	Seam           chatSeam
	RaterID        string
	Endpoint       string
	RequestedModel string
	MaxTokens      *int
	Hosted         bool
}

// Rater scores slices of responses with one model and one rubric kind.
type Rater struct {
	seam           chatSeam
	judge          judgment
	params         model.GenParams
	raterID        string
	endpoint       string
	requestedModel string
	now            func() time.Time
}

// NewGrounded builds a rater for the grounded-glyph codes (C/Ii/Ic/I/N) against
// the given class rubric.
func NewGrounded(cfg Config, rubric string) *Rater {
	return newRater(cfg, groundedJudgment{rubric: rubric})
}

// NewAction builds a rater for the committed-action verdict
// (A_local/A_canon/neither/unscoreable).
func NewAction(cfg Config) *Rater {
	return newRater(cfg, actionJudgment{})
}

func newRater(cfg Config, judge judgment) *Rater {
	params := model.GenParams{
		Temperature: model.Float64(0),
		MaxTokens:   model.Int(EffectiveMaxTokens(cfg.MaxTokens, cfg.Hosted)),
	}
	// A local llama-server honours a seed; hosted APIs may reject it. At
	// temperature 0 decoding is greedy and the seed is inert, but the reference
	// rater pins it for the local path, so keep that.
	if !cfg.Hosted {
		params.Seed = model.Int(1)
	}
	return &Rater{
		seam:           cfg.Seam,
		judge:          judge,
		params:         params,
		raterID:        cfg.RaterID,
		endpoint:       cfg.Endpoint,
		requestedModel: cfg.RequestedModel,
		now:            time.Now,
	}
}

// RateSlice scores every line in a slice and returns the {id: code} map with the
// provenance gathered along the way. It stops on the first inference error, so a
// partial slice never promotes to a canonical result.
func (r *Rater) RateSlice(ctx context.Context, lines []SliceLine) (SliceScores, error) {
	scores := make(map[string]string, len(lines))
	var prov *Provenance
	var usage model.Usage
	calls, unparsed := 0, 0
	for _, line := range lines {
		res, err := r.seam.RawChat(ctx, r.judge.prompt(line), r.params)
		if err != nil {
			return SliceScores{}, fmt.Errorf("rater %s: id %s: %w", r.raterID, line.ID, err)
		}
		calls++
		usage = addUsage(usage, res.Usage)
		if prov == nil {
			prov = &Provenance{
				ModelReported:     res.Model,
				Endpoint:          r.endpoint,
				RequestedModel:    r.requestedModel,
				ResponseID:        res.ResponseID,
				SystemFingerprint: res.SystemFingerprint,
				RunDate:           r.now().Format("2006-01-02T15:04:05-0700"),
			}
		}
		code := r.judge.score(res.Content, res.ReasoningContent)
		if code == "" {
			unparsed++
			code = r.judge.unparsedCode()
		}
		scores[line.ID] = code
	}
	if prov != nil {
		prov.Calls = calls
		prov.Usage = usage
	}
	return SliceScores{Scores: scores, Provenance: prov, Unparsed: unparsed}, nil
}

func addUsage(a, b model.Usage) model.Usage {
	return model.Usage{
		PromptTokens:          a.PromptTokens + b.PromptTokens,
		CompletionTokens:      a.CompletionTokens + b.CompletionTokens,
		TotalTokens:           a.TotalTokens + b.TotalTokens,
		PromptCacheHitTokens:  a.PromptCacheHitTokens + b.PromptCacheHitTokens,
		PromptCacheMissTokens: a.PromptCacheMissTokens + b.PromptCacheMissTokens,
	}
}
