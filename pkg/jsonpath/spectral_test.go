package jsonpath

import (
	"regexp"
	"strings"
	"testing"

	"github.com/pb33f/jsonpath/pkg/jsonpath/config"
	"go.yaml.in/yaml/v4"
)

func parseSpectralTestYAML(t *testing.T, source string) *yaml.Node {
	t.Helper()
	var document yaml.Node
	if err := yaml.Unmarshal([]byte(source), &document); err != nil {
		t.Fatalf("parse YAML: %v", err)
	}
	return &document
}

func querySpectral(t *testing.T, expression, source string, extra ...config.Option) []*yaml.Node {
	t.Helper()
	options := append([]config.Option{config.WithSpectralCompatibility()}, extra...)
	path, err := NewPath(expression, options...)
	if err != nil {
		t.Fatalf("compile %s: %v", expression, err)
	}
	return path.Query(parseSpectralTestYAML(t, source))
}

func nodeValues(nodes []*yaml.Node) []string {
	values := make([]string, len(nodes))
	for i, node := range nodes {
		values[i] = node.Value
	}
	return values
}

func TestSpectralIssue936Selector(t *testing.T) {
	source := `
paths:
  /pets:
    get: {}
  /openapi.json:
    get: {}
  /nested/openapi.json:
    get: {}
  /openapi.json/pets:
    get: {}
`
	nodes := querySpectral(t, `$.paths[?(@property && !@property.match(/\/openapi\.json/))]~`, source)
	if got, want := nodeValues(nodes), []string{"/pets"}; !equalStrings(got, want) {
		t.Fatalf("selected paths = %v, want %v", got, want)
	}
}

func TestSpectralResponseCodeMatchStringifiesMappingKeys(t *testing.T) {
	source := `
paths:
  /pets:
    responses:
      200: {description: ok}
      404: {description: missing}
      default: {description: fallback}
`
	nodes := querySpectral(t, `$.paths[*].responses[?(@property.match(/^(4|5)/))]`, source)
	if len(nodes) != 1 || nodes[0].Kind != yaml.MappingNode {
		t.Fatalf("selected %d nodes, want one response object", len(nodes))
	}
}

func TestSpectralIndexOfAndParentSelector(t *testing.T) {
	source := `
content:
  application/json: {schema: {type: object}}
  application/xml: {schema: {type: object}}
  text/json-seq: {schema: {type: object}}
`
	nodes := querySpectral(t, `$.content[?(@property.indexOf('json') === -1)]^`, source)
	if len(nodes) != 1 || nodes[0].Kind != yaml.MappingNode {
		t.Fatalf("selected %d nodes, want the content mapping once", len(nodes))
	}
}

func TestSpectralConstructorNameAndTruthiness(t *testing.T) {
	source := `
schemas:
  one: {enum: [a, b]}
  two: {enum: value}
  three: {type: string}
`
	nodes := querySpectral(t, `$..[?(@property !== 'properties' && @.enum && @.enum.constructor.name === 'Array')]`, source)
	if len(nodes) != 1 || nodes[0].Kind != yaml.MappingNode {
		t.Fatalf("selected %d nodes, want the mapping with an array enum", len(nodes))
	}
}

func TestSpectralOfficialArrayIncludesAndUndefined(t *testing.T) {
	source := `
schemas:
  list: {type: [string, array], example: []}
  scalar: {type: string}
  implicit: {type: string, default: null}
`
	nodes := querySpectral(t, `$..[?(@ && @.type && @.type.constructor.name === 'Array' && @.type.includes('array'))]`, source)
	if len(nodes) != 1 {
		t.Fatalf("array includes selected %d nodes, want one", len(nodes))
	}
	nodes = querySpectral(t, `$..[?(@ && (@.example !== void 0 || @.default !== void 0))]`, source)
	if len(nodes) != 2 {
		t.Fatalf("undefined checks selected %d nodes, want two", len(nodes))
	}
}

func TestSpectralLengthAndUTF16Indexing(t *testing.T) {
	source := `
values:
  - {text: "a😀b", list: [1, 2]}
  - {text: "ab", list: []}
`
	nodes := querySpectral(t, `$.values[?(@.text.length === 4 && @.text.indexOf('b') === 3 && @.list.length > 0)]`, source)
	if len(nodes) != 1 {
		t.Fatalf("selected %d nodes, want one UTF-16-compatible value", len(nodes))
	}
}

func TestSpectralPropertyTypingForSequenceIndexes(t *testing.T) {
	source := "values: [zero, one, two, three]\n"
	if got := nodeValues(querySpectral(t, `$.values[?(@property === 3)]`, source)); !equalStrings(got, []string{"three"}) {
		t.Fatalf("numeric property result = %v", got)
	}
	if got := querySpectral(t, `$.values[?(@property === '3')]`, source); len(got) != 0 {
		t.Fatalf("string property selected %d nodes, want none", len(got))
	}
	if got := querySpectral(t, `$.values[?(@property.match(/3/))]`, source); len(got) != 1 || got[0].Value != "three" {
		t.Fatalf("Nimma-compatible string method selected %v, want three", nodeValues(got))
	}
	if got := querySpectral(t, `$.values[?(!@property.match(/3/))]`, source); len(got) != 3 {
		t.Fatalf("negated Nimma-compatible string method selected %d nodes, want three", len(got))
	}
}

func TestSpectralParentPropertyTypingForSequenceIndexes(t *testing.T) {
	source := `
groups:
  - items: [a]
  - items: [b]
`
	nodes := querySpectral(t, `$.groups[0][?(@parentProperty === 0)]`, source)
	if len(nodes) != 1 || nodes[0].Kind != yaml.SequenceNode {
		t.Fatalf("numeric parent property selected %d nodes, want the items sequence", len(nodes))
	}
	if got := querySpectral(t, `$.groups[0][?(@parentProperty === '0')]`, source); len(got) != 0 {
		t.Fatalf("string parent property selected %d nodes, want none", len(got))
	}
	if got := querySpectral(t, `$.groups[0][?(@parentProperty.match(/^0$/))]`, source); len(got) != 1 {
		t.Fatalf("Nimma-compatible parent-property method selected %d nodes, want one", len(got))
	}
}

func TestSpectralPreservesRFCFunctionExpressions(t *testing.T) {
	source := `
values:
  - {name: alpha}
  - {name: xy}
`
	nodes := querySpectral(t, `$.values[?(length(@.name) > 3 && match(@.name, 'a.*'))]`, source)
	if len(nodes) != 1 {
		t.Fatalf("RFC functions selected %d nodes, want one", len(nodes))
	}
}

func TestSpectralRegexFlagsAndEscapes(t *testing.T) {
	source := "values: [FoO, bar/baz, abc123]\n"
	tests := []struct {
		expression string
		want       string
	}{
		{`$.values[?(@.match(/foo/i))]`, "FoO"},
		{`$.values[?(@.match(/bar\/baz/))]`, "bar/baz"},
		{`$.values[?(@.match(/[a-z]+\d+/))]`, "abc123"},
	}
	for _, test := range tests {
		if got := nodeValues(querySpectral(t, test.expression, source)); !equalStrings(got, []string{test.want}) {
			t.Errorf("%s selected %v, want %q", test.expression, got, test.want)
		}
	}
}

func TestSpectralStrictStructuralEqualityUsesIdentity(t *testing.T) {
	source := `
values:
  - {a: [1, 2], b: [1, 2]}
  - &same {a: &items [1, 2], b: *items}
`
	nodes := querySpectral(t, `$.values[?(@.a === @.b)]`, source)
	if len(nodes) > 1 {
		t.Fatalf("strict equality selected %d nodes, want at most the YAML alias identity case", len(nodes))
	}
	if got := querySpectral(t, `$.values[?(@.a == @.b)]`, source); len(got) != 0 {
		t.Fatalf("JavaScript loose reference equality selected %d distinct arrays", len(got))
	}
}

func TestSpectralDefaultAndStrictModesRemainIsolated(t *testing.T) {
	expression := `$.paths[?(@property.match(/x/))]~`
	if _, err := NewPath(expression); err == nil {
		t.Fatal("default mode unexpectedly accepted Spectral method syntax")
	}
	if _, err := NewPath(expression, config.WithStrictRFC9535()); err == nil {
		t.Fatal("strict mode unexpectedly accepted Spectral method syntax")
	}
	orders := [][]config.Option{
		{config.WithStrictRFC9535(), config.WithSpectralCompatibility()},
		{config.WithSpectralCompatibility(), config.WithStrictRFC9535()},
	}
	for _, options := range orders {
		if _, err := NewPath(`$`, options...); err == nil || !strings.Contains(err.Error(), "cannot be combined") {
			t.Fatalf("conflicting options error = %v", err)
		}
	}

	document := parseSpectralTestYAML(t, "- [one]\n")
	defaultPath, err := NewPath(`$[0][?(@parentProperty == '0')]`)
	if err != nil {
		t.Fatalf("compile default context query: %v", err)
	}
	if got := defaultPath.Query(document); len(got) != 0 {
		t.Fatalf("default context behavior changed: selected %d nodes", len(got))
	}
	if got := querySpectral(t, `$[0][?(@parentProperty === 0)]`, "- [one]\n"); len(got) != 1 {
		t.Fatalf("Spectral parent property typing selected %d nodes, want one", len(got))
	}
}

func TestSpectralRejectsUnsafeAndUnsupportedSyntax(t *testing.T) {
	tests := []struct {
		expression string
		contains   string
	}{
		{`$[?(@.constructor.constructor('return 1')())]`, "constructor.name"},
		{`$[?(@.__proto__)]`, "prototype"},
		{`$[?(@.__defineGetter__)]`, "prototype"},
		{`$[?(@.__defineSetter__)]`, "prototype"},
		{`$[?(@['constructor'])]`, "computed constructor"},
		{`$[?(@property.startsWith('x'))]`, "unsupported method"},
		{`$[?(@.value = 1)]`, "assignments are not supported"},
		{`$[?(@.value; true)]`, "statement separators are not supported"},
		{`$[?(function() {})]`, "function declarations"},
		{`$[?(@property.match(/(a)\1/))]`, "backreferences"},
		{`$[?(@property.match(/a(?=b)/))]`, "lookaround"},
		{`$[?(@property.match(/x/dd))]`, "duplicate regex flag"},
		{`$[?(@property.match(/x/y))]`, "regex flag"},
	}
	for _, test := range tests {
		_, err := NewPath(test.expression, config.WithSpectralCompatibility())
		if err == nil || !strings.Contains(err.Error(), test.contains) || !strings.Contains(err.Error(), "line 1, column") {
			t.Errorf("%s error = %v, want substring %q", test.expression, err, test.contains)
		}
	}
}

func TestSpectralRejectsMalformedPostfixAndRegexForms(t *testing.T) {
	expressions := []string{
		`$[?(@property.match('x'))]`,
		`$[?(@property.match(/x/, /y/))]`,
		`$[?(@property.match(/x/) === true)]`,
		`$[?(@property.match(/x/).length)]`,
		`$[?(@property.indexOf())]`,
		`$[?(@property.indexOf('x', '1'))]`,
		`$[?(@property.includes())]`,
		`$[?(@property.includes('x', '1'))]`,
		`$[?(@property['length'])]`,
		`$[?(@.constructor)]`,
		`$[?(@.constructor.name.value)]`,
		`$[?(@.unknown())]`,
		`$[?(@property.match(/x/u).length)]`,
		`$[?(@property.match(/\u{1F600}/u))]`,
		`$[?(@property.match(/x/v))]`,
		`$[?(@property.match(/x/z))]`,
		`$[?(@property.match(/[a-/))]`,
	}
	for _, expression := range expressions {
		if _, err := NewPath(expression, config.WithSpectralCompatibility()); err == nil {
			t.Errorf("expected compilation failure for %s", expression)
		}
	}
}

func TestSpectralRegexIsCompiledIntoReusableAST(t *testing.T) {
	path, err := NewPath(`$.values[?(@property.match(/^x/i))]`, config.WithSpectralCompatibility())
	if err != nil {
		t.Fatal(err)
	}
	compiled := spectralCompiledRegexes(path)
	if len(compiled) != 1 || compiled[0] == nil {
		t.Fatalf("compiled regexes = %v", compiled)
	}
	document := parseSpectralTestYAML(t, "values: {x: true, y: false}\n")
	_ = path.Query(document)
	_ = path.Query(document)
	after := spectralCompiledRegexes(path)
	if len(after) != 1 || after[0] != compiled[0] {
		t.Fatal("query replaced or recompiled the cached regex")
	}
}

func TestSpectralLazyContextTrackingMatchesEager(t *testing.T) {
	expression := `$.values[?(@property === 1 || @.name.length > 3)]`
	source := "values: [{name: one}, {name: two}, {name: three}]\n"
	eager := querySpectral(t, expression, source)
	lazy := querySpectral(t, expression, source, config.WithLazyContextTracking())
	if got, want := nodeValues(lazy), nodeValues(eager); !equalStrings(got, want) {
		t.Fatalf("lazy results = %v, eager = %v", got, want)
	}
}

func TestSpectralPinnedOfficialFilterSelectorsCompile(t *testing.T) {
	expressions := []string{
		`$..[?(@ && @.type=="array")]`,
		`$..[?(@ && @.type && @.type.constructor.name === "Array" && @.type.includes("array"))]`,
		`$..[?(@property !== 'properties' && @.enum && @.enum.constructor.name === 'Array')]`,
		`$..[?(@property === '$ref')]`,
		`$..[?(@ && @.enum && @.type)]`,
		`$.definitions[?(@.discriminator)]`,
		`$..parameters[?(@ && @.in)]`,
		`$..definitions..[?(@property !== 'properties' && @ && (@.example !== void 0 || @['x-example'] !== void 0 || @.default !== void 0) && (@.enum || @.type || @.format || @.$ref || @.properties || @.items))]`,
		`$..responses..[?(@ && @.schema && @.examples)]`,
		`$.components.parameters[?(@ && @.in)]`,
		`$..content..[?(@ && @.schema && (@.example !== void 0 || @.examples))]`,
		`$.components.schemas..[?(@property !== 'properties' && @ && (@ && @.example !== void 0 || @.default !== void 0) && (@.enum || @.type || @.format || @.$ref || @.properties || @.items))]`,
		`$.components.messageTraits[?(@.schemaFormat === void 0)].payload.default^`,
		`$.channels[*][publish,subscribe][?(@property === 'message' && @.schemaFormat === void 0)].payload`,
	}
	for _, expression := range expressions {
		path, err := NewPath(expression, config.WithSpectralCompatibility())
		if err != nil {
			t.Errorf("compile %s: %v", expression, err)
			continue
		}
		if path.String() == "" {
			t.Errorf("empty round trip for %s", expression)
		}
	}
}

func TestSpectralRoundTripPreservesBracketMemberSyntax(t *testing.T) {
	expression := `$[?(@['x-example'] && @['content-type'] && @.$ref)]`
	path, err := NewPath(expression, config.WithSpectralCompatibility())
	if err != nil {
		t.Fatalf("compile %s: %v", expression, err)
	}

	rendered := path.String()
	if rendered != expression {
		t.Fatalf("round trip = %q, want %q", rendered, expression)
	}
	_, err = NewPath(rendered, config.WithSpectralCompatibility())
	if err != nil {
		t.Fatalf("compile round trip %s: %v", rendered, err)
	}
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func spectralCompiledRegexes(path *JSONPath) []*regexp.Regexp {
	var result []*regexp.Regexp
	var walkExpr func(*spectralBoolExpr)
	walkExpr = func(expression *spectralBoolExpr) {
		if expression == nil {
			return
		}
		if expression.value != nil {
			if expression.value.regex != nil {
				result = append(result, expression.value.regex.compiled)
			}
			for _, postfix := range expression.value.postfix {
				if postfix.regex != nil {
					result = append(result, postfix.regex.compiled)
				}
			}
		}
		walkExpr(expression.left)
		walkExpr(expression.right)
	}
	for _, segment := range path.ast.segments {
		var inner *innerSegment
		if segment.child != nil {
			inner = segment.child
		} else {
			inner = segment.descendant
		}
		if inner == nil {
			continue
		}
		for _, selector := range inner.selectors {
			if selector.filter.present() {
				walkExpr(selector.filter.spectralExpression)
			}
		}
	}
	return result
}
