package replay

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// TurnResult is what one principal turn produced.
type TurnResult struct {
	N        int       `json:"n"`
	Prompt   string    `json:"prompt"`
	Started  time.Time `json:"started"`
	Seconds  float64   `json:"seconds"`
	Exit     int       `json:"exit"`
	RunID    string    `json:"run_id,omitempty"`
	CostUSD  float64   `json:"cost_usd"`
	InTok    int       `json:"in_tok"`
	OutTok   int       `json:"out_tok"`
	Tools    []string  `json:"tools,omitempty"`
	Answer   string    `json:"answer_excerpt"`
	OutFile  string    `json:"out_file"`
	ErrFile  string    `json:"err_file"`
	ErrorMsg string    `json:"error,omitempty"`
}

// BwrapArgs builds the bubblewrap command line for one corpos turn: the host
// root read-only, each copy bound at its real path, corpos's state and cache
// dirs on sandbox directories, a private /tmp, and corpos pointed at the
// sandbox toolkit under the state's risk-gate posture.
func BwrapArgs(st State, prompt string) []string {
	args := []string{
		"--ro-bind", "/", "/",
		"--dev", "/dev", "--proc", "/proc", "--tmpfs", "/tmp",
		// Hide the per-user runtime dir: it holds the systemd, D-Bus and podman
		// sockets, through which a process could restart live services even
		// with the file system read-only.
		"--tmpfs", "/run/user",
		"--bind", filepath.Join(st.Dir, "corpos-state"), filepath.Join(st.Home, ".local", "state", "corpos"),
		"--bind", filepath.Join(st.Dir, "cache"), filepath.Join(st.Home, ".cache"),
	}
	if st.DevRoot != "" {
		args = append(args, "--bind", filepath.Join(st.Dir, "devroot"), st.DevRoot)
		for _, name := range st.DevEntries {
			p := filepath.Join(st.DevRoot, name)
			if _, copied := st.RepoCopies[p]; copied {
				continue
			}
			args = append(args, "--ro-bind-try", p, p)
		}
	}
	for _, real := range sortedKeys(st.RepoCopies) {
		args = append(args, "--bind", st.RepoCopies[real], real)
	}
	for _, real := range sortedKeys(st.FileCopies) {
		args = append(args, "--bind", st.FileCopies[real], real)
	}
	args = append(args, "--chdir", st.Case.Cwd, "--", st.CorposBin,
		"-mcp-url", fmt.Sprintf("http://localhost:%d", st.Port),
		"-project", st.Case.Project,
		"-risk-gate", st.RiskGate,
		"-session-dir", filepath.Join(st.Home, ".local", "state", "corpos", "sessions"),
	)
	if st.Case.Profile != "" {
		args = append(args, "-profile", st.Case.Profile)
	}
	if st.RunID != "" {
		args = append(args, "-resume", st.RunID)
	}
	return append(args, "-prompt", prompt)
}

var (
	sessionLine = regexp.MustCompile(`corpos: (?:session (\S+) \(resume with|resumed session (\S+) \()`)
	costLine    = regexp.MustCompile(`session cost \$([0-9.]+)`)
	tokLine     = regexp.MustCompile(`([0-9]+) in (?:\([0-9]+ cached\) )?\+ ([0-9]+) out tok`)
	toolLine    = regexp.MustCompile(`\[tool ([a-z_]+\.[a-z_]+) -> ([a-z_]+)\]`)
)

// ParseTurn reads corpos's stderr summary lines into r.
func ParseTurn(stderr string, r *TurnResult) {
	if m := sessionLine.FindStringSubmatch(stderr); m != nil {
		r.RunID = m[1] + m[2]
	}
	if m := costLine.FindStringSubmatch(stderr); m != nil {
		r.CostUSD, _ = strconv.ParseFloat(m[1], 64)
	}
	// One token line per model on the ladder; the turn's total is their sum.
	for _, m := range tokLine.FindAllStringSubmatch(stderr, -1) {
		in, _ := strconv.Atoi(m[1])
		out, _ := strconv.Atoi(m[2])
		r.InTok += in
		r.OutTok += out
	}
	for _, m := range toolLine.FindAllStringSubmatch(stderr, -1) {
		r.Tools = append(r.Tools, m[1]+":"+m[2])
	}
}

// RunTurn sends prompt to corpos inside the sandbox, saves its stdout and
// stderr under turns/, and appends the result to st (saved to state.json).
// A non-zero corpos exit is a result, not an error: the turn happened.
func RunTurn(ctx context.Context, r Runner, st *State, prompt string, now func() time.Time) (TurnResult, error) {
	n := len(st.Turns) + 1
	res := TurnResult{N: n, Prompt: prompt, Started: now().UTC()}
	out, errOut, err := r.Run(ctx, "", nil, "bwrap", BwrapArgs(*st, prompt)...)
	res.Seconds = now().Sub(res.Started).Seconds()
	var exitErr *exec.ExitError
	switch {
	case err == nil:
	case errors.As(err, &exitErr):
		res.Exit = exitErr.ExitCode()
	default:
		res.Exit = -1
		res.ErrorMsg = err.Error()
	}
	res.OutFile = filepath.Join(st.Dir, "turns", fmt.Sprintf("%02d.out", n))
	res.ErrFile = filepath.Join(st.Dir, "turns", fmt.Sprintf("%02d.err", n))
	if werr := os.WriteFile(res.OutFile, out, 0o600); werr != nil {
		return res, fmt.Errorf("save turn output: %w", werr)
	}
	if werr := os.WriteFile(res.ErrFile, errOut, 0o600); werr != nil {
		return res, fmt.Errorf("save turn stderr: %w", werr)
	}
	ParseTurn(string(errOut), &res)
	res.Answer = excerpt(string(out), 600)
	if st.RunID == "" {
		st.RunID = res.RunID
	}
	st.Turns = append(st.Turns, res)
	return res, SaveState(*st)
}

func excerpt(s string, n int) string {
	s = strings.TrimSpace(s)
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}

func sortedKeys(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	// Bind parents before children so a nested copy is not hidden.
	sortStrings(out)
	return out
}
