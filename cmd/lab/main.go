// Command lab is the host-side freeze-by-digest tool for containerized
// assays. It pins a study run to exact content digests of its image and
// materials, and verifies a run against that pin before launch — a mismatch
// is a hard refusal (non-zero exit), never a warning.
//
// Usage:
//
//	lab pin    -in DIR -image REF [-out manifest.json]
//	lab verify -in DIR -image REF -manifest manifest.json
//
// The run-orchestration (podman run, volume mounts, /out collection) lives in
// the lab controller (chain task corpos-as-lab-controller); this tool owns the
// pin/verify half of the contract.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"

	"corpos-lab/internal/image"
	"corpos-lab/internal/manifest"
)

// podmanRunner inspects ref with rootless podman and returns the raw digest
// line. Lives here (not in internal/image) so that package stays sans-IO.
func podmanRunner(ctx context.Context, ref string) ([]byte, error) {
	out, err := exec.CommandContext(ctx, "podman", "image", "inspect", "--format", "{{.Digest}}", ref).Output()
	if err != nil {
		return nil, fmt.Errorf("podman inspect %s: %w", ref, err)
	}
	return out, nil
}

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: lab <pin|verify> ...")
		return 2
	}
	switch args[0] {
	case "pin":
		return runPin(args[1:])
	case "verify":
		return runVerify(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "lab: unknown subcommand %q (want pin|verify)\n", args[0])
		return 2
	}
}

func runPin(args []string) int {
	fs := flag.NewFlagSet("pin", flag.ContinueOnError)
	inDir := fs.String("in", "", "input dir holding study.json + materials")
	imageRef := fs.String("image", "", "assay image reference (e.g. localhost/lab-grounded-glyph-probe:dev)")
	outFile := fs.String("out", "", "manifest output path (default: <in>/manifest.json)")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *inDir == "" || *imageRef == "" {
		fmt.Fprintln(os.Stderr, "lab pin: -in and -image are required")
		return 2
	}

	digest, err := image.Digest(context.Background(), podmanRunner, *imageRef)
	if err != nil {
		fmt.Fprintf(os.Stderr, "lab pin: %v\n", err)
		return 1
	}
	pinned, err := manifest.Compute(*inDir, manifest.ImagePin{Ref: *imageRef, Digest: digest})
	if err != nil {
		fmt.Fprintf(os.Stderr, "lab pin: %v\n", err)
		return 1
	}

	path := *outFile
	if path == "" {
		path = *inDir + "/manifest.json"
	}
	raw, err := json.MarshalIndent(pinned, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "lab pin: marshal manifest: %v\n", err)
		return 1
	}
	if err := os.WriteFile(path, append(raw, '\n'), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "lab pin: write %s: %v\n", path, err)
		return 1
	}
	fmt.Printf("lab: pinned %s @ %s\n", *imageRef, digest[:19])
	fmt.Printf("lab: manifest -> %s\n", path)
	return 0
}

func runVerify(args []string) int {
	fs := flag.NewFlagSet("verify", flag.ContinueOnError)
	inDir := fs.String("in", "", "input dir holding study.json + materials")
	imageRef := fs.String("image", "", "assay image reference to verify against the pin")
	manifestFile := fs.String("manifest", "", "pinned manifest to verify against")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *inDir == "" || *imageRef == "" || *manifestFile == "" {
		fmt.Fprintln(os.Stderr, "lab verify: -in, -image, and -manifest are required")
		return 2
	}

	raw, err := os.ReadFile(*manifestFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "lab verify: read manifest: %v\n", err)
		return 1
	}
	var pinned manifest.RunManifest
	if err := json.Unmarshal(raw, &pinned); err != nil {
		fmt.Fprintf(os.Stderr, "lab verify: parse manifest: %v\n", err)
		return 1
	}

	digest, err := image.Digest(context.Background(), podmanRunner, *imageRef)
	if err != nil {
		fmt.Fprintf(os.Stderr, "lab verify: %v\n", err)
		return 1
	}

	if err := manifest.Verify(*inDir, manifest.ImagePin{Ref: *imageRef, Digest: digest}, pinned); err != nil {
		// Hard refusal — the run must not proceed.
		fmt.Fprintf(os.Stderr, "lab verify: REFUSED\n%v\n", err)
		return 1
	}
	fmt.Println("lab verify: OK — run is faithful to the pinned manifest")
	return 0
}
