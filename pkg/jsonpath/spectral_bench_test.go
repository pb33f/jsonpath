package jsonpath

import (
	"fmt"
	"strings"
	"testing"

	"github.com/pb33f/jsonpath/pkg/jsonpath/config"
	"go.yaml.in/yaml/v4"
)

func BenchmarkSpectralRegexFilter(b *testing.B) {
	var source strings.Builder
	source.WriteString("paths:\n")
	for i := 0; i < 1000; i++ {
		fmt.Fprintf(&source, "  /resource/%d:\n    responses:\n      %d: {}\n", i, 200+i%400)
	}
	var document yaml.Node
	if err := yaml.Unmarshal([]byte(source.String()), &document); err != nil {
		b.Fatal(err)
	}
	path, err := NewPath(`$.paths[*].responses[?(@property.match(/^(4|5)/))]`, config.WithSpectralCompatibility())
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = path.Query(&document)
	}
}

func BenchmarkSpectralRegexCompilation(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := NewPath(`$.paths[?(@property.match(/^\/api\/v[0-9]+/i))]`, config.WithSpectralCompatibility()); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDefaultRFCFilter(b *testing.B) {
	var document yaml.Node
	if err := yaml.Unmarshal([]byte("values: [{count: 1}, {count: 2}, {count: 3}]\n"), &document); err != nil {
		b.Fatal(err)
	}
	path, err := NewPath(`$.values[?(@.count > 1)]`)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = path.Query(&document)
	}
}

func BenchmarkDefaultRFCCompilation(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := NewPath(`$.values[?(@.count > 1)]`); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSpectralContextTracking(b *testing.B) {
	var document yaml.Node
	if err := yaml.Unmarshal([]byte("values: [{name: one}, {name: two}, {name: three}]\n"), &document); err != nil {
		b.Fatal(err)
	}
	tests := []struct {
		name       string
		expression string
		options    []config.Option
	}{
		{name: "DefaultEager", expression: `$.values[?(@.name == 'three')]`},
		{name: "DefaultLazy", expression: `$.values[?(@.name == 'three')]`, options: []config.Option{config.WithLazyContextTracking()}},
		{name: "SpectralEager", expression: `$.values[?(@property === 2 && @.name.length === 5)]`, options: []config.Option{config.WithSpectralCompatibility()}},
		{name: "SpectralLazy", expression: `$.values[?(@property === 2 && @.name.length === 5)]`, options: []config.Option{config.WithSpectralCompatibility(), config.WithLazyContextTracking()}},
	}
	for _, test := range tests {
		b.Run(test.name, func(b *testing.B) {
			path, err := NewPath(test.expression, test.options...)
			if err != nil {
				b.Fatal(err)
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = path.Query(&document)
			}
		})
	}
}
