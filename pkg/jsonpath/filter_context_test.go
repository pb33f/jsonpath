package jsonpath

import (
	"testing"

	"go.yaml.in/yaml/v4"
)

func TestFilterContextLazyGuardsAndTrackingFlags(t *testing.T) {
	root := &yaml.Node{Kind: yaml.MappingNode}
	ctx := newFilterContextLazy(root).(*filterContext)
	node := &yaml.Node{Kind: yaml.ScalarNode, Value: "v"}

	// Guards should no-op before tracking is enabled.
	ctx.SetPropertyName("name")
	if ctx.PropertyName() != "" {
		t.Fatalf("expected empty property name before enabling tracking")
	}
	ctx.SetParent(root)
	if ctx.Parent() != nil {
		t.Fatalf("expected nil parent before enabling tracking")
	}
	ctx.SetParentPropertyName("pp")
	if ctx.ParentPropertyName() != "" {
		t.Fatalf("expected empty parent property before enabling path tracking")
	}
	ctx.SetIndex(3)
	if ctx.Index() != -1 {
		t.Fatalf("expected default index before enabling index tracking")
	}
	if got := ctx.Path(); got != "$" {
		t.Fatalf("expected root path when path tracking disabled, got %q", got)
	}
	ctx.PushPathSegment("['a']")
	ctx.PopPathSegment()
	ctx.SetPendingPathSegment(node, "['a']")
	if got := ctx.GetAndClearPendingPathSegment(node); got != "" {
		t.Fatalf("expected empty pending path when path tracking disabled, got %q", got)
	}
	ctx.SetPendingPropertyName(node, "a")
	if got := ctx.GetAndClearPendingPropertyName(node); got != "" {
		t.Fatalf("expected empty pending property when path tracking disabled, got %q", got)
	}
	if got := ctx.getPropertyKey(node); got != nil {
		t.Fatalf("expected nil property key when property tracking disabled")
	}
	if got := ctx.getParentNode(node); got != nil {
		t.Fatalf("expected nil parent node when parent tracking disabled")
	}
	ctx.parentNodes = nil
	ctx.setParentNode(node, root)
	if ctx.parentNodes != nil {
		t.Fatalf("expected setParentNode to no-op when parent tracking is disabled")
	}

	// Enable all tracking modes and verify values are persisted.
	ctx.EnablePropertyTracking()
	ctx.EnablePathTracking()
	ctx.EnableIndexTracking()
	ctx.EnableParentTracking()
	if !ctx.PropertyTrackingEnabled() || !ctx.PathTrackingEnabled() || !ctx.IndexTrackingEnabled() || !ctx.ParentTrackingEnabled() {
		t.Fatalf("expected all tracking flags enabled")
	}

	ctx.SetPropertyName("name")
	if ctx.PropertyName() != "name" {
		t.Fatalf("expected property name to be set")
	}
	ctx.SetParent(root)
	if ctx.Parent() != root {
		t.Fatalf("expected parent node to be set")
	}
	ctx.SetParentPropertyName("pp")
	if ctx.ParentPropertyName() != "pp" {
		t.Fatalf("expected parent property name to be set")
	}
	ctx.SetIndex(7)
	if ctx.Index() != 7 {
		t.Fatalf("expected index to be set")
	}

	ctx.PushPathSegment("['a']")
	if got := ctx.Path(); got != "$['a']" {
		t.Fatalf("unexpected path %q", got)
	}
	ctx.PopPathSegment()
	if got := ctx.Path(); got != "$" {
		t.Fatalf("expected path reset to root, got %q", got)
	}

	ctx.pendingPathSegments = nil
	ctx.SetPendingPathSegment(node, "['x']")
	if got := ctx.GetAndClearPendingPathSegment(node); got != "['x']" {
		t.Fatalf("expected pending path segment to be returned, got %q", got)
	}
	ctx.pendingPropertyNames = nil
	ctx.SetPendingPropertyName(node, "x")
	if got := ctx.GetAndClearPendingPropertyName(node); got != "x" {
		t.Fatalf("expected pending property name to be returned, got %q", got)
	}

	ctx.propertyKeys = nil
	ctx.setPropertyKey(node, root)
	if got := ctx.getPropertyKey(node); got != root {
		t.Fatalf("expected stored property key value")
	}

	ctx.parentNodes = nil
	ctx.setParentNode(node, root)
	if got := ctx.getParentNode(node); got != root {
		t.Fatalf("expected stored parent node value")
	}
}

func TestFilterContextClonePreservesTrackingFlags(t *testing.T) {
	root := &yaml.Node{Kind: yaml.MappingNode}
	ctx := newFilterContextLazy(root).(*filterContext)
	ctx.EnableParentTracking()
	ctx.EnablePropertyTracking()
	ctx.EnablePathTracking()
	ctx.EnableIndexTracking()
	ctx.SetPropertyName("name")
	ctx.SetParentPropertyName("pp")
	ctx.SetIndex(4)
	ctx.PushPathSegment("['a']")

	cloned := ctx.Clone().(*filterContext)
	if !cloned.ParentTrackingEnabled() || !cloned.PropertyTrackingEnabled() || !cloned.PathTrackingEnabled() || !cloned.IndexTrackingEnabled() {
		t.Fatalf("expected clone to preserve tracking flags")
	}
	if cloned.PropertyName() != "name" || cloned.ParentPropertyName() != "pp" || cloned.Index() != 4 {
		t.Fatalf("expected clone to preserve tracked values")
	}
	if got := cloned.Path(); got != "$['a']" {
		t.Fatalf("expected clone to preserve path, got %q", got)
	}
}
