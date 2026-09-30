package neutralprefix

import "testing"

func TestOCounter(t *testing.T) {
	c := NewOCounter()
	if c.PyDictRepr() != "{}" {
		t.Fatalf("empty repr = %q", c.PyDictRepr())
	}
	c.Inc("C")
	c.Inc("N")
	c.Inc("C")
	if c.Get("C") != 2 || c.Get("N") != 1 || c.Get("absent") != 0 {
		t.Fatalf("counts wrong: C=%d N=%d", c.Get("C"), c.Get("N"))
	}
	// first-seen order: C before N.
	if len(c.Keys) != 2 || c.Keys[0] != "C" || c.Keys[1] != "N" {
		t.Fatalf("keys order wrong: %v", c.Keys)
	}
	if c.PyDictRepr() != "{'C': 2, 'N': 1}" {
		t.Fatalf("repr = %q", c.PyDictRepr())
	}
	m := c.ToOMap()
	if m.Len() != 2 || m.keys[0] != "C" || m.vals[0].(int) != 2 {
		t.Fatalf("ToOMap wrong: %s", PyDumps(m, -1, true))
	}
}
