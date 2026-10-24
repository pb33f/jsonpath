package token

import (
	"testing"

	"github.com/pb33f/jsonpath/pkg/jsonpath/config"
	"github.com/pb33f/testify/require"
)

func TestBacktickTokens(t *testing.T) {
	for _, tc := range []struct {
		expression string
		tokens     []TokenInfo
	}{
		{"$.`$ref.name", []TokenInfo{
			{Token: ROOT, Line: 1, Column: 0, Len: 1},
			{Token: CHILD, Line: 1, Column: 1, Len: 1},
			{Token: ESCAPED_NAME, Line: 1, Column: 2, Len: 5, Literal: "$ref"},
			{Token: CHILD, Line: 1, Column: 7, Len: 1},
			{Token: STRING, Line: 1, Column: 8, Len: 4, Literal: "name"},
		}},
		{"$[`*,`$ref]", []TokenInfo{
			{Token: ROOT, Line: 1, Column: 0, Len: 1},
			{Token: BRACKET_LEFT, Line: 1, Column: 1, Len: 1},
			{Token: ESCAPED_NAME, Line: 1, Column: 2, Len: 2, Literal: "*"},
			{Token: COMMA, Line: 1, Column: 4, Len: 1},
			{Token: ESCAPED_NAME, Line: 1, Column: 5, Len: 5, Literal: "$ref"},
			{Token: BRACKET_RIGHT, Line: 1, Column: 10, Len: 1},
		}},
	} {
		t.Run(tc.expression, func(t *testing.T) {
			tokens := NewTokenizer(tc.expression).Tokenize()
			require.Equal(t, tc.tokens, []TokenInfo(tokens))
			require.False(t, tokens.IsSimple())
		})
	}
	require.Equal(t, "ESCAPED_NAME", ESCAPED_NAME.String())
	tokens := NewTokenizer("$.`$ref", config.WithStrictRFC9535()).Tokenize()
	require.Equal(t, ILLEGAL, tokens[2].Token)
}
