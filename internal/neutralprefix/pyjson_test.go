package neutralprefix

import "testing"

// Ground truth captured from CPython json.dumps.
func TestPyDumpsItemsIndent1NoASCII(t *testing.T) {
	obj := NewOMap().Set("id", "H01").Set("blocks", []any{
		NewOMap().Set("heading", "a").Set("kind", "context").Set("text", "line1\nline2 <b> & é"),
	})
	got := PyDumps([]any{obj}, 1, false)
	want := "[\n {\n  \"id\": \"H01\",\n  \"blocks\": [\n   {\n    \"heading\": \"a\",\n    \"kind\": \"context\",\n    \"text\": \"line1\\nline2 <b> & é\"\n   }\n  ]\n }\n]"
	if got != want {
		t.Fatalf("items indent=1 ensure_ascii=False:\n got %q\nwant %q", got, want)
	}
}

func TestPyDumpsKeyIndent1ASCII(t *testing.T) {
	obj := NewOMap().Set("H01", NewOMap().
		Set("resp_id", "r1").Set("run", 11).Set("claude", "Ii").Set("txt", "é"))
	got := PyDumps(obj, 1, true)
	want := "{\n \"H01\": {\n  \"resp_id\": \"r1\",\n  \"run\": 11,\n  \"claude\": \"Ii\",\n  \"txt\": \"\\u00e9\"\n }\n}"
	if got != want {
		t.Fatalf("key indent=1 ensure_ascii=True:\n got %q\nwant %q", got, want)
	}
}

func TestPyDumpsSliceCompact(t *testing.T) {
	got := PyDumps(NewOMap().Set("id", "r1").Set("text", "a\nb"), -1, true)
	want := `{"id": "r1", "text": "a\nb"}`
	if got != want {
		t.Fatalf("compact:\n got %q\nwant %q", got, want)
	}
}

func TestPyDumpsIndent0(t *testing.T) {
	obj := NewOMap().Set("a", NewOMap().Set("b", 1)).Set("c", 2)
	got := PyDumps(obj, 0, true)
	want := "{\n\"a\": {\n\"b\": 1\n},\n\"c\": 2\n}"
	if got != want {
		t.Fatalf("indent=0:\n got %q\nwant %q", got, want)
	}
}

func TestPyDumpsEmpty(t *testing.T) {
	obj := NewOMap().Set("x", NewOMap()).Set("y", []any{})
	got := PyDumps(obj, 1, true)
	want := "{\n \"x\": {},\n \"y\": []\n}"
	if got != want {
		t.Fatalf("empty containers:\n got %q\nwant %q", got, want)
	}
}

func TestPyDumpsSurrogatePair(t *testing.T) {
	// U+1F600 -> surrogate pair when ensure_ascii=True; raw when False.
	got := PyDumps("😀", -1, true)
	if got != `"\ud83d\ude00"` {
		t.Fatalf("surrogate ascii: got %q", got)
	}
	got = PyDumps("😀", -1, false)
	if got != `"😀"` {
		t.Fatalf("surrogate raw: got %q", got)
	}
}

func TestPyDumpsInt64(t *testing.T) {
	if got := PyDumps(NewOMap().Set("n", int64(42)), -1, true); got != `{"n": 42}` {
		t.Fatalf("int64: got %q", got)
	}
}

func TestPyDumpsUnsupportedTypePanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on unsupported type")
		}
	}()
	_ = PyDumps(NewOMap().Set("bad", 3.14), -1, true)
}

func TestPyDumpsControlChars(t *testing.T) {
	got := PyDumps("\x00\x1f\t\r\n\b\f", -1, true)
	want := `"\u0000\u001f\t\r\n\b\f"`
	if got != want {
		t.Fatalf("control chars:\n got %q\nwant %q", got, want)
	}
}

func TestPyDumpsCompactArray(t *testing.T) {
	if got := PyDumps([]any{1, 2, 3}, -1, true); got != "[1, 2, 3]" {
		t.Fatalf("compact array: got %q", got)
	}
}
