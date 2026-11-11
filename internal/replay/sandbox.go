package replay

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Runner runs an external command. The sandbox reaches the world only through
// it (git, podman, bwrap, corpos, cp), so tests script every effect.
type Runner interface {
	Run(ctx context.Context, dir string, env []string, name string, args ...string) (stdout, stderr []byte, err error)
}

// ExecRunner runs commands for real. env, when non-nil, is appended to the
// current environment.
type ExecRunner struct{}

// Run executes name with args in dir.
func (ExecRunner) Run(ctx context.Context, dir string, env []string, name string, args ...string) ([]byte, []byte, error) {
	cmd := exec.CommandContext(ctx, name, args...) //nolint:gosec // the harness composes every command
	cmd.Dir = dir
	if env != nil {
		cmd.Env = append(os.Environ(), env...)
	}
	var out, errb bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errb
	err := cmd.Run()
	return out.Bytes(), errb.Bytes(), err
}

// Config is where the sandbox lives and what it is built from.
type Config struct {
	// Root holds one directory per case (private: it keeps transcripts).
	Root string
	// Home is the principal's home directory, used to place binds.
	Home string
	// BackupDir holds the nightly ledger backups (toolkit-<ts>.db).
	BackupDir string
	// VaultDir is copied so toolkit vault writes land in the sandbox.
	VaultDir string
	// Image, Network and Account configure the throwaway the toolkit service.
	Image   string
	Network string
	Account string
	// Port is the host port the sandbox toolkit publishes on 127.0.0.1.
	Port int
	// CorposSrc is the corpos checkout built into each sandbox, so a replay
	// runs current corpos rather than whatever binary is on PATH (which was
	// found ten weeks stale). GoBin builds it.
	CorposSrc string
	GoBin     string
	// RiskGate is corpos's -risk-gate posture. The default is build-test, the
	// swap window's posture (SWAP_VALIDATION_CRITERIA.md §4): it lets a coding
	// worker run builds and tests, which only reach the sandbox copies.
	RiskGate string
	// HealthWait bounds how long up waits for the sandbox toolkit.
	HealthWait time.Duration
}

// RepoState is a live repo's identity, taken before and after a replay.
type RepoState struct {
	Path   string `json:"path"`
	Head   string `json:"head"`
	Status string `json:"status_sha256"`
}

// State is a sandbox's record of itself, kept in <dir>/state.json.
type State struct {
	Case        Case              `json:"case"`
	Dir         string            `json:"dir"`
	Port        int               `json:"port"`
	Container   string            `json:"container"`
	Backup      string            `json:"backup"`
	StartedAt   time.Time         `json:"started_at"`
	RepoCopies  map[string]string `json:"repo_copies"`
	FileCopies  map[string]string `json:"file_copies"`
	LiveBefore  []RepoState       `json:"live_before"`
	CopyHeads   map[string]string `json:"copy_heads"`
	CorposBin   string            `json:"corpos_bin"`
	CorposVer   string            `json:"corpos_version"`
	ContainerID string            `json:"container_id"`
	// CorposCommit is the corpos commit the sandbox binary was built from;
	// a "+dirty" suffix marks uncommitted changes in the source tree.
	CorposCommit string `json:"corpos_commit"`
	Home         string `json:"home"`
	RiskGate     string `json:"risk_gate"`
	// DevRoot is the principal's dev root (~/dev). Inside the sandbox it is a
	// writable throwaway directory with each of DevEntries re-mounted
	// read-only, so tools that create sibling checkouts (git-flow worktrees at
	// ~/dev/<repo>-wt-<slug>) work. Empty when the dev root could not be
	// listed at up; then ~/dev stays read-only.
	DevRoot    string       `json:"dev_root,omitempty"`
	DevEntries []string     `json:"dev_entries,omitempty"`
	RunID      string       `json:"run_id,omitempty"`
	Turns      []TurnResult `json:"turns"`
}

// StateFile is the per-case state file name.
const StateFile = "state.json"

// disabledPush is the push URL every repo copy gets, so no push leaves it.
const disabledPush = "no-push://replay-sandbox"

// Up builds the sandbox for c. It refuses to reuse an existing case dir.
func Up(ctx context.Context, r Runner, cfg Config, c Case, now time.Time) (State, error) {
	if err := c.Validate(); err != nil {
		return State{}, err
	}
	dir := filepath.Join(cfg.Root, c.ID)
	if _, err := os.Stat(dir); err == nil {
		return State{}, fmt.Errorf("sandbox %s already exists; run down first", dir)
	}
	st := State{Case: c, Dir: dir, Port: cfg.Port, StartedAt: now.UTC(), Home: cfg.Home, CorposBin: filepath.Join(dir, "bin", "corpos"),
		RepoCopies: map[string]string{}, FileCopies: map[string]string{}, CopyHeads: map[string]string{},
		Container: "replay-tk-" + c.ID, RiskGate: cfg.RiskGate}
	if st.RiskGate == "" {
		st.RiskGate = "build-test"
	}
	for _, sub := range []string{"repos", "files", "toolkit-data", "corpos-state", "cache", "turns", "bin", "devroot"} {
		if err := os.MkdirAll(filepath.Join(dir, sub), 0o750); err != nil {
			return st, fmt.Errorf("make sandbox: %w", err)
		}
	}
	var err error
	if st.LiveBefore, err = SnapshotRepos(ctx, r, c.Repos); err != nil {
		return st, err
	}
	for i, src := range c.Repos {
		dst := filepath.Join(dir, "repos", fmt.Sprintf("%02d-%s", i, filepath.Base(src)))
		if err := copyRepo(ctx, r, src, dst); err != nil {
			return st, err
		}
		st.RepoCopies[src] = dst
		head, err := gitOut(ctx, r, dst, "rev-parse", "HEAD")
		if err != nil {
			return st, err
		}
		st.CopyHeads[src] = head
	}
	for i, src := range c.Files {
		dst := filepath.Join(dir, "files", fmt.Sprintf("%02d-%s", i, filepath.Base(src)))
		if _, _, err := r.Run(ctx, "", nil, "cp", "-a", "--", src, dst); err != nil {
			return st, fmt.Errorf("copy %s: %w", src, err)
		}
		st.FileCopies[src] = dst
	}
	devRoot := filepath.Join(cfg.Home, "dev")
	if entries, derr := os.ReadDir(devRoot); derr == nil {
		st.DevRoot = devRoot
		for _, e := range entries {
			st.DevEntries = append(st.DevEntries, e.Name())
		}
	}
	if st.Backup, err = NewestBackup(cfg.BackupDir); err != nil {
		return st, err
	}
	if _, _, err := r.Run(ctx, "", nil, "cp", "--", st.Backup, filepath.Join(dir, "toolkit-data", "toolkit.db")); err != nil {
		return st, fmt.Errorf("copy ledger backup: %w", err)
	}
	if _, _, err := r.Run(ctx, "", nil, "cp", "-a", "--", cfg.VaultDir, filepath.Join(dir, "vault")); err != nil {
		return st, fmt.Errorf("copy vault: %w", err)
	}
	if st.CorposCommit, err = sourceCommit(ctx, r, cfg.CorposSrc); err != nil {
		return st, err
	}
	if _, errOut, err := r.Run(ctx, cfg.CorposSrc, []string{"CGO_ENABLED=0"}, cfg.GoBin, "build", "-o", st.CorposBin, "./cmd/corpos"); err != nil {
		return st, fmt.Errorf("build corpos from %s: %w: %s", cfg.CorposSrc, err, strings.TrimSpace(string(errOut)))
	}
	if out, _, err := r.Run(ctx, "", nil, st.CorposBin, "-version"); err == nil {
		st.CorposVer = strings.TrimSpace(string(out))
	}
	out, errOut, err := r.Run(ctx, "", nil, "podman", toolkitRunArgs(cfg, st)...)
	if err != nil {
		return st, fmt.Errorf("start sandbox toolkit: %w: %s", err, strings.TrimSpace(string(errOut)))
	}
	st.ContainerID = strings.TrimSpace(string(out))
	if err := SaveState(st); err != nil {
		return st, err
	}
	return st, nil
}

// toolkitRunArgs mirrors the live the toolkit service unit, minus everything that
// reaches outside: no team-sync secret, no transcript or corpos-session
// mounts, the dev root read-only, the ledger and vault from the sandbox, and
// the port published on loopback only.
func toolkitRunArgs(cfg Config, st State) []string {
	return []string{
		"run", "-d", "--rm", "--name", st.Container,
		"--network", cfg.Network,
		"-p", fmt.Sprintf("127.0.0.1:%d:3000", st.Port),
		"--userns", "keep-id:uid=65532,gid=65532",
		"-v", filepath.Join(st.Dir, "toolkit-data") + ":/data",
		"-v", filepath.Join(st.Dir, "vault") + ":/home/nonroot/.claude/vault",
		"-v", filepath.Join(cfg.Home, "dev") + ":" + filepath.Join(cfg.Home, "dev") + ":ro",
		"-e", "TOOLKIT_DEV_ROOT=" + filepath.Join(cfg.Home, "dev"),
		"-e", "TOOLKIT_ACCOUNT=" + cfg.Account,
		"-e", "TOOLKIT_LOCAL_URL=http://llama-server:8081",
		"-e", "LLAMA_CPP_BASE_URL=http://llama-server:8081",
		"-e", "ML_ONNX_SERVE_URL=http://onnx-serve:9000",
		"-e", "ML_MODELS_ROOT=/models",
		"--add-host=kiwix-personal:172.18.0.2",
		"-e", "KIWIX_BASE_URL=http://kiwix-personal:8080",
		cfg.Image,
	}
}

// copyRepo clones src into dst and rewires the remotes: the clone's origin
// points at the LIVE checkout, so it is removed, and src's own remotes are
// re-added for fetch with their push URL disabled.
func copyRepo(ctx context.Context, r Runner, src, dst string) error {
	if _, errOut, err := r.Run(ctx, "", nil, "git", "clone", "--quiet", "--no-hardlinks", "--", src, dst); err != nil {
		return fmt.Errorf("clone %s: %w: %s", src, err, strings.TrimSpace(string(errOut)))
	}
	if _, err := gitOut(ctx, r, dst, "remote", "remove", "origin"); err != nil {
		return err
	}
	remotes, err := gitOut(ctx, r, src, "remote")
	if err != nil {
		return err
	}
	for _, name := range strings.Fields(remotes) {
		url, err := gitOut(ctx, r, src, "remote", "get-url", name)
		if err != nil {
			return err
		}
		if _, err := gitOut(ctx, r, dst, "remote", "add", name, url); err != nil {
			return err
		}
		if _, err := gitOut(ctx, r, dst, "remote", "set-url", "--push", name, disabledPush); err != nil {
			return err
		}
	}
	return nil
}

// sourceCommit names the commit a source tree is at, marking a dirty tree.
func sourceCommit(ctx context.Context, r Runner, dir string) (string, error) {
	head, err := gitOut(ctx, r, dir, "rev-parse", "HEAD")
	if err != nil {
		return "", err
	}
	status, err := gitOut(ctx, r, dir, "status", "--porcelain", "--untracked-files=no")
	if err != nil {
		return "", err
	}
	if status != "" {
		head += "+dirty"
	}
	return head, nil
}

func gitOut(ctx context.Context, r Runner, dir string, args ...string) (string, error) {
	out, errOut, err := r.Run(ctx, "", nil, "git", append([]string{"-C", dir}, args...)...)
	if err != nil {
		return "", fmt.Errorf("git -C %s %s: %w: %s", dir, strings.Join(args, " "), err, strings.TrimSpace(string(errOut)))
	}
	return strings.TrimSpace(string(out)), nil
}

// SnapshotRepos records each live repo's HEAD and a hash of its status.
func SnapshotRepos(ctx context.Context, r Runner, repos []string) ([]RepoState, error) {
	out := make([]RepoState, 0, len(repos))
	for _, p := range repos {
		head, err := gitOut(ctx, r, p, "rev-parse", "HEAD")
		if err != nil {
			return nil, err
		}
		status, err := gitOut(ctx, r, p, "status", "--porcelain")
		if err != nil {
			return nil, err
		}
		sum := sha256.Sum256([]byte(status))
		out = append(out, RepoState{Path: p, Head: head, Status: hex.EncodeToString(sum[:])})
	}
	return out, nil
}

// NewestBackup returns the newest toolkit-<ts>.db in dir. The timestamp is in
// the name, so the lexical maximum is the newest.
func NewestBackup(dir string) (string, error) {
	matches, err := filepath.Glob(filepath.Join(dir, "toolkit-2*.db"))
	if err != nil {
		return "", fmt.Errorf("list backups: %w", err)
	}
	if len(matches) == 0 {
		return "", fmt.Errorf("no ledger backup (toolkit-<ts>.db) in %s", dir)
	}
	sort.Strings(matches)
	return matches[len(matches)-1], nil
}

// SaveState writes state.json.
func SaveState(st State) error {
	raw, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal state: %w", err)
	}
	return os.WriteFile(filepath.Join(st.Dir, StateFile), append(raw, '\n'), 0o600)
}

// LoadState reads a case's state.json.
func LoadState(root, caseID string) (State, error) {
	raw, err := os.ReadFile(filepath.Join(root, caseID, StateFile)) //nolint:gosec // inside the replay root
	if err != nil {
		return State{}, fmt.Errorf("read sandbox state: %w", err)
	}
	var st State
	if err := json.Unmarshal(raw, &st); err != nil {
		return State{}, fmt.Errorf("sandbox state: %w", err)
	}
	return st, nil
}
