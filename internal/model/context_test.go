package model

import (
	"errors"
	"fmt"
	"testing"
)

const overflowBody = `{"error":{"code":400,"message":"request (24275 tokens) exceeds the available context size (16384 tokens), try increasing it","type":"exceed_context_size_error","n_prompt_tokens":24275,"n_ctx":16384}}`

func TestIsContextOverflow(t *testing.T) {
	overflow := &APIError{ModelID: "m", Op: "completion", StatusCode: 400, Body: overflowBody}
	cases := map[string]struct {
		err  error
		want bool
	}{
		"overflow":         {overflow, true},
		"wrapped overflow": {fmt.Errorf("turn 6: %w", overflow), true},
		"other 400":        {&APIError{StatusCode: 400, Body: `{"error":{"type":"invalid_request_error"}}`}, false},
		"500 with text":    {&APIError{StatusCode: 500, Body: overflowBody}, false},
		"plain error":      {errors.New("exceed_context_size_error"), false},
		"nil":              {nil, false},
	}
	for name, c := range cases {
		if got := IsContextOverflow(c.err); got != c.want {
			t.Errorf("%s: IsContextOverflow = %v, want %v", name, got, c.want)
		}
	}
}
