package control

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// mountSources returns the source side of every -v in argv, i.e. the "/host/x"
// of "/host/x:/in:ro,Z". This is the surface the bug lived on.
func mountSources(argv []string) []string {
	var out []string
	for i, a := range argv {
		if a == "-v" && i+1 < len(argv) {
			out = append(out, strings.SplitN(argv[i+1], ":", 2)[0])
		}
	}
	return out
}

func argvValue(argv []string, flag string) string {
	for i, a := range argv {
		if a == flag && i+1 < len(argv) {
			return argv[i+1]
		}
	}
	return ""
}

// THE REGRESSION TEST. podman reads a relative -v source as a NAMED VOLUME
// rather than a bind mount, so the documented invocation
//
//	corpos-lab run-study studies/x/study.toml
//
// derived a relative workDir and died with exit 125 before the assay started.
// A controller test cannot catch this: fakeLauncher never builds podman args,
// and t.TempDir() hands back absolute paths, so every existing test exercises
// only the absolute case. The assertion has to be over ARGV.
func TestPodmanArgsAbsolutizesRelativeMountSources(t *testing.T) {
	argv, err := PodmanArgs(LaunchSpec{
		Image:   "localhost/lab-grounded-glyph-probe@sha256:abc",
		InDir:   "studies/casg-direct-grounded-probe/runs/x/in",
		OutDir:  "studies/casg-direct-grounded-probe/runs/x/out",
		Network: "corpos-net",
	})
	if err != nil {
		t.Fatalf("PodmanArgs: %v", err)
	}

	srcs := mountSources(argv)
	if len(srcs) != 2 {
		t.Fatalf("expected 2 -v mounts, got %d: %v", len(srcs), argv)
	}
	for _, src := range srcs {
		if !filepath.IsAbs(src) {
			t.Errorf("mount source %q is relative — podman parses that as a named volume name, "+
				"not a host dir, and exits 125 on the '/' in it", src)
		}
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	wantIn := filepath.Join(cwd, "studies/casg-direct-grounded-probe/runs/x/in")
	if srcs[0] != wantIn {
		t.Errorf("in mount = %q, want %q", srcs[0], wantIn)
	}
}

// Absolutizing must not disturb the mount options. :ro on /in is what keeps a
// container from editing the stimulus it is being measured against.
func TestPodmanArgsPreservesMountOptions(t *testing.T) {
	dir := t.TempDir()
	argv, err := PodmanArgs(LaunchSpec{
		Image: "img", InDir: filepath.Join(dir, "in"), OutDir: filepath.Join(dir, "out"),
		Network: "corpos-net",
	})
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(argv, " ")
	if !strings.Contains(joined, filepath.Join(dir, "in")+":/in:ro,Z") {
		t.Errorf("/in must stay read-only with the SELinux label: %v", argv)
	}
	if !strings.Contains(joined, filepath.Join(dir, "out")+":/out:Z") {
		t.Errorf("/out mount malformed: %v", argv)
	}
}

// An already-absolute path is passed through unchanged — the fix must not
// rewrite what was already correct.
func TestPodmanArgsLeavesAbsolutePathsAlone(t *testing.T) {
	dir := t.TempDir()
	argv, err := PodmanArgs(LaunchSpec{
		Image: "img", InDir: dir + "/in", OutDir: dir + "/out", Network: "n",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := mountSources(argv); got[0] != dir+"/in" || got[1] != dir+"/out" {
		t.Errorf("absolute paths were rewritten: %v", got)
	}
}

// A relative -work value must WORK, not be rejected. The invocation is correct;
// the launcher was wrong.
func TestPodmanArgsAcceptsRelativeWorkDir(t *testing.T) {
	argv, err := PodmanArgs(LaunchSpec{
		Image: "img", InDir: "./runs/leg-1/in", OutDir: "runs/leg-1/out", Network: "n",
	})
	if err != nil {
		t.Fatalf("a relative -work must be supported, not refused: %v", err)
	}
	for _, src := range mountSources(argv) {
		if !filepath.IsAbs(src) {
			t.Errorf("mount source %q still relative", src)
		}
	}
}

// Dot-segments resolve rather than reaching podman as a literal "..".
func TestPodmanArgsCleansDotSegments(t *testing.T) {
	argv, err := PodmanArgs(LaunchSpec{
		Image: "img", InDir: "runs/../runs/x/in", OutDir: "./runs/./x/out", Network: "n",
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, src := range mountSources(argv) {
		if strings.Contains(src, "..") || strings.Contains(src, "/./") {
			t.Errorf("unresolved dot-segment in mount source %q", src)
		}
	}
}

// Per the bug's last criterion: refuse an unusable spec with a corpos-lab error
// so a future regression surfaces as our message rather than a podman
// volume-name charset complaint two layers away.
func TestPodmanArgsRefusesUnusableSpecs(t *testing.T) {
	dir := t.TempDir()
	for name, spec := range map[string]LaunchSpec{
		"no image":   {InDir: dir, OutDir: dir},
		"no in dir":  {Image: "img", OutDir: dir},
		"no out dir": {Image: "img", InDir: dir},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := PodmanArgs(spec)
			if err == nil {
				t.Fatalf("expected a refusal for %q", name)
			}
			if !strings.Contains(err.Error(), "control:") {
				t.Errorf("error should be ours, not podman's: %v", err)
			}
		})
	}
}

// The shape of the command podman is actually handed.
func TestPodmanArgsBuildsTheRunInvocation(t *testing.T) {
	dir := t.TempDir()
	argv, err := PodmanArgs(LaunchSpec{
		Image: "localhost/probe@sha256:abc", InDir: dir + "/in", OutDir: dir + "/out",
		Network: "corpos-net",
	})
	if err != nil {
		t.Fatal(err)
	}
	if argv[0] != "run" || argv[1] != "--rm" {
		t.Errorf("argv should start with a disposable run: %v", argv[:2])
	}
	if argvValue(argv, "--network") != "corpos-net" {
		t.Errorf("network = %q", argvValue(argv, "--network"))
	}
	// --userns=keep-id + --user let the container write the host-owned /out.
	if argvValue(argv, "--user") == "" || !strings.Contains(strings.Join(argv, " "), "--userns=keep-id") {
		t.Errorf("uid mapping missing: %v", argv)
	}
	// The image and the assay's own subcommand come last, in that order.
	if argv[len(argv)-2] != "localhost/probe@sha256:abc" || argv[len(argv)-1] != "run" {
		t.Errorf("argv should end with <image> run, got %v", argv[len(argv)-2:])
	}
}

// An empty network must not emit a bare "--network", which would swallow the
// image as its value and produce a baffling podman error.
func TestPodmanArgsOmitsEmptyNetwork(t *testing.T) {
	dir := t.TempDir()
	argv, err := PodmanArgs(LaunchSpec{Image: "img", InDir: dir, OutDir: dir})
	if err != nil {
		t.Fatal(err)
	}
	for _, a := range argv {
		if a == "--network" {
			t.Fatalf("bare --network emitted with no value: %v", argv)
		}
	}
	if argv[len(argv)-2] != "img" {
		t.Errorf("image should still be second-to-last: %v", argv)
	}
}
