package jsonpath

import (
	"testing"

	"go.yaml.in/yaml/v4"
)

func TestContextUsageAndPropertyReferenceDetection(t *testing.T) {
	propSeg := &segment{kind: segmentKindProperyName}

	// nil-safe pointer wrapper
	var nilAST *jsonPathAST
	if nilAST.hasPropertyNameReferencesPtr() {
		t.Fatalf("expected nil AST wrapper to report false")
	}
	nilAST.collectContextVarUsage(&contextVarUsage{})

	// empty and positive AST checks
	if (jsonPathAST{}).hasPropertyNameReferences() {
		t.Fatalf("expected empty AST to report no property-name references")
	}
	if !(jsonPathAST{segments: []*segment{propSeg}}).hasPropertyNameReferences() {
		t.Fatalf("expected AST with property-name segment to report true")
	}

	// negative branches across all helper types
	if (&segment{}).hasPropertyNameReferences() {
		t.Fatalf("expected empty segment to report false")
	}
	if (&innerSegment{}).hasPropertyNameReferences() {
		t.Fatalf("expected empty inner segment to report false")
	}
	if (&selector{}).hasPropertyNameReferences() {
		t.Fatalf("expected selector without filter to report false")
	}
	if (&filterSelector{}).hasPropertyNameReferences() {
		t.Fatalf("expected filter selector without expression to report false")
	}
	if (&logicalOrExpr{}).hasPropertyNameReferences() {
		t.Fatalf("expected empty logicalOrExpr to report false")
	}
	if (&logicalAndExpr{}).hasPropertyNameReferences() {
		t.Fatalf("expected empty logicalAndExpr to report false")
	}
	if (&basicExpr{}).hasPropertyNameReferences() {
		t.Fatalf("expected empty basicExpr to report false")
	}
	if (&comparisonExpr{}).hasPropertyNameReferences() {
		t.Fatalf("expected empty comparisonExpr to report false")
	}
	if (&comparable{}).hasPropertyNameReferences() {
		t.Fatalf("expected empty comparable to report false")
	}
	if (&testExpr{}).hasPropertyNameReferences() {
		t.Fatalf("expected empty testExpr to report false")
	}
	if (&functionExpr{}).hasPropertyNameReferences() {
		t.Fatalf("expected empty functionExpr to report false")
	}
	if (&functionArgument{}).hasPropertyNameReferences() {
		t.Fatalf("expected empty functionArgument to report false")
	}
	if (&filterQuery{}).hasPropertyNameReferences() {
		t.Fatalf("expected empty filterQuery to report false")
	}
	if (&relQuery{}).hasPropertyNameReferences() {
		t.Fatalf("expected empty relQuery to report false")
	}
	if (&absQuery{}).hasPropertyNameReferences() {
		t.Fatalf("expected empty absQuery to report false")
	}
	if (&singularQuery{}).hasPropertyNameReferences() {
		t.Fatalf("expected empty singularQuery to report false")
	}
	var nilFilterQuery *filterQuery
	if nilFilterQuery.hasPropertyNameReferences() {
		t.Fatalf("expected nil filterQuery to report false")
	}

	// positive branches across each helper type
	truthySelector := &selector{filter: &filterSelector{expression: &logicalOrExpr{
		expressions: []*logicalAndExpr{
			{
				expressions: []*basicExpr{
					{testExpr: &testExpr{filterQuery: &filterQuery{relQuery: &relQuery{segments: []*segment{propSeg}}}}},
				},
			},
		},
	}}}
	if !(&innerSegment{selectors: []*selector{truthySelector}}).hasPropertyNameReferences() {
		t.Fatalf("expected inner segment to report true")
	}
	if !(&segment{child: &innerSegment{selectors: []*selector{truthySelector}}}).hasPropertyNameReferences() {
		t.Fatalf("expected segment(child) to report true")
	}
	if !(&segment{descendant: &innerSegment{selectors: []*selector{truthySelector}}}).hasPropertyNameReferences() {
		t.Fatalf("expected segment(descendant) to report true")
	}

	comp := &comparable{singularQuery: &singularQuery{relQuery: &relQuery{segments: []*segment{propSeg}}}}
	if !(&comparisonExpr{left: comp}).hasPropertyNameReferences() {
		t.Fatalf("expected comparisonExpr(left) to report true")
	}
	if !(&comparisonExpr{right: comp}).hasPropertyNameReferences() {
		t.Fatalf("expected comparisonExpr(right) to report true")
	}
	if !(&comparable{functionExpr: &functionExpr{args: []*functionArgument{{filterQuery: &filterQuery{relQuery: &relQuery{segments: []*segment{propSeg}}}}}}}).hasPropertyNameReferences() {
		t.Fatalf("expected comparable(functionExpr) to report true")
	}
	if !(&testExpr{functionExpr: &functionExpr{args: []*functionArgument{{filterQuery: &filterQuery{relQuery: &relQuery{segments: []*segment{propSeg}}}}}}}).hasPropertyNameReferences() {
		t.Fatalf("expected testExpr(functionExpr) to report true")
	}
	logicalArg := &functionArgument{
		logicalExpr: &logicalOrExpr{
			expressions: []*logicalAndExpr{
				{
					expressions: []*basicExpr{
						{
							testExpr: &testExpr{
								filterQuery: &filterQuery{
									relQuery: &relQuery{segments: []*segment{propSeg}},
								},
							},
						},
					},
				},
			},
		},
	}
	if !logicalArg.hasPropertyNameReferences() {
		t.Fatalf("expected functionArgument(logicalExpr) to report true")
	}
	if !(&functionArgument{functionExpr: &functionExpr{args: []*functionArgument{{filterQuery: &filterQuery{relQuery: &relQuery{segments: []*segment{propSeg}}}}}}}).hasPropertyNameReferences() {
		t.Fatalf("expected functionArgument(functionExpr) to report true")
	}
	if !(&filterQuery{jsonPathQuery: &jsonPathAST{segments: []*segment{propSeg}}}).hasPropertyNameReferences() {
		t.Fatalf("expected filterQuery(jsonPathQuery) to report true")
	}
	if !(&absQuery{segments: []*segment{propSeg}}).hasPropertyNameReferences() {
		t.Fatalf("expected absQuery to report true")
	}
	if !(&singularQuery{absQuery: &absQuery{segments: []*segment{propSeg}}}).hasPropertyNameReferences() {
		t.Fatalf("expected singularQuery(absQuery) to report true")
	}
}

func TestCollectContextVarUsageCoversBranches(t *testing.T) {
	var usage contextVarUsage

	// mark all context variable kinds
	for _, kind := range []contextVarKind{
		contextVarProperty,
		contextVarParent,
		contextVarParentProperty,
		contextVarPath,
		contextVarIndex,
	} {
		usage.mark(kind)
	}
	if !usage.property || !usage.parent || !usage.parentProperty || !usage.path || !usage.index {
		t.Fatalf("expected all context variable flags to be marked")
	}

	// exercise traversal methods including nil guards and optional children
	var nilAST *jsonPathAST
	nilAST.collectContextVarUsage(&usage)
	var nilFilterQuery *filterQuery
	nilFilterQuery.collectContextVarUsage(&usage)

	leftComp := &comparable{contextVar: &contextVariable{kind: contextVarProperty}}
	rightComp := &comparable{functionExpr: &functionExpr{
		args: []*functionArgument{{contextVar: &contextVariable{kind: contextVarIndex}}},
	}}
	firstBasic := &basicExpr{
		parenExpr: &parenExpr{
			expr: &logicalOrExpr{
				expressions: []*logicalAndExpr{
					{expressions: []*basicExpr{{comparisonExpr: &comparisonExpr{left: leftComp, right: rightComp}}}},
				},
			},
		},
	}
	childSelector := &selector{
		filter: &filterSelector{expression: &logicalOrExpr{
			expressions: []*logicalAndExpr{{expressions: []*basicExpr{firstBasic}}},
		}},
	}

	pathArg := &functionArgument{
		logicalExpr: &logicalOrExpr{
			expressions: []*logicalAndExpr{{expressions: []*basicExpr{
				{testExpr: &testExpr{functionExpr: &functionExpr{args: []*functionArgument{
					{contextVar: &contextVariable{kind: contextVarPath}},
				}}}},
			}}},
		},
	}
	filterArg := &functionArgument{
		functionExpr: &functionExpr{
			args: []*functionArgument{{
				filterQuery: &filterQuery{relQuery: &relQuery{segments: []*segment{{kind: segmentKindChild}}}},
			}},
		},
	}
	jsonPathArg := &functionArgument{
		filterQuery: &filterQuery{jsonPathQuery: &jsonPathAST{segments: []*segment{{kind: segmentKindChild}}}},
	}
	secondBasic := &basicExpr{
		testExpr: &testExpr{
			filterQuery: &filterQuery{
				relQuery:      &relQuery{segments: []*segment{{kind: segmentKindChild}}},
				jsonPathQuery: &jsonPathAST{segments: []*segment{{kind: segmentKindChild}}},
			},
			functionExpr: &functionExpr{args: []*functionArgument{
				{contextVar: &contextVariable{kind: contextVarParentProperty}},
				pathArg,
				filterArg,
				jsonPathArg,
			}},
		},
	}
	descSelector := &selector{
		filter: &filterSelector{expression: &logicalOrExpr{
			expressions: []*logicalAndExpr{{expressions: []*basicExpr{secondBasic}}},
		}},
	}

	tree := &jsonPathAST{segments: []*segment{
		{child: &innerSegment{selectors: []*selector{childSelector}}},
		{descendant: &innerSegment{selectors: []*selector{descSelector}}},
	}}
	tree.collectContextVarUsage(&usage)

	// direct calls to cover rel/abs/singular and empty paths
	(&relQuery{}).collectContextVarUsage(&usage)
	(&absQuery{}).collectContextVarUsage(&usage)
	(&singularQuery{}).collectContextVarUsage(&usage)
	(&singularQuery{relQuery: &relQuery{segments: []*segment{{kind: segmentKindChild}}}}).collectContextVarUsage(&usage)
	(&singularQuery{absQuery: &absQuery{segments: []*segment{{kind: segmentKindChild}}}}).collectContextVarUsage(&usage)

	if !usage.property || !usage.parentProperty || !usage.path || !usage.index {
		t.Fatalf("expected collect traversal to retain marked usage flags")
	}
}

func TestDescendApplyAndDescendantSegmentQuery(t *testing.T) {
	calls := 0
	descendApply(nil, func(*yaml.Node) {
		calls++
	})
	if calls != 0 {
		t.Fatalf("expected no callback calls for nil root")
	}

	root := &yaml.Node{Kind: yaml.MappingNode, Content: []*yaml.Node{
		{Kind: yaml.ScalarNode, Value: "a"},
		{Kind: yaml.SequenceNode, Content: []*yaml.Node{
			{Kind: yaml.ScalarNode, Value: "x"},
		}},
	}}
	descendApply(root, func(*yaml.Node) {
		calls++
	})
	if calls < 3 {
		t.Fatalf("expected descendApply to visit full tree, got %d", calls)
	}

	idx := &_index{propertyKeys: map[*yaml.Node]*yaml.Node{}, parentNodes: map[*yaml.Node]*yaml.Node{}}
	q := segment{
		kind: segmentKindDescendant,
		descendant: &innerSegment{
			kind: segmentDotWildcard,
		},
	}
	result := q.Query(idx, root, root)
	if len(result) == 0 {
		t.Fatalf("expected descendant query to return at least one node")
	}
}

type bareFilterContext struct {
	_index
}

func (b *bareFilterContext) PropertyName() string                                 { return "" }
func (b *bareFilterContext) SetPropertyName(string)                               {}
func (b *bareFilterContext) Parent() *yaml.Node                                   { return nil }
func (b *bareFilterContext) SetParent(*yaml.Node)                                 {}
func (b *bareFilterContext) ParentPropertyName() string                           { return "" }
func (b *bareFilterContext) SetParentPropertyName(string)                         {}
func (b *bareFilterContext) Path() string                                         { return "$" }
func (b *bareFilterContext) PushPathSegment(string)                               {}
func (b *bareFilterContext) PopPathSegment()                                      {}
func (b *bareFilterContext) SetPendingPathSegment(*yaml.Node, string)             {}
func (b *bareFilterContext) GetAndClearPendingPathSegment(*yaml.Node) string      { return "" }
func (b *bareFilterContext) SetPendingPropertyName(*yaml.Node, string)            {}
func (b *bareFilterContext) GetAndClearPendingPropertyName(*yaml.Node) string     { return "" }
func (b *bareFilterContext) Root() *yaml.Node                                     { return nil }
func (b *bareFilterContext) SetRoot(*yaml.Node)                                   {}
func (b *bareFilterContext) Index() int                                           { return -1 }
func (b *bareFilterContext) SetIndex(int)                                         {}
func (b *bareFilterContext) EnableParentTracking()                                {}
func (b *bareFilterContext) ParentTrackingEnabled() bool                          { return false }
func (b *bareFilterContext) Clone() FilterContext                                 { return b }

func TestEnableTrackingHelpersNoOpForMissingOptionalMethods(t *testing.T) {
	ctx := &bareFilterContext{_index: _index{
		propertyKeys: make(map[*yaml.Node]*yaml.Node),
		parentNodes:  make(map[*yaml.Node]*yaml.Node),
	}}
	enablePropertyTracking(ctx)
	enablePathTracking(ctx)
	enableIndexTracking(ctx)
}
