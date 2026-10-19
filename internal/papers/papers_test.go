// Package papers reads papers/REGISTRY.toml, the machine-readable list of the
// lab's papers and their DOIs, and checks that CITATION.cff cites every
// published one by its concept DOI. Before the registry the canonical list lived
// only in an agent memory entry, and CITATION.cff, the corrections ledger and
// citing papers drifted from it (CITATION.cff once lacked the synthesis paper).
//
// The package is test-only: TestCommittedRegistry is the gate, and the
// registry reader below exists only for it.
package papers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
)

// Paper is one registry entry.
type Paper struct {
	Slug        string   `toml:"slug"`
	Title       string   `toml:"title"`
	ConceptDOI  string   `toml:"concept_doi"`
	LatestDOI   string   `toml:"latest_doi"`
	VersionDOIs []string `toml:"version_dois"`
	State       string   `toml:"state"`
}

// Parse reads a registry and validates each entry: a slug, a concept DOI, and
// for a published paper a latest DOI that is one of its versions.
func Parse(raw string) ([]Paper, error) {
	var reg struct {
		Paper []Paper `toml:"paper"`
	}
	if _, err := toml.Decode(raw, &reg); err != nil {
		return nil, fmt.Errorf("papers: parse registry: %w", err)
	}
	seen := map[string]bool{}
	for _, p := range reg.Paper {
		switch {
		case p.Slug == "" || p.ConceptDOI == "":
			return nil, fmt.Errorf("papers: entry %q needs slug and concept_doi", p.Slug)
		case seen[p.Slug]:
			return nil, fmt.Errorf("papers: duplicate slug %q", p.Slug)
		case p.State == "published" && !contains(p.VersionDOIs, p.LatestDOI):
			return nil, fmt.Errorf("papers: %s latest_doi %q is not in its version_dois", p.Slug, p.LatestDOI)
		case contains(p.VersionDOIs, p.ConceptDOI):
			return nil, fmt.Errorf("papers: %s lists its concept DOI as a version", p.Slug)
		}
		seen[p.Slug] = true
	}
	return reg.Paper, nil
}

// MissingFromCitation returns the slugs of published papers whose concept DOI
// does not appear in the CITATION.cff text.
func MissingFromCitation(papers []Paper, cff string) []string {
	var missing []string
	for _, p := range papers {
		if p.State == "published" && !strings.Contains(cff, p.ConceptDOI) {
			missing = append(missing, p.Slug)
		}
	}
	return missing
}

// VersionDOIsIn returns the version DOIs of registered papers that appear in
// text, each paired with the concept DOI it should be.
func VersionDOIsIn(papers []Paper, text string) map[string]string {
	out := map[string]string{}
	for _, p := range papers {
		for _, v := range p.VersionDOIs {
			if strings.Contains(text, v) {
				out[v] = p.ConceptDOI
			}
		}
	}
	return out
}

func contains(xs []string, x string) bool {
	for _, v := range xs {
		if v == x {
			return true
		}
	}
	return false
}

const reg = `
[[paper]]
slug = "a"
concept_doi = "10.5281/zenodo.100"
latest_doi = "10.5281/zenodo.102"
version_dois = ["10.5281/zenodo.101", "10.5281/zenodo.102"]
state = "published"

[[paper]]
slug = "b"
concept_doi = "10.5281/zenodo.200"
state = "drafting"
`

func TestParseAndChecks(t *testing.T) {
	ps, err := Parse(reg)
	if err != nil {
		t.Fatal(err)
	}
	if len(ps) != 2 {
		t.Fatalf("papers = %d", len(ps))
	}
	if m := MissingFromCitation(ps, "doi: 10.5281/zenodo.100"); len(m) != 0 {
		t.Errorf("missing = %v, want none (drafting paper b is not required)", m)
	}
	if m := MissingFromCitation(ps, "nothing"); len(m) != 1 || m[0] != "a" {
		t.Errorf("missing = %v, want [a]", m)
	}
	if v := VersionDOIsIn(ps, "see 10.5281/zenodo.101 and 10.5281/zenodo.100"); len(v) != 1 || v["10.5281/zenodo.101"] != "10.5281/zenodo.100" {
		t.Errorf("version DOIs = %v", v)
	}
}

func TestParseRejects(t *testing.T) {
	for name, raw := range map[string]string{
		"bad toml":       "[[paper]\n",
		"no concept":     "[[paper]]\nslug = \"a\"\n",
		"duplicate":      "[[paper]]\nslug = \"a\"\nconcept_doi = \"x\"\n[[paper]]\nslug = \"a\"\nconcept_doi = \"y\"\n",
		"latest unknown": "[[paper]]\nslug = \"a\"\nconcept_doi = \"x\"\nlatest_doi = \"z\"\nversion_dois = [\"y\"]\nstate = \"published\"\n",
		"concept listed": "[[paper]]\nslug = \"a\"\nconcept_doi = \"x\"\nlatest_doi = \"x\"\nversion_dois = [\"x\"]\nstate = \"published\"\n",
	} {
		if _, err := Parse(raw); err == nil {
			t.Errorf("%s: want an error", name)
		}
	}
}

// The committed registry is valid, lists every paper directory, and every
// published paper is cited in CITATION.cff by its concept DOI.
func TestCommittedRegistry(t *testing.T) {
	root := filepath.Join("..", "..")
	raw, err := os.ReadFile(filepath.Join(root, "papers", "REGISTRY.toml"))
	if err != nil {
		t.Fatal(err)
	}
	ps, err := Parse(string(raw))
	if err != nil {
		t.Fatal(err)
	}
	cff, err := os.ReadFile(filepath.Join(root, "CITATION.cff"))
	if err != nil {
		t.Fatal(err)
	}
	if m := MissingFromCitation(ps, string(cff)); len(m) != 0 {
		t.Errorf("CITATION.cff omits published papers %v; add each by its concept DOI", m)
	}
	if v := VersionDOIsIn(ps, string(cff)); len(v) != 0 {
		t.Errorf("CITATION.cff cites version DOIs %v; cite the concept DOIs", v)
	}
	listed := map[string]bool{}
	for _, p := range ps {
		listed[p.Slug] = true
	}
	dirs, err := os.ReadDir(filepath.Join(root, "papers"))
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range dirs {
		if d.IsDir() && !listed[d.Name()] {
			t.Errorf("papers/%s has no registry entry", d.Name())
		}
	}
}
