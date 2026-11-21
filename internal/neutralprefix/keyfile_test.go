package neutralprefix

import "testing"

func TestParseKeyFileOrder(t *testing.T) {
	data := []byte(`{
"rZ": {"cls":"casg-direct","scenario":"1","model":"mistral","condition":"baseline","run":3},
"rA": {"cls":"parent-state-check-bypass","scenario":"2","model":"phi4","condition":"glyph_only","run":7}
}`)
	kf, err := ParseKeyFile(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(kf.Order) != 2 || kf.Order[0] != "rZ" || kf.Order[1] != "rA" {
		t.Fatalf("order not preserved: %v", kf.Order)
	}
	m := kf.Meta["rA"]
	if m.Cls != "parent-state-check-bypass" || m.Scenario != "2" || m.Model != "phi4" || m.Condition != "glyph_only" || m.Run != 7 {
		t.Fatalf("meta parsed wrong: %+v", m)
	}
}

func TestParseKeyFileErrors(t *testing.T) {
	if _, err := ParseKeyFile([]byte(`[]`)); err == nil {
		t.Fatal("expected error for non-object top level")
	}
	if _, err := ParseKeyFile([]byte(`{`)); err == nil {
		t.Fatal("expected error for truncated input")
	}
	if _, err := ParseKeyFile([]byte(`not json`)); err == nil {
		t.Fatal("expected error for invalid token")
	}
	if _, err := ParseKeyFile([]byte(`{"a": "notanobject"}`)); err == nil {
		t.Fatal("expected error for non-object value")
	}
}

func TestParseKeyFileEmpty(t *testing.T) {
	if _, err := ParseKeyFile([]byte("")); err == nil {
		t.Fatal("expected error for empty input")
	}
	if _, err := ParseKeyFile([]byte("{}")); err != nil {
		t.Fatalf("empty object should parse: %v", err)
	}
}
