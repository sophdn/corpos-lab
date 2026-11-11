package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"corpos-lab/internal/replay"
	"corpos-lab/internal/taskcatalogue"
)

// runReplay drives one catalogue case through corpos in a sandbox (chain
// popular-task-catalogue, task build-replay-harness). The verbs:
//
//	up    <case.json>             build the sandbox and start its toolkit
//	turn  <case-id> -prompt TEXT  send one principal turn (first starts, later resume)
//	check <case-id>               run the leak checks and print the ledger row
//	down  <case-id> -ledger FILE  check, append the row, tear the sandbox down
func runReplay(args []string) int {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "replay: want up <case.json> | turn <case-id> -prompt TEXT | check <case-id> | down <case-id> -ledger FILE")
		return 2
	}
	verb, target := args[0], args[1]
	home, _ := os.UserHomeDir()
	fs := flag.NewFlagSet("replay "+verb, flag.ContinueOnError)
	root := fs.String("root", filepath.Join(home, ".local", "state", "corpos-lab", "replay"), "sandbox root (private)")
	port := fs.Int("port", 3091, "host port for the sandbox toolkit")
	prompt := fs.String("prompt", "", "principal turn text (turn)")
	ledger := fs.String("ledger", "", "replay ledger JSONL to append to (down)")
	liveURL := fs.String("live-url", "http://localhost", "live the toolkit service, read for the leak check")
	riskGate := fs.String("risk-gate", "build-test", "corpos -risk-gate posture for the case (up)")
	corposSrc := fs.String("corpos-src", filepath.Join(home, "dev", "corpos"), "corpos checkout to build into the sandbox (up)")
	if err := fs.Parse(args[2:]); err != nil {
		return 2
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	r := replay.ExecRunner{}
	httpc := &http.Client{Timeout: 30 * time.Second}

	fail := func(err error) int {
		fmt.Fprintf(os.Stderr, "replay %s: %v\n", verb, err)
		return 1
	}
	switch verb {
	case "up":
		c, err := replay.LoadCase(target)
		if err != nil {
			return fail(err)
		}
		cfg := replay.Config{
			Root: *root, Home: home, Port: *port,
			BackupDir: filepath.Join(home, ".local", "share", "toolkit", "backups"),
			VaultDir:  filepath.Join(home, ".claude", "vault"),
			Image:     "localhost/the toolkit service:dev", Network: "lab-net", Account: "sophi",
			CorposSrc: *corposSrc, GoBin: "/usr/local/go/bin/go", RiskGate: *riskGate,
		}
		st, err := replay.Up(ctx, r, cfg, c, time.Now())
		if err != nil {
			return fail(err)
		}
		sbx := taskcatalogue.Toolkit{BaseURL: fmt.Sprintf("http://localhost:%d", st.Port), Project: c.Project, HTTP: httpc}
		if err := replay.WaitReady(ctx, sbx, 90*time.Second, 2*time.Second); err != nil {
			return fail(err)
		}
		fmt.Printf("replay up: %s ready (toolkit :%d, ledger from %s, %d repo copy(ies))\n", st.Dir, st.Port, filepath.Base(st.Backup), len(st.RepoCopies))
		return 0
	case "turn":
		if *prompt == "" {
			return fail(fmt.Errorf("-prompt is required"))
		}
		st, err := replay.LoadState(*root, target)
		if err != nil {
			return fail(err)
		}
		res, err := replay.RunTurn(ctx, r, &st, *prompt, time.Now)
		if err != nil {
			return fail(err)
		}
		fmt.Printf("replay turn %d: exit %d, %.0fs, $%.4f, run %s\n  out: %s\n  err: %s\n", res.N, res.Exit, res.Seconds, res.CostUSD, res.RunID, res.OutFile, res.ErrFile)
		return 0
	case "check", "down":
		st, err := replay.LoadState(*root, target)
		if err != nil {
			return fail(err)
		}
		sbx := taskcatalogue.Toolkit{BaseURL: fmt.Sprintf("http://localhost:%d", st.Port), Project: st.Case.Project, HTTP: httpc}
		live := taskcatalogue.Toolkit{BaseURL: *liveURL, Project: st.Case.Project, HTTP: httpc}
		row, err := replay.Check(ctx, r, st, sbx, live)
		if err != nil {
			return fail(err)
		}
		raw, _ := json.MarshalIndent(row, "", "  ")
		fmt.Println(string(raw))
		if len(row.LedgerLeaks) > 0 {
			fmt.Fprintf(os.Stderr, "replay %s: LEAK — %d ledger entity(ies) changed in the live ledger during the replay\n", verb, len(row.LedgerLeaks))
		}
		if verb == "check" {
			return 0
		}
		if *ledger == "" {
			return fail(fmt.Errorf("-ledger is required"))
		}
		if err := replay.AppendRow(*ledger, row); err != nil {
			return fail(err)
		}
		if err := replay.Teardown(ctx, r, st); err != nil {
			return fail(err)
		}
		fmt.Printf("replay down: %s torn down; row appended to %s\n", st.Dir, *ledger)
		return 0
	default:
		return fail(fmt.Errorf("unknown verb"))
	}
}
