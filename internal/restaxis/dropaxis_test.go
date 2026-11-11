package restaxis

import (
	"reflect"
	"testing"
)

// TestDropAxis covers the drop cases against outputs computed by the python
// oracle (drop_axis.py) on the same inputs, so the port is verified byte-for-byte.
func TestDropAxis(t *testing.T) {
	cases := []struct {
		name      string
		axis      string
		in        string
		wantOut   string
		wantFound bool
	}{
		{
			name:      "heading style, block ends at next heading",
			axis:      "Rest",
			in:        "# Glyph\n\n## Marker\nmarker text\n\n## Aim\naim text\n\n## Rest\nrest text\n\n## Notes\nnotes text\n",
			wantOut:   "# Glyph\n\n## Marker\nmarker text\n\n## Aim\naim text\n\n## Notes\nnotes text\n",
			wantFound: true,
		},
		{
			name:      "heading style, block ends at EOF",
			axis:      "Rest",
			in:        "# Glyph\n\n## Aim\naim text\n\n## Rest\nrest text\nmore rest\n",
			wantOut:   "# Glyph\n\n## Aim\naim text\n",
			wantFound: true,
		},
		{
			name:      "block ends at the next axis section, keeping its separator",
			axis:      "Rest",
			in:        "## Aim\naim text\n\n---\n\n## Rest\nrest text\n\n---\n\n## Marker\nm\n",
			wantOut:   "## Aim\naim text\n\n---\n\n## Marker\nm\n",
			wantFound: true,
		},
		{
			name:      "bold label **Rest**",
			axis:      "Rest",
			in:        "**Aim**\naim text\n\n**Rest**\nrest text\n",
			wantOut:   "**Aim**\naim text\n",
			wantFound: true,
		},
		{
			name:      "bold label **Rest:** with inline text",
			axis:      "Rest",
			in:        "intro\n\n**Aim:** a\n\n**Rest:** r\n",
			wantOut:   "intro\n\n**Aim:** a\n",
			wantFound: true,
		},
		{
			name:      "axis absent, text unchanged",
			axis:      "Rest",
			in:        "# Glyph\n\n## Aim\naim text\n",
			wantOut:   "# Glyph\n\n## Aim\naim text\n",
			wantFound: false,
		},
		{
			name:      "axis at start of file",
			axis:      "Rest",
			in:        "## Rest\nrest text\n\n## Notes\nn\n",
			wantOut:   "## Notes\nn\n",
			wantFound: true,
		},
		{
			name:      "doubled blank before removed block is preserved (seam is the tail heading)",
			axis:      "Rest",
			in:        "## Aim\na\n\n\n## Rest\nr\n\n\n## Notes\nn\n",
			wantOut:   "## Aim\na\n\n\n## Notes\nn\n",
			wantFound: true,
		},
		{
			name:      "dangling separator at EOF is dropped",
			axis:      "Rest",
			in:        "## Aim\na\n\n---\n\n## Rest\nr\n\n---\n",
			wantOut:   "## Aim\na\n",
			wantFound: true,
		},
		{
			name:      "custom axis (Aim), ends at next axis (Rest)",
			axis:      "Aim",
			in:        "## Marker\nm\n\n## Aim\na\n\n## Rest\nr\n",
			wantOut:   "## Marker\nm\n\n## Rest\nr\n",
			wantFound: true,
		},
		{
			name:      "axis name is case-insensitive",
			axis:      "REST",
			in:        "## Aim\na\n\n## Rest\nr\n",
			wantOut:   "## Aim\na\n",
			wantFound: true,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			out, found := DropAxis(c.in, c.axis)
			if found != c.wantFound {
				t.Errorf("found = %v, want %v", found, c.wantFound)
			}
			if out != c.wantOut {
				t.Errorf("output mismatch\n got: %q\nwant: %q", out, c.wantOut)
			}
		})
	}
}

// TestSectionAxis checks the section-start classifier directly, including the
// non-axis and non-section rejections.
func TestSectionAxis(t *testing.T) {
	cases := []struct {
		line string
		want string
	}{
		{"## Rest axis", "rest"},
		{"###### marker", "marker"},
		{"**Aim**", "aim"},
		{"**Rest:** something", "rest"},
		{"## Notes on behaviour", ""},         // heading, but not an axis word
		{"**Restored 2026** from backup", ""}, // bold word "restored", not "rest"
		{"just a paragraph", ""},              // not a section start
		{"", ""},                              // blank line
		{"#NoSpace heading", ""},              // no space after hashes: not a heading
		{"####### seven hashes", ""},          // seven hashes, then no space to match
	}
	for _, c := range cases {
		if got := sectionAxis(c.line); got != c.want {
			t.Errorf("sectionAxis(%q) = %q, want %q", c.line, got, c.want)
		}
	}
}

// TestTidySeamCollapse exercises the seam blank-collapse branch directly. DropAxis
// never produces a kept slice with a blank line at the seam (a block always ends
// at a non-blank heading or at EOF), so the branch is only reachable via a
// hand-built slice — this keeps the faithful port covered.
func TestTidySeam(t *testing.T) {
	cases := []struct {
		name string
		in   []string
		seam int
		want []string
	}{
		{
			name: "collapse a doubled blank at the seam to one",
			in:   []string{"a", "", "", "b"},
			seam: 1,
			want: []string{"a", "", "b"},
		},
		{
			name: "collapse a longer run of blanks to one",
			in:   []string{"a", "", "", "", "b"},
			seam: 1,
			want: []string{"a", "", "b"},
		},
		{
			name: "seam at index 0 keeps a leading blank",
			in:   []string{"", "", "b"},
			seam: 0,
			want: []string{"", "b"},
		},
		{
			name: "trailing separator and blanks are dropped",
			in:   []string{"a", "", "---", ""},
			seam: 4,
			want: []string{"a"},
		},
		{
			name: "empty slice is returned unchanged",
			in:   []string{},
			seam: 0,
			want: []string(nil),
		},
		{
			name: "all-blank/separator slice collapses to empty",
			in:   []string{"", "---", ""},
			seam: 3,
			want: []string{},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := tidySeam(append([]string(nil), c.in...), c.seam)
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("tidySeam(%q, %d) = %q, want %q", c.in, c.seam, got, c.want)
			}
		})
	}
}
