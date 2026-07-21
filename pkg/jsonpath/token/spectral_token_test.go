package token

import (
	"testing"

	"github.com/pb33f/jsonpath/pkg/jsonpath/config"
)

func TestSpectralRegexToken(t *testing.T) {
	expression := "$.values[?(\n@property.match(/a[\\/]\\d+/imsug)\n)]"
	tokens := NewTokenizer(expression, config.WithSpectralCompatibility()).Tokenize()
	var regex *TokenInfo
	for i := range tokens {
		if tokens[i].Token == REGEX {
			regex = &tokens[i]
			break
		}
	}
	if regex == nil {
		t.Fatal("missing regex token")
	}
	if regex.Literal != `/a[\/]\d+/imsug` || regex.Line != 2 || regex.Column != 16 || regex.Len != 15 {
		t.Fatalf("regex token = %+v", *regex)
	}
}

func TestSpectralRegexAndUnquotedMemberAreDistinguished(t *testing.T) {
	tokens := NewTokenizer(`$.paths[/openapi.json]`, config.WithSpectralCompatibility()).Tokenize()
	if len(tokens) != 6 || tokens[4].Token != STRING_LITERAL || tokens[4].Literal != "/openapi.json" {
		t.Fatalf("unquoted member tokens = %+v", tokens)
	}
	tokens = NewTokenizer(`$[?(@property.match(/[/]/))]`, config.WithSpectralCompatibility()).Tokenize()
	found := false
	for _, tok := range tokens {
		if tok.Token == REGEX && tok.Literal == `/[/]/` {
			found = true
		}
	}
	if !found {
		t.Fatalf("character-class regex tokens = %+v", tokens)
	}
}

func TestSpectralUnterminatedRegexToken(t *testing.T) {
	tests := []struct {
		expression string
		message    string
	}{
		{`$[?(@property.match(/abc))]`, "unterminated regex literal"},
		{`$[?(@property.match(/[abc`, "unterminated regex character class"},
	}
	for _, test := range tests {
		tokens := NewTokenizer(test.expression, config.WithSpectralCompatibility()).Tokenize()
		found := false
		for _, tok := range tokens {
			if tok.Token == ILLEGAL && tok.Literal == test.message {
				found = true
			}
		}
		if !found {
			t.Errorf("%s tokens = %+v", test.expression, tokens)
		}
	}
}

func TestStrictRFCRejectsStrictJavaScriptEqualityToken(t *testing.T) {
	for _, expression := range []string{`$[?(@.a === @.b)]`, `$[?(@.a !== @.b)]`} {
		tokens := NewTokenizer(expression, config.WithStrictRFC9535()).Tokenize()
		foundIllegal := false
		for _, tok := range tokens {
			if tok.Token == ILLEGAL {
				foundIllegal = true
			}
		}
		if !foundIllegal {
			t.Errorf("strict mode accepted %s: %+v", expression, tokens)
		}
	}
}
