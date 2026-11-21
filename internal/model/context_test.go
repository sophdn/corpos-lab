package model

import (
	"errors"
	"fmt"
	"testing"
)

const overflowBody = `{"error":{"code":400,"message":"request (24275 tokens) exceeds the available context size (16384 tokens), try increasing it","type":"exceed_context_size_error","n_prompt_tokens":24275,"n_ctx":16384}}`

// ContextOverflow reports the server body only for a 400 naming
// exceed_context_size_error, wrapped or not; anything else is not an overflow.
func TestContextOverflow(t *testing.T) {
	overflow := &APIError{ModelID: "m", Op: "completion", StatusCode: 400, Body: overflowBody}
	cases := map[string]struct {
		err      error
		wantOK   bool
		wantBody string
	}{
		"overflow":         {overflow, true, overflowBody},
		"wrapped overflow": {fmt.Errorf("turn 6: %w", overflow), true, overflowBody},
		"other 400":        {&APIError{StatusCode: 400, Body: `{"error":{"type":"invalid_request_error"}}`}, false, ""},
		"500 with text":    {&APIError{StatusCode: 500, Body: overflowBody}, false, ""},
		"plain error":      {errors.New("exceed_context_size_error"), false, ""},
		"nil":              {nil, false, ""},
	}
	for name, c := range cases {
		body, ok := ContextOverflow(c.err)
		if ok != c.wantOK || body != c.wantBody {
			t.Errorf("%s: ContextOverflow = (%q, %v), want (%q, %v)", name, body, ok, c.wantBody, c.wantOK)
		}
	}
}
