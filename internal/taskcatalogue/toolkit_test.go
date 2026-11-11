package taskcatalogue

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

// fakeLedger answers db_query by matching a substring of the SQL. Each answer
// is a function of the OFFSET so paging can be exercised.
type fakeLedger map[string]func(offset int) any

func (f fakeLedger) server(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/mcp/admin" {
			t.Errorf("path %s", r.URL.Path)
		}
		var env queryEnvelope
		if err := json.NewDecoder(r.Body).Decode(&env); err != nil {
			t.Errorf("decode: %v", err)
		}
		if env.Action != "db_query" || env.Rationale == "" {
			t.Errorf("envelope %+v", env)
		}
		var offset int
		if i := strings.Index(env.Params.SQL, "OFFSET "); i >= 0 {
			_, _ = fmt.Sscan(env.Params.SQL[i+7:], &offset)
		}
		for key, fn := range f {
			if strings.Contains(env.Params.SQL, key) {
				_ = json.NewEncoder(w).Encode(fn(offset))
				return
			}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"rows": [][]any{}})
	}))
}

func rows(r ...[]any) func(int) any {
	return func(offset int) any {
		if offset > 0 {
			return map[string]any{"rows": [][]any{}}
		}
		return map[string]any{"rows": r}
	}
}

func TestRowsPages(t *testing.T) {
	srv := fakeLedger{"FROM big": func(offset int) any {
		n := pageSize
		if offset > 0 {
			n = 3
		}
		out := make([][]any, n)
		for i := range out {
			out[i] = []any{offset + i, nil}
		}
		return map[string]any{"rows": out, "truncated": offset == 0, "truncated_reason": "row limit 1000 reached"}
	}}.server(t)
	defer srv.Close()
	got, err := Toolkit{BaseURL: srv.URL + "/"}.Rows(context.Background(), "SELECT a, b FROM big ORDER BY a")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != pageSize+3 || got[pageSize+2][0] != "1002" || got[0][1] != "" {
		t.Fatalf("got %d rows, last %v", len(got), got[len(got)-1])
	}
}

func TestRowsErrors(t *testing.T) {
	cases := map[string]http.HandlerFunc{
		"status 500": func(w http.ResponseWriter, _ *http.Request) { http.Error(w, "boom", 500) },
		"non-JSON":   func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("<html>")) },
		"refused": func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"error":"db_query runs one SELECT"}`))
		},
		"byte cut": func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"rows":[["a"]],"truncated":true,"truncated_reason":"byte limit 50 reached"}`))
		},
	}
	for name, h := range cases {
		srv := httptest.NewServer(h)
		_, err := Toolkit{BaseURL: srv.URL, HTTP: srv.Client()}.Rows(context.Background(), "SELECT 1")
		srv.Close()
		if err == nil {
			t.Errorf("%s: want error", name)
		}
	}
	if _, err := (Toolkit{BaseURL: "http://127.0.0.1:1"}).Rows(context.Background(), "SELECT 1"); err == nil {
		t.Error("unreachable server: want error")
	}
	if _, err := (Toolkit{BaseURL: "http://bad host"}).Rows(context.Background(), "SELECT 1"); err == nil {
		t.Error("bad URL: want error")
	}
}

func TestLedgerTerms(t *testing.T) {
	srv := fakeLedger{
		"FROM projects": rows([]any{"alpha", "Alpha Project"}, []any{"beta", "beta"}),
		"FROM canon_aliases": rows(
			[]any{"old-alpha", "alpha (the gitea repo)"},
			[]any{"ghost", "nowhere"},
			[]any{"blank", ""},
		),
		"FROM proj_chain_status":        rows([]any{"short-slug", "alpha"}, []any{"a-long-chain-slug", "alpha"}),
		"FROM proj_current_bugs":        rows([]any{"bug-in-the-thing", "beta"}),
		"FROM proj_current_suggestions": rows(),
		"FROM proj_current_proposals":   rows(),
	}.server(t)
	defer srv.Close()
	got, err := Toolkit{BaseURL: srv.URL}.LedgerTerms(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	want := []Term{
		{Text: "alpha", Project: "alpha"}, {Text: "Alpha Project", Project: "alpha"},
		{Text: "beta", Project: "beta"}, {Text: "beta", Project: "beta"},
		{Text: "old-alpha", Project: "alpha"},
		{Text: "a-long-chain-slug", Project: "alpha"},
		{Text: "bug-in-the-thing", Project: "beta"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v\nwant %v", got, want)
	}
}

func TestLedgerTermsPropagatesErrors(t *testing.T) {
	for _, failing := range []string{"FROM projects", "FROM canon_aliases", "FROM proj_chain_status"} {
		ledger := fakeLedger{
			"FROM projects":      rows([]any{"alpha", "alpha"}),
			"FROM canon_aliases": rows(),
		}
		ledger[failing] = func(int) any { return map[string]any{"error": "nope"} }
		srv := ledger.server(t)
		_, err := Toolkit{BaseURL: srv.URL}.LedgerTerms(context.Background())
		srv.Close()
		if err == nil {
			t.Errorf("%s failing: want error", failing)
		}
	}
}

func TestSessionProjects(t *testing.T) {
	srv := fakeLedger{"FROM tool_calls": rows(
		[]any{"s1", "alpha", 3}, []any{"s1", "beta", 5}, []any{"s2", "alpha", 1},
	)}.server(t)
	defer srv.Close()
	got, err := Toolkit{BaseURL: srv.URL}.SessionProjects(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, map[string]string{"s1": "beta", "s2": "alpha"}) {
		t.Errorf("got %v", got)
	}

	bad := fakeLedger{"FROM tool_calls": rows([]any{"s1", "alpha", "many"})}.server(t)
	defer bad.Close()
	if _, err := (Toolkit{BaseURL: bad.URL}).SessionProjects(context.Background()); err == nil {
		t.Error("non-numeric count: want error")
	}
	failing := fakeLedger{"FROM tool_calls": func(int) any { return map[string]any{"error": "nope"} }}.server(t)
	defer failing.Close()
	if _, err := (Toolkit{BaseURL: failing.URL}).SessionProjects(context.Background()); err == nil {
		t.Error("query error: want error")
	}
}
