// Package calibrationstudies generates the cell definition TOMLs and renders the
// consensus scoring table for the two calibration studies: calibrated mechanism
// re-typing (task 4227, prefix "cmr") and scramble-strength calibration (task
// 4258, prefix "scb"). It is the Go replacement for the four study shell scripts
// gen-defs.sh and scoring/consensus_table.sh, which the whole-tree language
// policy forbids (shell is sanctioned only for CI glue, not under studies/).
//
// All templating here is pure: the caller supplies the IO (write the TOMLs,
// print the table). The consensus vote reuses lengthdistraction.Majority so the
// two studies share the one measure-of-record consensus rule.
package calibrationstudies

import "fmt"

// image is the pinned probe image both studies run against. It is referenced by
// content digest, never a mutable tag, as the run record requires.
const image = "localhost/lab-grounded-glyph-probe@sha256:43ac8df6d5ac63eff3c0bddecb881949074374ef9d57a4352c716aeedf9b22d3"

// samplingBlock is the sampler section shared by every cell. Its leading newline
// is the blank line before the [sampling] header; it ends with a trailing
// newline after the last field.
const samplingBlock = `
[sampling]
temperature = 0.8
seeds = [1, 2, 3, 4, 5, 6, 7, 8]
max_tokens = 512

top_n_sigma = -1.0
top_k = 0
typical_p = 1.0
top_p = 1.0
min_p = 0.05

repeat_penalty = 1.0
repeat_last_n = 0
presence_penalty = 0.0
frequency_penalty = 0.0

xtc_probability = 0.0
dry_multiplier = 0.0
`

// modelSpec names one model arm: the short label used in file names and
// templates, the GGUF the server loads, and the version tag recorded per run.
type modelSpec struct {
	short   string
	modelID string
	version string
}

// models is the three-model arm shared by both studies, in the order the shell
// scripts emit them.
var models = []modelSpec{
	{"mistral", "Mistral-7B-Instruct-v0.3.Q4_K_M.gguf", "v0.3-q4km"},
	{"phi4", "phi-4-Q4_K_M.gguf", "q4km"},
	{"qwen38", "Qwen3.8-27B-Q4_K_M.gguf", "q4km"},
}

// promptTemplate is the raw-completion chat template for one model short label.
// The literal \n sequences are two characters each, written into the TOML as-is;
// the assay expands {prompt} at run time.
func promptTemplate(short string) string {
	switch short {
	case "mistral":
		return `[INST] {prompt}\n\n---\nNote: You are running as a single completion. There is no harness, no tools, no filesystem, no shell. You cannot call functions or read files. Give your answer directly as text. [/INST]`
	case "phi4":
		return `<|im_start|>user<|im_sep|>{prompt}\n\n---\nNote: You are running as a single completion. There is no harness, no tools, no filesystem, no shell. You cannot call functions or read files. Give your answer directly as text.<|im_end|><|im_start|>assistant<|im_sep|>`
	case "qwen38":
		return `<|im_start|>user\n{prompt}\n\n---\nNote: You are running as a single completion. There is no harness, no tools, no filesystem, no shell. You cannot call functions or read files. Give your answer directly as text.<|im_end|>\n<|im_start|>assistant\n<think>\n\n</think>\n\n`
	default:
		return ""
	}
}

// modelBlock renders the [model] section for one arm. Its leading newline is the
// blank line before the [model] header; it ends with a trailing newline after
// prompt_template.
func modelBlock(m modelSpec) string {
	return fmt.Sprintf(`
[model]
base_url = "http://llama-server:8081/v1"
model_id = "%s"
version = "%s"
endpoint = "completion"
prompt_template = "%s"
`, m.modelID, m.version, promptTemplate(m.short))
}

// Cell is one generated cell definition: the file name it writes to and its full
// TOML content.
type Cell struct {
	Name string
	TOML string
}

// cmrClasses is the six decision classes re-typed under the calibrated dial, in
// emit order.
var cmrClasses = []string{
	"casg-direct",
	"formal-step-context-bypass",
	"structural-ceiling-bypass",
	"post-write-verification-absent",
	"governed-operation-protocol-bypass",
	"parent-state-check-bypass",
}

// CMRCells returns the 18 calibrated-mechanism-retyping cells (six classes times
// three models), in the order gen-defs.sh emits them.
func CMRCells() []Cell {
	out := make([]Cell, 0, len(cmrClasses)*len(models))
	for _, class := range cmrClasses {
		for _, m := range models {
			out = append(out, Cell{
				Name: fmt.Sprintf("cmr-%s-%s.toml", class, m.short),
				TOML: cmrCell(class, m),
			})
		}
	}
	return out
}

func cmrCell(class string, m modelSpec) string {
	return fmt.Sprintf(`# calibrated mechanism re-typing (task 4227) — %s — %s
name = "cmr-%s-s1-%s"
assay = "grounded-glyph-probe"
item_id = "%s"

image = "%s"
network = "lab-net"

conditions = ["baseline", "glyph_only", "scrambled_glyph", "off_target_glyph"]
runs_per_cell = 8
%s
[materials]
scenario   = "../%s/materials/scenario.md"
glyph      = "../%s/materials/glyph.md"
scrambled  = "../%s/materials/scrambled_75.md"
off_target = "../%s/materials/off_target_glyph.md"
%s`,
		class, m.short,
		class, m.short,
		class,
		image,
		modelBlock(m),
		class, class, class, class,
		samplingBlock)
}

// scbArm pairs the short arm directory name (used in material paths) with the
// full class name (used in cell names and item ids).
type scbArm struct {
	arm   string
	class string
}

var scbArms = []scbArm{
	{"casg", "casg-direct"},
	{"formal", "formal-step-context-bypass"},
	{"parent", "parent-state-check-bypass"},
}

// scbStrength pairs the scramble strength number (the scrambled_<n>.md suffix)
// with its zero-padded form (the item_id and file-name suffix).
type scbStrength struct {
	strength string
	pad      string
}

var scbStrengths = []scbStrength{
	{"0", "000"},
	{"25", "025"},
	{"50", "050"},
	{"75", "075"},
	{"100", "100"},
}

// SCBCells returns the 54 scramble-strength-calibration cells: for each arm and
// model, one ref cell (baseline + glyph_only) then one cell per strength, in the
// order gen-defs.sh emits them.
func SCBCells() []Cell {
	out := make([]Cell, 0, len(scbArms)*len(models)*(1+len(scbStrengths)))
	for _, a := range scbArms {
		for _, m := range models {
			out = append(out, Cell{
				Name: fmt.Sprintf("scb-%s-ref-%s.toml", a.class, m.short),
				TOML: scbRefCell(a, m),
			})
			for _, s := range scbStrengths {
				out = append(out, Cell{
					Name: fmt.Sprintf("scb-%s-str%s-%s.toml", a.class, s.pad, m.short),
					TOML: scbStrengthCell(a, s, m),
				})
			}
		}
	}
	return out
}

func scbRefCell(a scbArm, m modelSpec) string {
	return fmt.Sprintf(`# scramble calibration (task 4258) — %s ref (baseline + glyph_only) — %s
name = "scb-%s-ref-s1-%s"
assay = "grounded-glyph-probe"
item_id = "%s"

image = "%s"
network = "lab-net"

conditions = ["baseline", "glyph_only"]
runs_per_cell = 8
%s
[materials]
scenario = "../%s/materials/scenario.md"
glyph    = "../%s/materials/glyph.md"
%s`,
		a.class, m.short,
		a.class, m.short,
		a.class,
		image,
		modelBlock(m),
		a.arm, a.arm,
		samplingBlock)
}

func scbStrengthCell(a scbArm, s scbStrength, m modelSpec) string {
	return fmt.Sprintf(`# scramble calibration (task 4258) — %s scrambled_glyph strength %s — %s
name = "scb-%s-str%s-s1-%s"
assay = "grounded-glyph-probe"
item_id = "%s-str%s"

image = "%s"
network = "lab-net"

conditions = ["scrambled_glyph"]
runs_per_cell = 8
%s
[materials]
scenario  = "../%s/materials/scenario.md"
scrambled = "../%s/materials/scrambled_%s.md"
%s`,
		a.class, s.strength, m.short,
		a.class, s.pad, m.short,
		a.class, s.pad,
		image,
		modelBlock(m),
		a.arm, a.arm, s.strength,
		samplingBlock)
}
