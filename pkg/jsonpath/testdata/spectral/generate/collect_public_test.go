package main

import (
	"reflect"
	"sort"
	"testing"

	"go.yaml.in/yaml/v4"
)

func TestCollectPublicSelectorsExpandsAliasesAndLists(t *testing.T) {
	source := `
aliases:
  ResponsesObject:
    targets:
      - given: $..responses
  HttpStatus:
    - "#ResponsesObject.*~"
rules:
  aliased:
    given: "#HttpStatus"
  listed:
    given:
      - $.paths
      - $.channels
`
	var document yaml.Node
	if err := yaml.Unmarshal([]byte(source), &document); err != nil {
		t.Fatal(err)
	}
	aliases := collectAliasDefinitions(&document)
	var collected []string
	collectGivenSelectors(&document, &collected)
	var expanded []string
	for _, expression := range collected {
		expanded = append(expanded, expandAliasExpression(expression, aliases, nil)...)
	}
	expanded = uniqueStrings(expanded)
	sort.Strings(expanded)
	want := []string{"$..responses", "$..responses.*~", "$.channels", "$.paths"}
	if !reflect.DeepEqual(expanded, want) {
		t.Fatalf("expanded selectors = %v, want %v", expanded, want)
	}
}

func TestExpandPublicAliasLeavesUnknownAndCyclesClassifiable(t *testing.T) {
	aliases := map[string][]string{
		"CycleA": {"#CycleB"},
		"CycleB": {"#CycleA"},
	}
	if got := expandAliasExpression("#Missing", aliases, nil); !reflect.DeepEqual(got, []string{"#Missing"}) {
		t.Fatalf("unknown alias expansion = %v", got)
	}
	if got := expandAliasExpression("#CycleA", aliases, nil); !reflect.DeepEqual(got, []string{"#CycleA"}) {
		t.Fatalf("cyclic alias expansion = %v", got)
	}
}
