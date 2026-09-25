package jsonpath

import (
	"testing"

	"github.com/pb33f/go-yaml"
)

func TestSpectralTruthinessAllValueKinds(t *testing.T) {
	source := `
values:
  - null
  - false
  - true
  - 0
  - 1
  - ""
  - x
  - []
  - {}
`
	nodes := querySpectral(t, `$.values[?(@)]`, source)
	if len(nodes) != 5 {
		t.Fatalf("truthiness selected %d nodes, want true, one, string, array and object", len(nodes))
	}
}

func TestSpectralSyntheticConstructorNames(t *testing.T) {
	source := `
values:
  string: value
  number: 42
  float: 4.2
  boolean: true
  array: []
  object: {}
  null: null
`
	tests := []struct {
		name string
		want int
	}{
		{"String", 1},
		{"Number", 2},
		{"Boolean", 1},
		{"Array", 1},
		{"Object", 1},
	}
	for _, test := range tests {
		expression := `$.values[?(@.constructor.name === '` + test.name + `')]`
		if got := querySpectral(t, expression, source); len(got) != test.want {
			t.Errorf("%s selected %d nodes, want %d", expression, len(got), test.want)
		}
	}
	if got := querySpectral(t, `$.values[?(!@.constructor.name)]`, source); len(got) != 0 {
		t.Fatalf("negated constructor type error selected %d nodes", len(got))
	}
}

func TestSpectralStringOperationsMatchJavaScriptIndexes(t *testing.T) {
	source := `
values:
  - a😀b
  - absent
`
	tests := []string{
		`$.values[?(@.indexOf('b', -3) === 3)]`,
		`$.values[?(@.indexOf('', 99) === 4)]`,
		`$.values[?(@.indexOf('missing') === -1)]`,
		`$.values[?(@.includes('😀', -1))]`,
		`$.values[?(@.length === 4)]`,
	}
	wants := []int{1, 1, 2, 1, 1}
	for i, expression := range tests {
		if got := querySpectral(t, expression, source); len(got) != wants[i] {
			t.Errorf("%s selected %d nodes, want %d", expression, len(got), wants[i])
		}
	}
}

func TestSpectralSequenceIncludesAndOffsets(t *testing.T) {
	source := `
values:
  - list: [one, two, three]
  - list: [three]
`
	if got := querySpectral(t, `$.values[?(@.list.includes('three', 2))]`, source); len(got) != 1 {
		t.Fatalf("offset includes selected %d nodes, want one", len(got))
	}
	if got := querySpectral(t, `$.values[?(@.list.includes('three', 99))]`, source); len(got) != 0 {
		t.Fatalf("oversized includes selected %d nodes, want none", len(got))
	}
	if got := querySpectral(t, `$.values[?(@.list.includes('one', -1))]`, source); len(got) != 0 {
		t.Fatalf("negative includes offset selected %d nodes before the normalized start", len(got))
	}
	if got := querySpectral(t, `$.values[?(@.list.includes('two', -2))]`, source); len(got) != 1 {
		t.Fatalf("negative includes offset selected %d nodes, want one", len(got))
	}
}

func TestSpectralLooseEqualityAndRelationalCoercion(t *testing.T) {
	source := `
values:
  - {a: "200", b: 200}
  - {a: false, b: 0}
  - {a: true, b: 1}
  - {a: abc, b: 0}
`
	if got := querySpectral(t, `$.values[?(@.a == @.b)]`, source); len(got) != 3 {
		t.Fatalf("loose equality selected %d nodes, want three", len(got))
	}
	if got := querySpectral(t, `$.values[?(@.a != @.b)]`, source); len(got) != 1 {
		t.Fatalf("loose inequality selected %d nodes, want one", len(got))
	}
	if got := querySpectral(t, `$.values[?(@.a > 100)]`, source); len(got) != 1 {
		t.Fatalf("numeric coercion selected %d nodes, want one", len(got))
	}
	if got := querySpectral(t, `$.values[?(@.missing == null)]`, source); len(got) != 4 {
		t.Fatalf("undefined/null equality selected %d nodes, want four", len(got))
	}
}

func TestSpectralRegexModes(t *testing.T) {
	source := "values: [\"foo\\nbar\", FoO, x]\n"
	tests := []struct {
		expression string
		want       int
	}{
		{`$.values[?(@.match(/^bar$/m))]`, 1},
		{`$.values[?(@.match(/foo.bar/s))]`, 1},
		{`$.values[?(@.match(/foo/ig))]`, 2},
		{`$.values[?(@.match(/foo/iu))]`, 2},
	}
	for _, test := range tests {
		if got := querySpectral(t, test.expression, source); len(got) != test.want {
			t.Errorf("%s selected %d nodes, want %d", test.expression, len(got), test.want)
		}
	}
}

func TestSpectralRootContextPathAndBracketQueries(t *testing.T) {
	source := `
expected: final
values:
  - {items: [first, final]}
  - {items: [other]}
`
	if got := querySpectral(t, `$.values[?(@.items[-1] === @root.expected)]`, source); len(got) != 1 {
		t.Fatalf("root and bracket query selected %d nodes, want one", len(got))
	}
	if got := querySpectral(t, `$.values[?(@path.match(/\[0\]$/))]`, source); len(got) != 1 {
		t.Fatalf("path method selected %d nodes, want one", len(got))
	}
}

func TestSpectralWrongTypePostfixNeverMatches(t *testing.T) {
	source := `
values:
  - {value: 1}
  - {value: {nested: true}}
  - {value: null}
`
	expressions := []string{
		`$.values[?(@.value.indexOf('x') === void 0)]`,
		`$.values[?(!@.value.length)]`,
		`$.values[?(@.value.includes('x'))]`,
	}
	for _, expression := range expressions {
		if got := querySpectral(t, expression, source); len(got) != 0 {
			t.Errorf("%s selected %d nodes after a type error", expression, len(got))
		}
	}
}

func TestSpectralScalarStrictEquality(t *testing.T) {
	source := `
values:
  - {a: 1, b: 1.0}
  - {a: true, b: true}
  - {a: null, b: null}
  - {a: one, b: one}
  - {a: one, b: two}
`
	if got := querySpectral(t, `$.values[?(@.a === @.b)]`, source); len(got) != 4 {
		t.Fatalf("strict equality selected %d nodes, want four", len(got))
	}
	if got := querySpectral(t, `$.values[?(@.a !== @.b)]`, source); len(got) != 1 {
		t.Fatalf("strict inequality selected %d nodes, want one", len(got))
	}
}

func TestSpectralLengthOnMappingIsInvalid(t *testing.T) {
	source := "values: [{a: 1}, [one], text]\n"
	nodes := querySpectral(t, `$.values[?(@.length)]`, source)
	if len(nodes) != 2 || nodes[0].Kind != yaml.SequenceNode || nodes[1].Value != "text" {
		t.Fatalf("length truthiness selected unexpected nodes: %+v", nodeValues(nodes))
	}
}
