//go:build lazytracking && stress

package jsonpath

import (
	"fmt"
	"os"
	"testing"

	"github.com/pb33f/jsonpath/pkg/jsonpath/config"
	"github.com/pb33f/jsonpath/pkg/jsonpath/token"
	"go.yaml.in/yaml/v4"
)

func BenchmarkLazyContextTrackingStress(b *testing.B) {
	const (
		stressItems         = 200
		stressChildren      = 40
		stressGrandchildren = 40
	)
	root := buildStressRoot(stressItems, stressChildren, stressGrandchildren)

	bench := func(b *testing.B, opts ...config.Option) {
		tokenizer := token.NewTokenizer("$.items[*].children[*].children[*].value", opts...)
		parser := newParserPrivate(tokenizer, tokenizer.Tokenize(), opts...)
		if err := parser.parse(); err != nil {
			b.Fatalf("Error parsing JSON Path: %v", err)
		}

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = parser.ast.Query(root, root)
		}
	}

	switch os.Getenv("LAZY_TRACKING_MODE") {
	case "eager":
		bench(b)
		return
	case "lazy":
		bench(b, config.WithLazyContextTracking())
		return
	}

	b.Run("EagerTracking", func(b *testing.B) {
		bench(b)
	})
	b.Run("LazyTracking", func(b *testing.B) {
		bench(b, config.WithLazyContextTracking())
	})
}

func buildStressRoot(items, children, grandchildren int) *yaml.Node {
	doc := &yaml.Node{Kind: yaml.DocumentNode}
	root := &yaml.Node{Kind: yaml.MappingNode}
	doc.Content = append(doc.Content, root)

	root.Content = append(root.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Value: "items"},
		buildStressItems(items, children, grandchildren),
	)

	return doc
}

func buildStressItems(items, children, grandchildren int) *yaml.Node {
	seq := &yaml.Node{Kind: yaml.SequenceNode}
	for i := 0; i < items; i++ {
		item := &yaml.Node{Kind: yaml.MappingNode}
		item.Content = append(item.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: fmt.Sprintf("item-%d", i)},
			&yaml.Node{Kind: yaml.ScalarNode, Value: "value"},
			&yaml.Node{Kind: yaml.ScalarNode, Value: "children"},
			buildStressChildren(children, grandchildren),
		)
		seq.Content = append(seq.Content, item)
	}
	return seq
}

func buildStressChildren(children, grandchildren int) *yaml.Node {
	seq := &yaml.Node{Kind: yaml.SequenceNode}
	for i := 0; i < children; i++ {
		child := &yaml.Node{Kind: yaml.MappingNode}
		child.Content = append(child.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: "children"},
			buildStressGrandchildren(grandchildren),
		)
		seq.Content = append(seq.Content, child)
	}
	return seq
}

func buildStressGrandchildren(grandchildren int) *yaml.Node {
	seq := &yaml.Node{Kind: yaml.SequenceNode}
	for i := 0; i < grandchildren; i++ {
		grandchild := &yaml.Node{Kind: yaml.MappingNode}
		grandchild.Content = append(grandchild.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: "value"},
			&yaml.Node{Kind: yaml.ScalarNode, Value: fmt.Sprintf("leaf-%d", i)},
		)
		seq.Content = append(seq.Content, grandchild)
	}
	return seq
}
