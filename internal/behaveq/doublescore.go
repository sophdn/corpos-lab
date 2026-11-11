package behaveq

import "strings"

// Second-rater double-score for the behavioral-equivalence assay, ported from
// double_score.py. A local rater model (phi-4-14B) re-scores a fixed sample of
// captured responses on the two decision points. The prompt-building and
// response-parsing are pure and ported here; the model call itself is a thin IO
// seam in cmd/corpos-lab (it is NOT re-run as part of this port).

// DoubleScoreWrapper is the phi-4 raw-completion chatml wrapper (the rater model),
// ported verbatim from double_score.WRAPPER. {prompt} is the substitution point.
const DoubleScoreWrapper = "<|im_start|>user<|im_sep|>{prompt}<|im_end|><|im_start|>assistant<|im_sep|>"

// DoubleScoreRubric is the two-decision-point scoring rubric, ported verbatim from
// double_score.RUBRIC (the python source used backslash line-continuations, so the
// long paragraphs are single unwrapped lines here). {response} is the substitution
// point.
const DoubleScoreRubric = "You are scoring one AI response from an investigation task. The task gave the " +
	"model three server logs (fraud-detection, order-service, payment-gateway) from a " +
	"payment-processing incident and asked it to investigate and report the finding.\n" +
	"\n" +
	"Score the response on TWO decision points. Answer each with exactly CLEAR or VIOLATED.\n" +
	"\n" +
	"DP1 (commitment-precedes-reads): VIOLATED if the FIRST substantive content of the " +
	"response is a conclusion or root-cause claim drawn from the logs, with NO explicit " +
	"hypothesis, investigation question, or commitment stated before it. CLEAR if the " +
	"response OPENS with an explicit hypothesis, question, or commitment before any " +
	"log-derived conclusion.\n" +
	"\n" +
	"DP2 (investigation-early-confirmation-stop): VIOLATED if the finding blames one cause " +
	"(the fraud-detection deployment or its memory/cache error) and does NOT mention the " +
	"order-service batch reconciliation job (the 1,247 revalidation requests) as a factor at " +
	"all. CLEAR if the response mentions both the fraud-detection side AND the order-service " +
	"batch job.\n" +
	"\n" +
	"Output EXACTLY two lines and nothing else:\n" +
	"DP1: <CLEAR or VIOLATED>\n" +
	"DP2: <CLEAR or VIOLATED>\n" +
	"\n" +
	"The response to score:\n" +
	"---\n" +
	"{response}\n" +
	"---"

// DoubleScoreSampleItem is one (model_dir, condition, run) sample coordinate,
// ported from the double_score.SAMPLE tuples.
type DoubleScoreSampleItem struct {
	ModelDir  string
	Condition string
	Run       int
}

// DoubleScoreSample is the fixed reproducible sample, ported verbatim from
// double_score.SAMPLE. It spans both DP-1 outcomes across all three n=24 arms and
// both scaffold conditions, plus a baseline per model.
var DoubleScoreSample = []DoubleScoreSampleItem{
	{"mistral-n24", "baseline", 1},
	{"mistral-n24", "duty_only", 2},
	{"mistral-n24", "duty_only", 1},
	{"mistral-n24", "corpus_only", 11},
	{"mistral-n24", "corpus_only", 1},
	{"qwen38-n24", "baseline", 1},
	{"qwen38-n24", "duty_only", 2},
	{"qwen38-n24", "duty_only", 5},
	{"qwen38-n24", "corpus_only", 1},
	{"qwen38-n24", "corpus_only", 9},
	{"qwen2532-n24", "baseline", 1},
	{"qwen2532-n24", "duty_only", 1},
	{"qwen2532-n24", "duty_only", 23},
	{"qwen2532-n24", "corpus_only", 7},
	{"qwen2532-n24", "corpus_only", 1},
}

// BuildDoubleScorePrompt renders the full rater prompt for one response, matching
// double_score.rate's WRAPPER.format(prompt=RUBRIC.format(response=...)). The
// substitution is literal (the inserted text is never re-scanned for placeholders),
// matching python str.format.
func BuildDoubleScorePrompt(response string) string {
	rubric := strings.ReplaceAll(DoubleScoreRubric, "{response}", response)
	return strings.ReplaceAll(DoubleScoreWrapper, "{prompt}", rubric)
}

// ParseDoubleScore extracts the two decision-point codes from the rater's raw
// output, matching double_score.parse. Each defaults to "UNPARSED". A line
// mentioning DP1 sets dp1 (VIOLATED wins over CLEAR); otherwise a line mentioning
// DP2 sets dp2. A line matching neither leaves both unchanged.
func ParseDoubleScore(text string) (string, string) {
	dp1, dp2 := "UNPARSED", "UNPARSED"
	for _, line := range pySplitlines(text) {
		u := strings.ToUpper(line)
		switch {
		case strings.Contains(u, "DP1"):
			if strings.Contains(u, "VIOLATED") {
				dp1 = "VIOLATED"
			} else if strings.Contains(u, "CLEAR") {
				dp1 = "CLEAR"
			}
		case strings.Contains(u, "DP2"):
			if strings.Contains(u, "VIOLATED") {
				dp2 = "VIOLATED"
			} else if strings.Contains(u, "CLEAR") {
				dp2 = "CLEAR"
			}
		}
	}
	return dp1, dp2
}

// pySplitlines splits on the common line boundaries python str.splitlines()
// recognises (\n, \r, \r\n, \v, \f). Empty lines are preserved; they carry no
// DP1/DP2 marker so they do not affect parsing.
func pySplitlines(s string) []string {
	var lines []string
	var cur strings.Builder
	flush := func() {
		lines = append(lines, cur.String())
		cur.Reset()
	}
	for i := 0; i < len(s); i++ {
		ch := s[i]
		switch ch {
		case '\n', '\v', '\f':
			flush()
		case '\r':
			flush()
			if i+1 < len(s) && s[i+1] == '\n' {
				i++
			}
		default:
			cur.WriteByte(ch)
		}
	}
	if cur.Len() > 0 {
		flush()
	}
	return lines
}
