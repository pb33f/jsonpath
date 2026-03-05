package overlay_test

import (
	"fmt"
	"github.com/pb33f/jsonpath/pkg/overlay"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v4"
	"os"
	"strings"
	"testing"
)

func LoadSpecification(path string) (*yaml.Node, error) {
	rs, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open schema from path %q: %w", path, err)
	}

	var ys yaml.Node
	err = yaml.NewDecoder(rs).Decode(&ys)
	if err != nil {
		return nil, fmt.Errorf("failed to parse schema at path %q: %w", path, err)
	}

	return &ys, nil
}

func LoadOverlay(path string) (*overlay.Overlay, error) {
	o, err := overlay.Parse(path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse overlay from path %q: %w", path, err)
	}

	return o, nil
}

func TestCompare(t *testing.T) {
	t.Parallel()

	node, err := LoadSpecification("testdata/openapi.yaml")
	require.NoError(t, err)
	node2, err := LoadSpecification("testdata/openapi-overlayed.yaml")
	require.NoError(t, err)

	o, err := LoadOverlay("testdata/overlay-generated.yaml")
	require.NoError(t, err)

	o2, err := overlay.Compare("Drinks Overlay", node, *node2)
	assert.NoError(t, err)

	o1s, err := o.ToString()
	assert.NoError(t, err)
	o2s, err := o2.ToString()
	assert.NoError(t, err)

	// Uncomment this if we've improved the output
	//os.WriteFile("testdata/overlay-generated.yaml", []byte(o2s), 0644)

	// Normalize line endings for cross-platform compatibility (Windows CRLF vs Unix LF)
	o1sNorm := strings.ReplaceAll(o1s, "\r\n", "\n")
	o2sNorm := strings.ReplaceAll(o2s, "\r\n", "\n")
	assert.Equal(t, o1sNorm, o2sNorm)

	// round trip it
	err = o.ApplyTo(node)
	assert.NoError(t, err)
	NodeMatchesFile(t, node, "testdata/openapi-overlayed.yaml")

}

func TestCompareWithAdditions(t *testing.T) {
	t.Parallel()

	var y1, y2 yaml.Node
	err := yaml.Unmarshal([]byte("key: value"), &y1)
	require.NoError(t, err)
	err = yaml.Unmarshal([]byte("key: value\nnewkey: newvalue"), &y2)
	require.NoError(t, err)

	o, err := overlay.Compare("Add Key", &y1, y2)
	require.NoError(t, err)
	assert.NotEmpty(t, o.Actions)
	assert.Equal(t, "Add Key", o.Info.Title)
}

func TestCompareWithRemovals(t *testing.T) {
	t.Parallel()

	var y1, y2 yaml.Node
	err := yaml.Unmarshal([]byte("key: value\noldkey: oldvalue"), &y1)
	require.NoError(t, err)
	err = yaml.Unmarshal([]byte("key: value"), &y2)
	require.NoError(t, err)

	o, err := overlay.Compare("Remove Key", &y1, y2)
	require.NoError(t, err)
	assert.NotEmpty(t, o.Actions)
	hasRemove := false
	for _, a := range o.Actions {
		if a.Remove {
			hasRemove = true
		}
	}
	assert.True(t, hasRemove)
}

func TestCompareWithScalarChange(t *testing.T) {
	t.Parallel()

	var y1, y2 yaml.Node
	err := yaml.Unmarshal([]byte("key: old"), &y1)
	require.NoError(t, err)
	err = yaml.Unmarshal([]byte("key: new"), &y2)
	require.NoError(t, err)

	o, err := overlay.Compare("Change Value", &y1, y2)
	require.NoError(t, err)
	require.Len(t, o.Actions, 1)
}

func TestCompareWithArrayAppend(t *testing.T) {
	t.Parallel()

	var y1, y2 yaml.Node
	err := yaml.Unmarshal([]byte("items: [a, b]"), &y1)
	require.NoError(t, err)
	err = yaml.Unmarshal([]byte("items: [a, b, c]"), &y2)
	require.NoError(t, err)

	o, err := overlay.Compare("Append Item", &y1, y2)
	require.NoError(t, err)
	assert.NotEmpty(t, o.Actions)
}

func TestCompareWithArrayTypeChange(t *testing.T) {
	t.Parallel()

	var y1, y2 yaml.Node
	err := yaml.Unmarshal([]byte("items: [a, b, c]"), &y1)
	require.NoError(t, err)
	err = yaml.Unmarshal([]byte("items: [x, y]"), &y2)
	require.NoError(t, err)

	o, err := overlay.Compare("Replace Array", &y1, y2)
	require.NoError(t, err)
	assert.NotEmpty(t, o.Actions)
}
