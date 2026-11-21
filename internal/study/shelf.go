package study

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"

	"corpos-lab/internal/image"
)

// A study definition may name two things by reference instead of by literal
// value, so a shelf upgrade or an image rebuild is a one-line change rather than
// a hand-edit across every study:
//
//   - model_id = "role:primary"        → the concrete gguf for that shelf role,
//     from deploy/shelf.toml.
//   - image    = "digest:lab-grounded-glyph-probe" → the pinned
//     localhost/<name>@sha256:<digest> ref, from deploy/IMAGE_DIGESTS.txt.
//
// LoadDef resolves both to their concrete values before it validates the
// definition, so every downstream consumer — the container spec and the run
// record — sees the concrete model and digest, never the reference token. That
// keeps "record what ran — observe, don't assert" true by construction: the
// reference is an authoring convenience, the record stays concrete.
//
// Both forms are opt-in. A literal model_id and a literal image digest keep
// working with no config file present — a definition that names no reference
// never reads deploy/ at all.
const (
	rolePrefix   = "role:"
	digestPrefix = "digest:"

	// shelfEnv and digestsEnv override the file locations. They exist for tests
	// and non-standard layouts; normally LoadDef finds the files under the repo's
	// deploy/ directory.
	shelfEnv   = "CORPOS_LAB_SHELF"
	digestsEnv = "CORPOS_LAB_IMAGE_DIGESTS"

	shelfRel   = "deploy/shelf.toml"
	digestsRel = "deploy/IMAGE_DIGESTS.txt"
)

// refResolver resolves a study's role and digest references to concrete values.
// It is an interface so LoadDef's file-backed resolver and the tests share one
// resolution path.
type refResolver interface {
	model(role string) (modelID, version string, err error)
	image(name string) (ref string, err error)
}

// hasRefs reports whether the definition uses any load-time reference that must
// be resolved before it can be validated.
func (d Def) hasRefs() bool {
	return strings.HasPrefix(d.Model.ModelID, rolePrefix) ||
		strings.HasPrefix(d.Image, digestPrefix)
}

// resolveRefs replaces role and digest references in place with the concrete
// values from r. A field that is not a reference is left untouched, and the
// version from the shelf fills in only when the study did not state its own —
// an explicit version is an author override and is kept.
func (d *Def) resolveRefs(r refResolver) error {
	if role, ok := strings.CutPrefix(d.Model.ModelID, rolePrefix); ok {
		modelID, version, err := r.model(role)
		if err != nil {
			return err
		}
		d.Model.ModelID = modelID
		if d.Model.Version == "" {
			d.Model.Version = version
		}
	}
	if name, ok := strings.CutPrefix(d.Image, digestPrefix); ok {
		ref, err := r.image(name)
		if err != nil {
			return err
		}
		d.Image = ref
	}
	return nil
}

// shelfEntry maps one shelf role to a concrete model.
type shelfEntry struct {
	ModelID string `toml:"model_id"`
	Version string `toml:"version"`
}

// fileResolver reads the shelf and image-digest files under a repo root. Each
// file is read at most once, on first use, so a definition that references only
// one of the two never requires the other to exist.
type fileResolver struct {
	shelfPath   string
	digestsPath string

	roles   map[string]shelfEntry
	digests map[string]string
}

// newFileResolver builds a resolver for the repo that contains the study at
// baseDir. It honors the CORPOS_LAB_SHELF and CORPOS_LAB_IMAGE_DIGESTS
// overrides; otherwise it walks up from baseDir to the first directory that
// holds deploy/IMAGE_DIGESTS.txt and reads both files from that deploy/.
func newFileResolver(baseDir string) (*fileResolver, error) {
	shelfPath := os.Getenv(shelfEnv)
	digestsPath := os.Getenv(digestsEnv)
	if shelfPath == "" || digestsPath == "" {
		root, err := findRepoRoot(baseDir)
		if err != nil {
			return nil, err
		}
		if shelfPath == "" {
			shelfPath = filepath.Join(root, shelfRel)
		}
		if digestsPath == "" {
			digestsPath = filepath.Join(root, digestsRel)
		}
	}
	return &fileResolver{shelfPath: shelfPath, digestsPath: digestsPath}, nil
}

// findRepoRoot walks up from dir to the first directory that holds
// deploy/IMAGE_DIGESTS.txt.
func findRepoRoot(dir string) (string, error) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("study: resolve base dir %q: %w", dir, err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, digestsRel)); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("study: no %s found above the study dir — a role: or "+
				"digest: reference needs the repo's deploy/ config; set %s and %s, or use "+
				"literal values", digestsRel, shelfEnv, digestsEnv)
		}
		dir = parent
	}
}

func (r *fileResolver) loadShelf() error {
	if r.roles != nil {
		return nil
	}
	raw, err := os.ReadFile(r.shelfPath)
	if err != nil {
		return fmt.Errorf("study: read shelf %s: %w", r.shelfPath, err)
	}
	var s struct {
		Roles map[string]shelfEntry `toml:"roles"`
	}
	if err := toml.Unmarshal(raw, &s); err != nil {
		return fmt.Errorf("study: parse shelf %s: %w", r.shelfPath, err)
	}
	if s.Roles == nil {
		s.Roles = map[string]shelfEntry{}
	}
	r.roles = s.Roles
	return nil
}

func (r *fileResolver) model(role string) (string, string, error) {
	if err := r.loadShelf(); err != nil {
		return "", "", err
	}
	e, ok := r.roles[role]
	if !ok {
		return "", "", fmt.Errorf("study: model_id references shelf role %q, absent from %s — "+
			"the shelf is the single source of truth for role→model; add the role there or use "+
			"a literal model_id", role, r.shelfPath)
	}
	if e.ModelID == "" {
		return "", "", fmt.Errorf("study: shelf role %q in %s has no model_id", role, r.shelfPath)
	}
	return e.ModelID, e.Version, nil
}

func (r *fileResolver) loadDigests() error {
	if r.digests != nil {
		return nil
	}
	raw, err := os.ReadFile(r.digestsPath)
	if err != nil {
		return fmt.Errorf("study: read image digests %s: %w", r.digestsPath, err)
	}
	m, err := image.ParseDigestFile(raw)
	if err != nil {
		return fmt.Errorf("study: image digests %s: %w", r.digestsPath, err)
	}
	r.digests = m
	return nil
}

func (r *fileResolver) image(name string) (string, error) {
	if err := r.loadDigests(); err != nil {
		return "", err
	}
	digest, ok := r.digests[name]
	if !ok {
		return "", fmt.Errorf("study: image references built variant %q, absent from %s — "+
			"run scripts/build-lab-images.sh to record it, or use a literal image digest",
			name, r.digestsPath)
	}
	return "localhost/" + name + "@" + digest, nil
}

// ShelfVersions reads the model shelf and returns a map from concrete model_id to
// its recorded version — the reverse of role resolution. baseline capture records
// the model the server reports and looks its version up here, so the provenance is
// complete without a hand-passed -version flag.
//
// It honors the CORPOS_LAB_SHELF override; otherwise it walks up from baseDir to
// the first deploy/shelf.toml. A missing shelf is NOT an error: it returns an
// empty map so an unresolved version stays a recorded gap, never a run failure
// (record-what-ran). A shelf that exists but does not parse IS an error. Only
// entries that carry both a model_id and a version are indexed.
func ShelfVersions(baseDir string) (map[string]string, error) {
	shelfPath := os.Getenv(shelfEnv)
	if shelfPath == "" {
		p, ok := findShelf(baseDir)
		if !ok {
			return map[string]string{}, nil
		}
		shelfPath = p
	}
	raw, err := os.ReadFile(shelfPath) //nolint:gosec // a config path, not a secret
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]string{}, nil
		}
		return nil, fmt.Errorf("study: read shelf %s: %w", shelfPath, err)
	}
	var s struct {
		Roles map[string]shelfEntry `toml:"roles"`
	}
	if err := toml.Unmarshal(raw, &s); err != nil {
		return nil, fmt.Errorf("study: parse shelf %s: %w", shelfPath, err)
	}
	out := map[string]string{}
	for _, e := range s.Roles {
		if e.ModelID != "" && e.Version != "" {
			out[e.ModelID] = e.Version
		}
	}
	return out, nil
}

// findShelf walks up from dir to the first directory holding deploy/shelf.toml,
// reporting the path and whether one was found. Unlike findRepoRoot it never
// errors — an absent shelf is a valid, non-fatal state for a version lookup.
func findShelf(dir string) (string, bool) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", false
	}
	for {
		p := filepath.Join(abs, shelfRel)
		if _, err := os.Stat(p); err == nil {
			return p, true
		}
		parent := filepath.Dir(abs)
		if parent == abs {
			return "", false
		}
		abs = parent
	}
}

// ShelfRole is one role on the model shelf.
type ShelfRole struct {
	Role    string
	ModelID string
}

// ShelfRoles parses a shelf file's contents and returns its roles in the order a
// shelf sweep runs them: every non-primary role by name, then "primary" last,
// so the sweep ends with the default model served. A role without a model_id is
// an error.
func ShelfRoles(raw string) ([]ShelfRole, error) {
	var s struct {
		Roles map[string]shelfEntry `toml:"roles"`
	}
	if err := toml.Unmarshal([]byte(raw), &s); err != nil {
		return nil, fmt.Errorf("study: parse shelf: %w", err)
	}
	var roles []ShelfRole
	var primary *ShelfRole
	for name, e := range s.Roles {
		if e.ModelID == "" {
			return nil, fmt.Errorf("study: shelf role %q has no model_id", name)
		}
		r := ShelfRole{Role: name, ModelID: e.ModelID}
		if name == "primary" {
			primary = &r
			continue
		}
		roles = append(roles, r)
	}
	sort.Slice(roles, func(i, j int) bool { return roles[i].Role < roles[j].Role })
	if primary != nil {
		roles = append(roles, *primary)
	}
	return roles, nil
}
