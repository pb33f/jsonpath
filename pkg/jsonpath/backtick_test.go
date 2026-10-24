package jsonpath

import (
	"testing"

	"github.com/pb33f/go-yaml"
	"github.com/pb33f/jsonpath/pkg/jsonpath/config"
	"github.com/pb33f/testify/require"
)

func TestBacktickMemberSelection(t *testing.T) {
	var root yaml.Node
	require.NoError(t, yaml.Unmarshal([]byte("{\"$ref\":\"target\",\"$ref`\":\"trailing\",\"*\":\"star\",\"\":\"empty\",\"`*\":\"literal\",\"application/json\":{\"$ref\":\"nested\"},\"@number()\":\"type\",\"true\":\"boolean-name\",\"123\":\"numeric-name\",\"日本語\":\"unicode\",\"a-b\":\"hyphen\"}"), &root))
	for _, tc := range []struct {
		name, path, normalized string
		values                 []string
	}{
		{"ref", "$.`$ref", "$['$ref']", []string{"target"}},
		{"trailing backtick is literal", "$.`$ref`", "$['$ref`']", []string{"trailing"}},
		{"wildcard is literal", "$.`*", "$['*']", []string{"star"}},
		{"empty key", "$.`", "$['']", []string{"empty"}},
		{"escaped backtick", "$.``*", "$['`*']", []string{"literal"}},
		{"nested media type", "$.`application/json.`$ref", "$['application/json']['$ref']", []string{"nested"}},
		{"recursive", "$..`$ref", "$..['$ref']", []string{"target", "nested"}},
		{"bracket union", "$[`$ref,`*]", "$['$ref', '*']", []string{"target", "star"}},
		{"mixed union", "$[`$ref,'*']", "$['$ref', '*']", []string{"target", "star"}},
		{"type selector is literal", "$.`@number()", "$['@number()']", []string{"type"}},
		{"boolean name", "$.`true", "$['true']", []string{"boolean-name"}},
		{"numeric name", "$.`123", "$['123']", []string{"numeric-name"}},
		{"unicode", "$.`日本語", "$['日本語']", []string{"unicode"}},
		{"hyphen", "$.`a-b", "$['a-b']", []string{"hyphen"}},
		{"missing", "$.`absent", "$['absent']", nil},
		{"quoted backtick stays literal", "$['`*']", "$['`*']", []string{"literal"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, options := range [][]config.Option{nil, {config.WithSpectralCompatibility(), config.WithLazyContextTracking()}} {
				path, err := NewPath(tc.path, options...)
				require.NoError(t, err)
				require.Equal(t, tc.normalized, path.String())
				var values []string
				for _, node := range path.Query(&root) {
					values = append(values, node.Value)
				}
				require.Equal(t, tc.values, values)
				normalized, err := NewPath(path.String(), config.WithStrictRFC9535())
				require.NoError(t, err)
				require.Equal(t, path.Query(&root), normalized.Query(&root))
			}
		})
	}
}

func TestBacktickStrictAndInvalidContexts(t *testing.T) {
	for _, expression := range []string{"$.`$ref", "$..`*", "$[`$ref]"} {
		_, err := NewPath(expression, config.WithStrictRFC9535())
		require.ErrorContains(t, err, "backtick member escapes require JSONPath Plus mode")
	}
	for _, expression := range []string{"`$ref", "$`$ref", "$.name`other", "$[?@.`$ref]", "$[`$ref", "$.`name]"} {
		t.Run(expression, func(t *testing.T) {
			_, err := NewPath(expression)
			require.Error(t, err)
		})
	}
	// Quoted keys remain RFC-compatible even when the key contains a backtick.
	_, err := NewPath("$['`*']", config.WithStrictRFC9535())
	require.NoError(t, err)
}

func TestBacktickNavigationAndMetadata(t *testing.T) {
	var root yaml.Node
	require.NoError(t, yaml.Unmarshal([]byte(`{"$ref":[{"name":"first"},{"name":"second"}]}`), &root))
	path, err := NewPath("$.`$ref[1].name")
	require.NoError(t, err)
	require.True(t, path.IsSingular())
	segments, err := path.GetSegmentInfo()
	require.NoError(t, err)
	require.Len(t, segments, 3)
	require.Equal(t, "$ref", segments[0].Key)
	require.Equal(t, int64(1), segments[1].Index)
	require.Equal(t, "name", segments[2].Key)
	nodes := path.Query(&root)
	require.Len(t, nodes, 1)
	require.Equal(t, "second", nodes[0].Value)

	path, err = NewPath("$.`$ref~", config.WithPropertyNameExtension())
	require.NoError(t, err)
	nodes = path.Query(&root)
	require.Len(t, nodes, 1)
	require.Equal(t, "$ref", nodes[0].Value)

	path, err = NewPath("$.`$ref^")
	require.NoError(t, err)
	nodes = path.Query(&root)
	require.Len(t, nodes, 1)
	require.Equal(t, root.Content[0], nodes[0])
}

func BenchmarkBacktickCompile(b *testing.B) {
	for _, expression := range []string{"$.paths.*.*.requestBody.content['application/json'].schema['$ref']", "$.paths.*.*.requestBody.content.`application/json.schema.`$ref"} {
		b.Run(expression, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				if _, err := NewPath(expression); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
