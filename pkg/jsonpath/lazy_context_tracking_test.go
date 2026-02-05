//go:build lazytracking

package jsonpath

import (
	"github.com/pb33f/jsonpath/pkg/jsonpath/config"
	"github.com/pb33f/jsonpath/pkg/jsonpath/token"
	"go.yaml.in/yaml/v4"
	"testing"
)

func TestPropertyNameQueryLazyTracking(t *testing.T) {
	yamlData := `
store: book-store
`
	var root yaml.Node
	err := yaml.Unmarshal([]byte(yamlData), &root)
	if err != nil {
		t.Fatalf("Error parsing YAML: %v", err)
	}

	tokenizer := token.NewTokenizer("$.store~", config.WithPropertyNameExtension(), config.WithLazyContextTracking())
	parser := newParserPrivate(tokenizer, tokenizer.Tokenize(), config.WithPropertyNameExtension(), config.WithLazyContextTracking())
	err = parser.parse()
	if err != nil {
		t.Fatalf("Error parsing JSON Path: %v", err)
	}

	result := parser.ast.Query(&root, &root)
	if len(result) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(result))
	}
	actual := nodeToString(result[0])
	if actual != "store" {
		t.Fatalf("Expected %q, got %q", "store", actual)
	}
}

func BenchmarkLazyContextTracking(b *testing.B) {
	yamlData := `
store:
  book:
    - title: "Book 1"
    - title: "Book 2"
  bicycle:
    details: { price: 20 }
paths:
  /users:
    get: { summary: "Get users" }
    post: { summary: "Create user" }
  /orders:
    get: { summary: "Get orders" }
items:
  - name: "First"
  - name: "Second"
  - name: "Third"
`
	var root yaml.Node
	if err := yaml.Unmarshal([]byte(yamlData), &root); err != nil {
		b.Fatalf("Error parsing YAML: %v", err)
	}

	bench := func(b *testing.B, opts ...config.Option) {
		tokenizer := token.NewTokenizer("$.store.book[*].title", opts...)
		parser := newParserPrivate(tokenizer, tokenizer.Tokenize(), opts...)
		if err := parser.parse(); err != nil {
			b.Fatalf("Error parsing JSON Path: %v", err)
		}

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = parser.ast.Query(&root, &root)
		}
	}

	b.Run("EagerTracking", func(b *testing.B) {
		bench(b)
	})
	b.Run("LazyTracking", func(b *testing.B) {
		bench(b, config.WithLazyContextTracking())
	})
}
