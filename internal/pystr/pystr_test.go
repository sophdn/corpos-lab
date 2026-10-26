package pystr

import "testing"

func TestPad(t *testing.T) {
	cases := []struct{ got, want string }{
		{Trunc("abcdef", 3), "abc"},
		{Trunc("ab", 3), "ab"},
		{Ljust("x", 3), "x  "},
		{Ljust("xyz", 2), "xyz"},
		{Rjust("x", 3), "  x"},
		{Rjust("xyz", 2), "xyz"},
	}
	for i, c := range cases {
		if c.got != c.want {
			t.Errorf("case %d = %q, want %q", i, c.got, c.want)
		}
	}
}
