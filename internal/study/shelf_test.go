package study

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fakeResolver is a refResolver whose answers the test controls, and which
// counts its calls so a test can assert a literal value never triggered a
// lookup.
type fakeResolver struct {
	modelID, version   string
	imageRef           string
	modelErr, imageErr error
	modelCalls         int
	imageCalls         int
}

func (f *fakeResolver) model(string) (string, string, error) {
	f.modelCalls++
	return f.modelID, f.version, f.modelErr
}

func (f *fakeResolver) image(string) (string, error) {
	f.imageCalls++
	return f.imageRef, f.imageErr
}

func TestHasRefs(t *testing.T) {
	cases := []struct {
		modelID, image string
		want           bool
	}{
		{"Qwen3.8-27B-Q4_K_M.gguf", "localhost/x@sha256:a", false},
		{"role:primary", "localhost/x@sha256:a", true},
		{"Qwen3.8-27B-Q4_K_M.gguf", "digest:lab-grounded-glyph-probe", true},
		{"role:primary", "digest:lab-grounded-glyph-probe", true},
	}
	for _, c := range cases {
		d := Def{Model: ModelDef{ModelID: c.modelID}, Image: c.image}
		if got := d.hasRefs(); got != c.want {
			t.Errorf("hasRefs(model=%q image=%q) = %v, want %v", c.modelID, c.image, got, c.want)
		}
	}
}

func TestResolveRefsModelRoleFillsVersion(t *testing.T) {
	d := Def{Model: ModelDef{ModelID: "role:primary"}, Image: "localhost/x@sha256:a"}
	r := &fakeResolver{modelID: "Qwen3.8-27B-Q4_K_M.gguf", version: "qwen3.8-27b-q4km"}
	if err := d.resolveRefs(r); err != nil {
		t.Fatalf("resolveRefs: %v", err)
	}
	if d.Model.ModelID != "Qwen3.8-27B-Q4_K_M.gguf" {
		t.Errorf("model_id = %q, want concrete gguf", d.Model.ModelID)
	}
	if d.Model.Version != "qwen3.8-27b-q4km" {
		t.Errorf("version = %q, want shelf version filled in", d.Model.Version)
	}
	if d.Image != "localhost/x@sha256:a" {
		t.Errorf("image = %q, want untouched literal", d.Image)
	}
	if r.imageCalls != 0 {
		t.Errorf("image resolver consulted for a literal image (%d calls)", r.imageCalls)
	}
}

func TestResolveRefsModelRoleKeepsExplicitVersion(t *testing.T) {
	d := Def{Model: ModelDef{ModelID: "role:primary", Version: "author-pinned"}}
	r := &fakeResolver{modelID: "Qwen3.8-27B-Q4_K_M.gguf", version: "qwen3.8-27b-q4km"}
	if err := d.resolveRefs(r); err != nil {
		t.Fatalf("resolveRefs: %v", err)
	}
	if d.Model.Version != "author-pinned" {
		t.Errorf("version = %q, want the author's explicit version kept", d.Model.Version)
	}
}

func TestResolveRefsImageDigest(t *testing.T) {
	d := Def{Model: ModelDef{ModelID: "Qwen3.8-27B-Q4_K_M.gguf"}, Image: "digest:lab-grounded-glyph-probe"}
	r := &fakeResolver{imageRef: "localhost/lab-grounded-glyph-probe@sha256:bbb"}
	if err := d.resolveRefs(r); err != nil {
		t.Fatalf("resolveRefs: %v", err)
	}
	if d.Image != "localhost/lab-grounded-glyph-probe@sha256:bbb" {
		t.Errorf("image = %q, want resolved digest ref", d.Image)
	}
	if r.modelCalls != 0 {
		t.Errorf("model resolver consulted for a literal model_id (%d calls)", r.modelCalls)
	}
}

func TestResolveRefsNoRefsConsultsNothing(t *testing.T) {
	d := Def{Model: ModelDef{ModelID: "Qwen3.8-27B-Q4_K_M.gguf"}, Image: "localhost/x@sha256:a"}
	r := &fakeResolver{}
	if err := d.resolveRefs(r); err != nil {
		t.Fatalf("resolveRefs: %v", err)
	}
	if r.modelCalls != 0 || r.imageCalls != 0 {
		t.Errorf("literal def consulted the resolver: model=%d image=%d", r.modelCalls, r.imageCalls)
	}
}

func TestResolveRefsPropagatesErrors(t *testing.T) {
	sentinel := errors.New("boom")
	d := Def{Model: ModelDef{ModelID: "role:primary"}}
	if err := d.resolveRefs(&fakeResolver{modelErr: sentinel}); !errors.Is(err, sentinel) {
		t.Errorf("model error not propagated: %v", err)
	}
	d = Def{Image: "digest:lab-grounded-glyph-probe"}
	if err := d.resolveRefs(&fakeResolver{imageErr: sentinel}); !errors.Is(err, sentinel) {
		t.Errorf("image error not propagated: %v", err)
	}
}

func writeFile(t *testing.T, dir, name, body string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

const testShelf = `
[roles.primary]
model_id = "Qwen3.8-27B-Q4_K_M.gguf"
version = "qwen3.8-27b-q4km"
`

const testDigests = `# corpos-lab assay image digests
# comment line
lab-base:dev sha256:aaa
lab-grounded-glyph-probe:dev sha256:bbb
`

func TestFileResolverModel(t *testing.T) {
	dir := t.TempDir()
	shelf := writeFile(t, dir, "shelf.toml", testShelf)
	r := &fileResolver{shelfPath: shelf}

	id, ver, err := r.model("primary")
	if err != nil {
		t.Fatalf("model(primary): %v", err)
	}
	if id != "Qwen3.8-27B-Q4_K_M.gguf" || ver != "qwen3.8-27b-q4km" {
		t.Fatalf("model(primary) = %q %q", id, ver)
	}
	// Second call is served from the cache, not a re-read.
	if _, _, err := r.model("primary"); err != nil {
		t.Fatalf("cached model(primary): %v", err)
	}
	if _, _, err := r.model("nonesuch"); err == nil {
		t.Error("expected error for an unknown role")
	}
}

func TestFileResolverModelEmptyModelID(t *testing.T) {
	dir := t.TempDir()
	shelf := writeFile(t, dir, "shelf.toml", "[roles.blank]\nversion = \"v\"\n")
	r := &fileResolver{shelfPath: shelf}
	if _, _, err := r.model("blank"); err == nil {
		t.Error("expected error for a role with no model_id")
	}
}

func TestFileResolverModelNoRolesTable(t *testing.T) {
	dir := t.TempDir()
	// A shelf file that parses but declares no [roles] table: every lookup is an
	// unknown-role error, not a nil-map panic.
	shelf := writeFile(t, dir, "shelf.toml", "# no roles here\n")
	r := &fileResolver{shelfPath: shelf}
	if _, _, err := r.model("primary"); err == nil {
		t.Error("expected unknown-role error from a shelf with no roles")
	}
}

func TestFileResolverModelBadFile(t *testing.T) {
	r := &fileResolver{shelfPath: filepath.Join(t.TempDir(), "absent.toml")}
	if _, _, err := r.model("primary"); err == nil {
		t.Error("expected error for a missing shelf file")
	}
	dir := t.TempDir()
	r = &fileResolver{shelfPath: writeFile(t, dir, "bad.toml", "roles = = broken")}
	if _, _, err := r.model("primary"); err == nil {
		t.Error("expected error for malformed shelf TOML")
	}
}

func TestFileResolverImage(t *testing.T) {
	dir := t.TempDir()
	digests := writeFile(t, dir, "digests.txt", testDigests)
	r := &fileResolver{digestsPath: digests}

	ref, err := r.image("lab-grounded-glyph-probe")
	if err != nil {
		t.Fatalf("image: %v", err)
	}
	if ref != "localhost/lab-grounded-glyph-probe@sha256:bbb" {
		t.Fatalf("image ref = %q", ref)
	}
	// Second call is cached.
	if _, err := r.image("lab-base"); err != nil {
		t.Fatalf("cached image(lab-base): %v", err)
	}
	if _, err := r.image("no-such-image"); err == nil {
		t.Error("expected error for an image absent from the digests file")
	}
}

func TestFileResolverImageBadFile(t *testing.T) {
	r := &fileResolver{digestsPath: filepath.Join(t.TempDir(), "absent.txt")}
	if _, err := r.image("lab-base"); err == nil {
		t.Error("expected error for a missing digests file")
	}
	dir := t.TempDir()
	r = &fileResolver{digestsPath: writeFile(t, dir, "bad.txt", "one two three four\n")}
	if _, err := r.image("one"); err == nil {
		t.Error("expected error for a malformed digests line")
	}
}

func TestFindRepoRoot(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "deploy"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(root, "deploy"), "IMAGE_DIGESTS.txt", testDigests)
	start := filepath.Join(root, "studies", "some-study")
	if err := os.MkdirAll(start, 0o755); err != nil {
		t.Fatal(err)
	}
	got, err := findRepoRoot(start)
	if err != nil {
		t.Fatalf("findRepoRoot: %v", err)
	}
	// t.TempDir may sit under a symlinked path (e.g. macOS /var); compare by the
	// marker file both paths resolve to rather than the raw strings.
	if _, err := os.Stat(filepath.Join(got, digestsRel)); err != nil {
		t.Fatalf("findRepoRoot returned %q, which has no %s: %v", got, digestsRel, err)
	}

	if _, err := findRepoRoot(t.TempDir()); err == nil {
		t.Error("expected error when no deploy/ exists above the start dir")
	}
}

func TestNewFileResolverEnvOverride(t *testing.T) {
	t.Setenv(shelfEnv, "/custom/shelf.toml")
	t.Setenv(digestsEnv, "/custom/digests.txt")
	r, err := newFileResolver("/anywhere")
	if err != nil {
		t.Fatalf("newFileResolver: %v", err)
	}
	if r.shelfPath != "/custom/shelf.toml" || r.digestsPath != "/custom/digests.txt" {
		t.Fatalf("env override ignored: %+v", r)
	}
}

func TestNewFileResolverWalksToRoot(t *testing.T) {
	t.Setenv(shelfEnv, "")
	t.Setenv(digestsEnv, "")
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "deploy"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(root, "deploy"), "IMAGE_DIGESTS.txt", testDigests)
	start := filepath.Join(root, "studies", "s")
	if err := os.MkdirAll(start, 0o755); err != nil {
		t.Fatal(err)
	}
	r, err := newFileResolver(start)
	if err != nil {
		t.Fatalf("newFileResolver: %v", err)
	}
	if !strings.HasSuffix(r.digestsPath, digestsRel) {
		t.Fatalf("digestsPath = %q, want it under %s", r.digestsPath, digestsRel)
	}
}

func TestNewFileResolverNoRootFails(t *testing.T) {
	t.Setenv(shelfEnv, "")
	t.Setenv(digestsEnv, "")
	if _, err := newFileResolver(t.TempDir()); err == nil {
		t.Error("expected error when no deploy/ config is found and no override is set")
	}
}

// roleDef references the shelf model by role and the image by built variant, so
// LoadDef must resolve both to concrete values.
const roleDef = `
name = "role-smoke"
assay = "grounded-glyph-probe"
item_id = "casg-direct"
image = "digest:lab-grounded-glyph-probe"
network = "corpos-net"
conditions = ["baseline", "glyph_only", "grounded_glyph"]
runs_per_cell = 2

[model]
base_url = "http://llama-server:8081/v1"
model_id = "role:primary"

[materials]
scenario = "scenario.md"
glyph = "glyph.md"
ground = "ground.md"

[sampling]
temperature = 0.8
seeds = [1, 2]
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

func TestLoadDefResolvesRoleAndDigest(t *testing.T) {
	cfg := t.TempDir()
	t.Setenv(shelfEnv, writeFile(t, cfg, "shelf.toml", testShelf))
	t.Setenv(digestsEnv, writeFile(t, cfg, "digests.txt", testDigests))

	d, err := LoadDef(writeDef(t, roleDef, allMaterials()))
	if err != nil {
		t.Fatalf("LoadDef: %v", err)
	}
	if d.Model.ModelID != "Qwen3.8-27B-Q4_K_M.gguf" {
		t.Errorf("model_id = %q, want resolved concrete gguf", d.Model.ModelID)
	}
	if d.Model.Version != "qwen3.8-27b-q4km" {
		t.Errorf("version = %q, want resolved from shelf", d.Model.Version)
	}
	if d.Image != "localhost/lab-grounded-glyph-probe@sha256:bbb" {
		t.Errorf("image = %q, want resolved digest ref", d.Image)
	}
}

func TestLoadDefRefWithoutConfigFails(t *testing.T) {
	// A definition that names a reference but sits where no deploy/ config can be
	// found, with no override set, fails at load rather than running a guess.
	t.Setenv(shelfEnv, "")
	t.Setenv(digestsEnv, "")
	if _, err := LoadDef(writeDef(t, roleDef, allMaterials())); err == nil {
		t.Error("expected LoadDef to fail when a reference cannot be resolved")
	}
}

func TestLoadDefUnknownRoleFails(t *testing.T) {
	cfg := t.TempDir()
	t.Setenv(shelfEnv, writeFile(t, cfg, "shelf.toml", testShelf))
	t.Setenv(digestsEnv, writeFile(t, cfg, "digests.txt", testDigests))

	body := strings.Replace(roleDef, `model_id = "role:primary"`, `model_id = "role:ghost"`, 1)
	if _, err := LoadDef(writeDef(t, body, allMaterials())); err == nil {
		t.Error("expected LoadDef to fail on an unknown shelf role")
	}
}
