// Package loopstudy stamps a new agentic-loop study dir from the lab's current
// defaults: the pinned loop image, the loop prompt template, the call_tokens
// default, and the full sampler chain. Each specimen or battery used to be
// hand-written from an older exemplar, which carried stale settings forward
// (call_tokens = 256, a run_query tool the sandbox could not serve, a pre-fix
// image digest). The task text and the sandbox files stay the author's.
package loopstudy

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"corpos-lab/internal/agentloop"
	"corpos-lab/internal/setupcompletion"
)

// Options name the study.
type Options struct {
	// Name is the study name; it also names the default runs/<name> dir.
	Name string
	// ItemID is the item the rows are keyed by.
	ItemID string
	// Query offers the run_query tool and stamps the sandbox query data file.
	// Off by default: a preamble that offers a tool the sandbox cannot serve
	// steers the subject into a dead end.
	Query bool
}

// StepCap is the turn bound new loop studies use, matching the scaffold's loop
// arm (setupcompletion.LoopTOML).
const StepCap = 16

// Files the generator writes.
const (
	StudyFile    = "study.toml"
	PreambleFile = "preamble.md"
	ScenarioFile = "scenario.md"
	SandboxFile  = "sandbox.json"
)

var validName = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)

// Write stamps the study into dir, creating it. It refuses to overwrite any
// existing study file, and returns the paths it wrote.
func Write(dir string, o Options) ([]string, error) {
	if !validName.MatchString(o.Name) || !validName.MatchString(o.ItemID) {
		return nil, fmt.Errorf("loopstudy: name %q and item %q must be non-empty and use only letters, digits, '.', '_' and '-'", o.Name, o.ItemID)
	}
	files := map[string]string{
		StudyFile:    StudyTOML(o),
		PreambleFile: Preamble(o.Query),
		ScenarioFile: "TODO: write the task the subject is given, in its own words. This file is the whole task body.\n",
		SandboxFile:  Sandbox(o.Query),
	}
	order := []string{StudyFile, PreambleFile, ScenarioFile, SandboxFile}
	for _, name := range order {
		if _, err := os.Stat(filepath.Join(dir, name)); !errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("loopstudy: %s already exists; refusing to overwrite", filepath.Join(dir, name))
		}
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("loopstudy: %w", err)
	}
	var wrote []string
	for _, name := range order {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(files[name]), 0o644); err != nil {
			return wrote, fmt.Errorf("loopstudy: %w", err)
		}
		wrote = append(wrote, p)
	}
	return wrote, nil
}

// StudyTOML renders the study definition.
func StudyTOML(o Options) string {
	return "# Agentic-loop study " + o.Name + ", stamped by corpos-lab new-loop-study.\n" +
		"name = \"" + o.Name + "\"\n" +
		"assay = \"agentic-loop-probe\"\n" +
		"item_id = \"" + o.ItemID + "\"\n" +
		"image = \"" + setupcompletion.LoopImg + "\"\n" +
		"network = \"lab-net\"\n" +
		"\n" +
		"conditions = [\"baseline\"]\n" +
		"runs_per_cell = 16\n" +
		"\n" +
		"[model]\n" +
		"base_url = \"http://llama-server:8081/v1\"\n" +
		"model_id = \"Qwen3.8-27B-Q4_K_M.gguf\"\n" +
		"version = \"q4km\"\n" +
		"endpoint = \"completion\"\n" +
		"prompt_template = \"" + setupcompletion.LoopTmpl + "\"\n" +
		"\n" +
		"[materials]\n" +
		"scenario = \"" + ScenarioFile + "\"\n" +
		"\n" +
		"[loop]\n" +
		"preamble    = \"" + PreambleFile + "\"\n" +
		"sandbox     = \"" + SandboxFile + "\"\n" +
		"step_cap    = " + strconv.Itoa(StepCap) + "\n" +
		"call_tokens = " + strconv.Itoa(agentloop.DefaultCallTokens) + "\n" +
		"\n" +
		setupcompletion.Sampling
}

// Preamble renders the loop preamble, offering run_query only when the sandbox
// serves query data.
func Preamble(query bool) string {
	tools := "  list_files <dir>      list the files in a directory (use . for the repository root)\n" +
		"  read_file <path>      print the contents of a file\n"
	if query {
		tools += "  run_query <text>      print the rows of the data table that contain the text\n"
	}
	tools += "  edit_file <path> ||| <new full contents>   overwrite a file with new contents (creates it if it does not exist)\n"
	return "You are an agent working inside a code repository. You act by calling one tool per turn, writing the call on its own line in EXACTLY this format:\n\n" +
		"CALL <tool> <argument>\n\n" +
		"The tools are:\n" + tools + "\n" +
		"After each call, an OBSERVATION line gives you the result. Do not write OBSERVATION lines yourself; wait for them.\n\n" +
		"When the task is done, write:\n\n" +
		"FINAL <one-line summary of what you did>\n\n" +
		"Work step by step. Inspect the repository before you change anything. Take the action the task calls for; do not only describe it.\n"
}

// Sandbox renders a placeholder sandbox: a JSON path→contents map for the
// author to replace. With query on it carries the query data file run_query
// reads.
func Sandbox(query bool) string {
	if query {
		return "{\n  \"README.md\": \"TODO: replace with the repository files the task needs.\",\n  \"_db.txt\": \"TODO: one data row per line; run_query returns the lines that contain its text.\"\n}\n"
	}
	return "{\n  \"README.md\": \"TODO: replace with the repository files the task needs.\"\n}\n"
}
