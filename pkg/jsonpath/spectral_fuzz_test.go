package jsonpath

import (
	"testing"

	"github.com/pb33f/jsonpath/pkg/jsonpath/config"
	"github.com/pb33f/jsonpath/pkg/jsonpath/token"
)

func FuzzSpectralRegexScanner(f *testing.F) {
	for _, seed := range []string{`x`, `\/`, `[a-z/]`, `\d+`, `(?=x)`, `(a)\1`, `😀`} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, pattern string) {
		if len(pattern) > 4096 {
			t.Skip()
		}
		expression := `$[?(@property.match(/` + pattern + `/))]`
		_ = token.NewTokenizer(expression, config.WithSpectralCompatibility()).Tokenize()
	})
}

func FuzzSpectralFilterCompilation(f *testing.F) {
	for _, seed := range []string{
		`@property === 'get'`,
		`@ && @.enum && @.type`,
		`!@property.match(/x/i)`,
		`@.enum.constructor.name === 'Array'`,
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, filter string) {
		if len(filter) > 4096 {
			t.Skip()
		}
		_, _ = NewPath(`$[?(`+filter+`)]`, config.WithSpectralCompatibility())
	})
}
