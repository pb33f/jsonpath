package jsonpath

import (
	"testing"

	"github.com/pb33f/jsonpath/pkg/jsonpath/config"
	"github.com/pb33f/jsonpath/pkg/jsonpath/token"
)

func TestParseTestExprPropagatesLazyContextTrackingToNestedJSONPathQuery(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		opts []config.Option
		want bool
	}{
		{
			name: "eager by default",
			want: false,
		},
		{
			name: "lazy with option",
			opts: []config.Option{config.WithLazyContextTracking()},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tokenizer := token.NewTokenizer("$[?$.x]", tt.opts...)
			parser := newParserPrivate(tokenizer, tokenizer.Tokenize(), tt.opts...)
			if err := parser.parse(); err != nil {
				t.Fatalf("parse failed: %v", err)
			}

			nested := nestedJSONPathFromTestExpr(parser.ast)
			if nested == nil {
				t.Fatalf("expected nested jsonpath query in test expression")
			}
			if nested.lazyContextTracking != tt.want {
				t.Fatalf("expected nested lazyContextTracking=%v, got %v", tt.want, nested.lazyContextTracking)
			}
		})
	}
}

func TestParseFunctionArgumentPropagatesLazyContextTrackingToNestedJSONPathQuery(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		opts []config.Option
		want bool
	}{
		{
			name: "eager by default",
			want: false,
		},
		{
			name: "lazy with option",
			opts: []config.Option{config.WithLazyContextTracking()},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tokenizer := token.NewTokenizer("$[?length($.x) > 0]", tt.opts...)
			parser := newParserPrivate(tokenizer, tokenizer.Tokenize(), tt.opts...)
			if err := parser.parse(); err != nil {
				t.Fatalf("parse failed: %v", err)
			}

			nested := nestedJSONPathFromFunctionArg(parser.ast)
			if nested == nil {
				t.Fatalf("expected nested jsonpath query in function argument")
			}
			if nested.lazyContextTracking != tt.want {
				t.Fatalf("expected nested lazyContextTracking=%v, got %v", tt.want, nested.lazyContextTracking)
			}
		})
	}
}

func nestedJSONPathFromTestExpr(ast jsonPathAST) *jsonPathAST {
	if len(ast.segments) == 0 {
		return nil
	}
	seg := ast.segments[0]
	if seg == nil || seg.child == nil || len(seg.child.selectors) == 0 {
		return nil
	}
	sel := seg.child.selectors[0]
	if sel == nil || sel.filter == nil || sel.filter.expression == nil || len(sel.filter.expression.expressions) == 0 {
		return nil
	}
	andExpr := sel.filter.expression.expressions[0]
	if andExpr == nil || len(andExpr.expressions) == 0 {
		return nil
	}
	basic := andExpr.expressions[0]
	if basic == nil || basic.testExpr == nil || basic.testExpr.filterQuery == nil {
		return nil
	}
	return basic.testExpr.filterQuery.jsonPathQuery
}

func nestedJSONPathFromFunctionArg(ast jsonPathAST) *jsonPathAST {
	if len(ast.segments) == 0 {
		return nil
	}
	seg := ast.segments[0]
	if seg == nil || seg.child == nil || len(seg.child.selectors) == 0 {
		return nil
	}
	sel := seg.child.selectors[0]
	if sel == nil || sel.filter == nil || sel.filter.expression == nil || len(sel.filter.expression.expressions) == 0 {
		return nil
	}
	andExpr := sel.filter.expression.expressions[0]
	if andExpr == nil || len(andExpr.expressions) == 0 {
		return nil
	}
	basic := andExpr.expressions[0]
	if basic == nil || basic.comparisonExpr == nil || basic.comparisonExpr.left == nil || basic.comparisonExpr.left.functionExpr == nil {
		return nil
	}
	args := basic.comparisonExpr.left.functionExpr.args
	if len(args) == 0 || args[0] == nil || args[0].filterQuery == nil {
		return nil
	}
	return args[0].filterQuery.jsonPathQuery
}
