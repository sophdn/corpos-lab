package restaxis

import "testing"

func TestPyStripAndStripText(t *testing.T) {
	cases := []struct{ in, want string }{
		{"  hello  ", "hello"},
		{"\n\t hi \r\n", "hi"},
		{"none", "none"},
		{"", ""},
	}
	for _, c := range cases {
		if got := pyStrip(c.in); got != c.want {
			t.Errorf("pyStrip(%q) = %q, want %q", c.in, got, c.want)
		}
		if got := StripText(c.in); got != c.want {
			t.Errorf("StripText(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestPyRStrip(t *testing.T) {
	if got := pyRStrip("  keep left  \n"); got != "  keep left" {
		t.Errorf("pyRStrip = %q, want '  keep left'", got)
	}
	if got := pyRStrip("nochange"); got != "nochange" {
		t.Errorf("pyRStrip(nochange) = %q", got)
	}
}
