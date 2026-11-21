package main

import (
	"flag"
	"fmt"
	"os"

	"corpos-lab/internal/image"
)

// runRepin rewrites the probe-image digest pins in the named files to the
// digests the last build recorded in deploy/IMAGE_DIGESTS.txt. It is the re-pin
// step of scripts/build-lab-images.sh --repin, so an image rebuild is no longer a
// hand edit across the scaffold constants and each study TOML. Only the named
// files change: a study that has already run keeps its literal pin.
func runRepin(args []string) int {
	fs := flag.NewFlagSet("repin", flag.ContinueOnError)
	digestsPath := fs.String("digests", "deploy/IMAGE_DIGESTS.txt", "image digests recorded by scripts/build-lab-images.sh")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	files := fs.Args()
	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "repin: name at least one file to re-pin")
		return 2
	}
	raw, err := os.ReadFile(*digestsPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "repin: read %s: %v (run scripts/build-lab-images.sh first to record the built digests)\n", *digestsPath, err)
		return 1
	}
	digests, err := image.ParseDigestFile(raw)
	if err != nil {
		fmt.Fprintf(os.Stderr, "repin: %s: %v\n", *digestsPath, err)
		return 1
	}
	total := 0
	for _, path := range files {
		n, err := repinFile(path, digests)
		if err != nil {
			fmt.Fprintf(os.Stderr, "repin: %v\n", err)
			return 1
		}
		fmt.Printf("repin: %s: %d pin(s) moved\n", path, n)
		total += n
	}
	fmt.Printf("repin: %d pin(s) moved across %d file(s)\n", total, len(files))
	return 0
}

func repinFile(path string, digests map[string]string) (int, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	src, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	out, n := image.Repin(string(src), digests)
	if n == 0 {
		return 0, nil
	}
	if err := os.WriteFile(path, []byte(out), info.Mode().Perm()); err != nil {
		return 0, fmt.Errorf("write %s: %w", path, err)
	}
	return n, nil
}
