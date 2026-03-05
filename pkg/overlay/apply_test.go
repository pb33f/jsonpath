package overlay_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/pb33f/jsonpath/pkg/overlay"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v4"
	"os"
	"strings"
)

// NodeMatchesFile is a test that marshals the YAML file from the given node,
// then compares those bytes to those found in the expected file.
func NodeMatchesFile(
	t *testing.T,
	actual *yaml.Node,
	expectedFile string,
	msgAndArgs ...any,
) {
	variadoc := func(pre ...any) []any { return append(msgAndArgs, pre...) }

	var actualBuf bytes.Buffer
	enc := yaml.NewEncoder(&actualBuf)
	enc.SetIndent(2)
	err := enc.Encode(actual)
	require.NoError(t, err, variadoc("failed to marshal node: ")...)

	expectedBytes, err := os.ReadFile(expectedFile)
	require.NoError(t, err, variadoc("failed to read expected file: ")...)

	// lazy redo snapshot
	//os.WriteFile(expectedFile, actualBuf.Bytes(), 0644)

	//t.Log("### EXPECT START ###\n" + string(expectedBytes) + "\n### EXPECT END ###\n")
	//t.Log("### ACTUAL START ###\n" + actualBuf.string() + "\n### ACTUAL END ###\n")

	// Normalize line endings for cross-platform compatibility (Windows CRLF vs Unix LF)
	expectedStr := strings.ReplaceAll(string(expectedBytes), "\r\n", "\n")
	actualStr := strings.ReplaceAll(actualBuf.String(), "\r\n", "\n")

	assert.Equal(t, expectedStr, actualStr, variadoc("node does not match expected file: ")...)
}

func TestApplyTo(t *testing.T) {
	t.Parallel()

	node, err := LoadSpecification("testdata/openapi.yaml")
	require.NoError(t, err)

	o, err := LoadOverlay("testdata/overlay.yaml")
	require.NoError(t, err)

	err = o.ApplyTo(node)
	assert.NoError(t, err)

	NodeMatchesFile(t, node, "testdata/openapi-overlayed.yaml")
}

func TestUpsertCreateNested(t *testing.T) {
	node, err := LoadSpecification("testdata/upsert/create-nested-input.yaml")
	require.NoError(t, err)

	o, err := LoadOverlay("testdata/upsert/create-nested-overlay.yaml")
	require.NoError(t, err)

	err = o.Validate()
	require.NoError(t, err)

	err = o.ApplyTo(node)
	require.NoError(t, err)

	NodeMatchesFile(t, node, "testdata/upsert/create-nested-expected.yaml")
}

func TestUpsertUpdateExisting(t *testing.T) {
	node, err := LoadSpecification("testdata/upsert/update-existing-input.yaml")
	require.NoError(t, err)

	o, err := LoadOverlay("testdata/upsert/update-existing-overlay.yaml")
	require.NoError(t, err)

	err = o.Validate()
	require.NoError(t, err)

	err = o.ApplyTo(node)
	require.NoError(t, err)

	NodeMatchesFile(t, node, "testdata/upsert/update-existing-expected.yaml")
}

func TestUpsertMultipleLevels(t *testing.T) {
	node, err := LoadSpecification("testdata/upsert/multiple-levels-input.yaml")
	require.NoError(t, err)

	o, err := LoadOverlay("testdata/upsert/multiple-levels-overlay.yaml")
	require.NoError(t, err)

	err = o.Validate()
	require.NoError(t, err)

	err = o.ApplyTo(node)
	require.NoError(t, err)

	NodeMatchesFile(t, node, "testdata/upsert/multiple-levels-expected.yaml")
}

func TestUpsertArrayElement(t *testing.T) {
	node, err := LoadSpecification("testdata/upsert/array-element-input.yaml")
	require.NoError(t, err)

	o, err := LoadOverlay("testdata/upsert/array-element-overlay.yaml")
	require.NoError(t, err)

	err = o.Validate()
	require.NoError(t, err)

	err = o.ApplyTo(node)
	require.NoError(t, err)

	NodeMatchesFile(t, node, "testdata/upsert/array-element-expected.yaml")
}

func TestUpsertRejectWildcard(t *testing.T) {
	o, err := LoadOverlay("testdata/upsert/reject-wildcard-overlay.yaml")
	require.NoError(t, err)

	err = o.Validate()
	require.Error(t, err)

	var valErrs overlay.ValidationErrors
	require.ErrorAs(t, err, &valErrs)
	found := false
	for _, e := range valErrs {
		var singErr *overlay.ActionUpsertNonSingularPathError
		if errors.As(e, &singErr) {
			found = true
			break
		}
	}
	assert.True(t, found, "expected ActionUpsertNonSingularPathError in validation errors")
}

func TestUpsertRejectRecursive(t *testing.T) {
	o, err := LoadOverlay("testdata/upsert/reject-recursive-overlay.yaml")
	require.NoError(t, err)

	err = o.Validate()
	require.Error(t, err)

	var valErrs overlay.ValidationErrors
	require.ErrorAs(t, err, &valErrs)
	found := false
	for _, e := range valErrs {
		var singErr *overlay.ActionUpsertNonSingularPathError
		if errors.As(e, &singErr) {
			found = true
			break
		}
	}
	assert.True(t, found, "expected ActionUpsertNonSingularPathError in validation errors")
}

func TestUpsertRejectFilter(t *testing.T) {
	o, err := LoadOverlay("testdata/upsert/reject-filter-overlay.yaml")
	require.NoError(t, err)

	err = o.Validate()
	require.Error(t, err)

	var valErrs overlay.ValidationErrors
	require.ErrorAs(t, err, &valErrs)
	found := false
	for _, e := range valErrs {
		var singErr *overlay.ActionUpsertNonSingularPathError
		if errors.As(e, &singErr) {
			found = true
			break
		}
	}
	assert.True(t, found, "expected ActionUpsertNonSingularPathError in validation errors")
}

func TestUpsertRejectUppsertWithRemove(t *testing.T) {
	o, err := LoadOverlay("testdata/upsert/reject-upsert-remove-overlay.yaml")
	require.NoError(t, err)

	err = o.Validate()
	require.Error(t, err)

	var valErrs overlay.ValidationErrors
	require.ErrorAs(t, err, &valErrs)
	found := false
	for _, e := range valErrs {
		var conflictErr *overlay.ActionUpsertRemoveConflictError
		if errors.As(e, &conflictErr) {
			found = true
			break
		}
	}
	assert.True(t, found, "expected ActionUpsertRemoveConflictError in validation errors")
}

func TestUpsertArrayOutOfBounds(t *testing.T) {
	node, err := LoadSpecification("testdata/upsert/array-element-input.yaml")
	require.NoError(t, err)

	var ovl overlay.Overlay
	data := []byte(`
overlay: "1.0.0"
info:
  title: test
  version: "1.0"
actions:
  - target: $.items[5].name
    update: value
    upsert: true
`)

	err = yaml.Unmarshal(data, &ovl)
	require.NoError(t, err)

	err = ovl.Validate()
	require.NoError(t, err)

	err = ovl.ApplyTo(node)
	require.Error(t, err)

	var oobErr *overlay.ArrayIndexOutOfBoundsError
	require.ErrorAs(t, err, &oobErr)
	assert.Equal(t, int64(5), oobErr.Index)
	assert.Equal(t, 1, oobErr.Length)
}

func TestUpsertTypeMismatch(t *testing.T) {
	node, err := LoadSpecification("testdata/upsert/array-element-input.yaml")
	require.NoError(t, err)

	var ovl overlay.Overlay
	data := []byte(`
overlay: "1.0.0"
info:
  title: test
  version: "1.0"
actions:
  - target: $.items[0].name.extra
    update: value
    upsert: true
`)

	err = yaml.Unmarshal(data, &ovl)
	require.NoError(t, err)

	err = ovl.Validate()
	require.NoError(t, err)

	err = ovl.ApplyTo(node)
	require.Error(t, err)

	var typeErr *overlay.TypeMismatchError
	require.ErrorAs(t, err, &typeErr)
	assert.Equal(t, "mapping node", typeErr.Expected)
	assert.Equal(t, "scalar node", typeErr.Actual)
}

func TestUpsertBracketNotation(t *testing.T) {
	node, err := LoadSpecification("testdata/upsert/multiple-levels-input.yaml")
	require.NoError(t, err)

	var ovl overlay.Overlay
	data := []byte(`
overlay: "1.0.0"
info:
  title: test
  version: "1.0"
actions:
  - target: $['spec']['config']['key']
    update: value
    upsert: true
`)
	err = yaml.Unmarshal(data, &ovl)
	require.NoError(t, err)

	err = ovl.Validate()
	require.NoError(t, err)

	err = ovl.ApplyTo(node)
	require.NoError(t, err)

	var actualBuf bytes.Buffer
	enc := yaml.NewEncoder(&actualBuf)
	enc.SetIndent(2)
	err = enc.Encode(node)
	require.NoError(t, err)

	assert.Contains(t, actualBuf.String(), "config:")
	assert.Contains(t, actualBuf.String(), "key: value")
}

func TestUpsertArrayIndexFinal(t *testing.T) {
	node, err := LoadSpecification("testdata/upsert/array-final-input.yaml")
	require.NoError(t, err)

	o, err := LoadOverlay("testdata/upsert/array-final-overlay.yaml")
	require.NoError(t, err)

	err = o.Validate()
	require.NoError(t, err)

	err = o.ApplyTo(node)
	require.NoError(t, err)

	NodeMatchesFile(t, node, "testdata/upsert/array-final-expected.yaml")
}

func TestUpsertArrayIntoNonSequence(t *testing.T) {
	node, err := LoadSpecification("testdata/upsert/array-into-nonseq-input.yaml")
	require.NoError(t, err)

	var ovl overlay.Overlay
	data := []byte(`
overlay: "1.0.0"
info:
  title: test
  version: "1.0"
actions:
  - target: $.scalar[0]
    update: value
    upsert: true
`)

	err = yaml.Unmarshal(data, &ovl)
	require.NoError(t, err)

	err = ovl.Validate()
	require.NoError(t, err)

	err = ovl.ApplyTo(node)
	require.Error(t, err)

	var typeErr *overlay.TypeMismatchError
	require.ErrorAs(t, err, &typeErr)
	assert.Equal(t, "sequence node", typeErr.Expected)
	assert.Equal(t, "scalar node", typeErr.Actual)
}

func TestUpsertMapKeyIntoNonMap(t *testing.T) {
	node, err := LoadSpecification("testdata/upsert/array-into-nonseq-input.yaml")
	require.NoError(t, err)

	var ovl overlay.Overlay
	data := []byte(`
overlay: "1.0.0"
info:
  title: test
  version: "1.0"
actions:
  - target: $.scalar.somekey
    update: value
    upsert: true
`)

	err = yaml.Unmarshal(data, &ovl)
	require.NoError(t, err)

	err = ovl.Validate()
	require.NoError(t, err)

	err = ovl.ApplyTo(node)
	require.Error(t, err)

	var typeErr *overlay.TypeMismatchError
	require.ErrorAs(t, err, &typeErr)
	assert.Equal(t, "mapping node", typeErr.Expected)
	assert.Equal(t, "scalar node", typeErr.Actual)
}

func TestUpsertSetArrayValueDirect(t *testing.T) {
	node, err := LoadSpecification("testdata/upsert/set-array-value-input.yaml")
	require.NoError(t, err)

	o, err := LoadOverlay("testdata/upsert/set-array-value-overlay.yaml")
	require.NoError(t, err)

	err = o.Validate()
	require.NoError(t, err)

	err = o.ApplyTo(node)
	require.NoError(t, err)

	NodeMatchesFile(t, node, "testdata/upsert/set-array-value-expected.yaml")
}

func TestUpsertCreateNestedArray(t *testing.T) {
	node, err := LoadSpecification("testdata/upsert/create-nested-array-input.yaml")
	require.NoError(t, err)

	o, err := LoadOverlay("testdata/upsert/create-nested-array-overlay.yaml")
	require.NoError(t, err)

	err = o.Validate()
	require.NoError(t, err)

	err = o.ApplyTo(node)
	require.NoError(t, err)

	var actualBuf bytes.Buffer
	enc := yaml.NewEncoder(&actualBuf)
	enc.SetIndent(2)
	err = enc.Encode(node)
	require.NoError(t, err)

	assert.Contains(t, actualBuf.String(), "items:")
	assert.Contains(t, actualBuf.String(), "name: first-item")
}

func TestUpsertArrayIndexUpdateExisting(t *testing.T) {
	node, err := LoadSpecification("testdata/upsert/array-index-update-input.yaml")
	require.NoError(t, err)

	o, err := LoadOverlay("testdata/upsert/array-index-update-overlay.yaml")
	require.NoError(t, err)

	err = o.Validate()
	require.NoError(t, err)

	err = o.ApplyTo(node)
	require.NoError(t, err)

	NodeMatchesFile(t, node, "testdata/upsert/array-index-update-expected.yaml")
}

func TestUpsertAppendToExistingArray(t *testing.T) {
	node, err := LoadSpecification("testdata/upsert/array-element-input.yaml")
	require.NoError(t, err)

	var ovl overlay.Overlay
	data := []byte(`
overlay: "1.0.0"
info:
  title: test
  version: "1.0"
actions:
  - target: $.items[1].name
    update: appended-item
    upsert: true
`)

	err = yaml.Unmarshal(data, &ovl)
	require.NoError(t, err)

	err = ovl.Validate()
	require.NoError(t, err)

	err = ovl.ApplyTo(node)
	require.NoError(t, err)

	var actualBuf bytes.Buffer
	enc := yaml.NewEncoder(&actualBuf)
	enc.SetIndent(2)
	err = enc.Encode(node)
	require.NoError(t, err)

	assert.Contains(t, actualBuf.String(), "- name: first")
	assert.Contains(t, actualBuf.String(), "- name: appended-item")
}

func TestUpsertIndexZeroOnEmptyArray(t *testing.T) {
	var node yaml.Node
	err := yaml.Unmarshal([]byte("items: []"), &node)
	require.NoError(t, err)

	var ovl overlay.Overlay
	data := []byte(`
overlay: "1.0.0"
info:
  title: test
  version: "1.0"
actions:
  - target: $.items[0]
    update: first-item
    upsert: true
`)

	err = yaml.Unmarshal(data, &ovl)
	require.NoError(t, err)

	err = ovl.Validate()
	require.NoError(t, err)

	err = ovl.ApplyTo(node.Content[0])
	require.NoError(t, err)

	var actualBuf bytes.Buffer
	enc := yaml.NewEncoder(&actualBuf)
	enc.SetIndent(2)
	err = enc.Encode(node.Content[0])
	require.NoError(t, err)

	assert.Contains(t, actualBuf.String(), "items:")
	assert.Contains(t, actualBuf.String(), "first-item")
}

func TestUpsertIndexZeroCreatesArray(t *testing.T) {
	var node yaml.Node
	err := yaml.Unmarshal([]byte("{}"), &node)
	require.NoError(t, err)

	var ovl overlay.Overlay
	data := []byte(`
overlay: "1.0.0"
info:
  title: test
  version: "1.0"
actions:
  - target: $.items[0].name
    update: created-item
    upsert: true
`)

	err = yaml.Unmarshal(data, &ovl)
	require.NoError(t, err)

	err = ovl.Validate()
	require.NoError(t, err)

	err = ovl.ApplyTo(node.Content[0])
	require.NoError(t, err)

	var actualBuf bytes.Buffer
	enc := yaml.NewEncoder(&actualBuf)
	enc.SetIndent(2)
	err = enc.Encode(node.Content[0])
	require.NoError(t, err)

	assert.Contains(t, actualBuf.String(), "items:")
	assert.Contains(t, actualBuf.String(), "name: created-item")
}

func TestUpsertRejectIndexBeyondLength(t *testing.T) {
	node, err := LoadSpecification("testdata/upsert/array-element-input.yaml")
	require.NoError(t, err)

	var ovl overlay.Overlay
	data := []byte(`
overlay: "1.0.0"
info:
  title: test
  version: "1.0"
actions:
  - target: $.items[5].name
    update: value
    upsert: true
`)

	err = yaml.Unmarshal(data, &ovl)
	require.NoError(t, err)

	err = ovl.Validate()
	require.NoError(t, err)

	err = ovl.ApplyTo(node)
	require.Error(t, err)

	var oobErr *overlay.ArrayIndexOutOfBoundsError
	require.ErrorAs(t, err, &oobErr)
	assert.Equal(t, int64(5), oobErr.Index)
	assert.Equal(t, 1, oobErr.Length)
	assert.False(t, oobErr.CanAppend)
}

func TestUpsertRejectNonZeroIndexNewArray(t *testing.T) {
	var node yaml.Node
	err := yaml.Unmarshal([]byte("{}"), &node)
	require.NoError(t, err)

	var ovl overlay.Overlay
	data := []byte(`
overlay: "1.0.0"
info:
  title: test
  version: "1.0"
actions:
  - target: $.items[2].name
    update: value
    upsert: true
`)

	err = yaml.Unmarshal(data, &ovl)
	require.NoError(t, err)

	err = ovl.Validate()
	require.NoError(t, err)

	err = ovl.ApplyTo(node.Content[0])
	require.Error(t, err)

	var createErr *overlay.CannotCreateArrayAtIndexError
	require.ErrorAs(t, err, &createErr)
	assert.Equal(t, int64(2), createErr.Index)
}

func TestUpsertRejectNonZeroIndexNestedNewArray(t *testing.T) {
	var node yaml.Node
	err := yaml.Unmarshal([]byte("{}"), &node)
	require.NoError(t, err)

	var ovl overlay.Overlay
	data := []byte(`
overlay: "1.0.0"
info:
  title: test
  version: "1.0"
actions:
  - target: $.a.b[3].c
    update: value
    upsert: true
`)

	err = yaml.Unmarshal(data, &ovl)
	require.NoError(t, err)

	err = ovl.Validate()
	require.NoError(t, err)

	err = ovl.ApplyTo(node.Content[0])
	require.Error(t, err)

	var createErr *overlay.CannotCreateArrayAtIndexError
	require.ErrorAs(t, err, &createErr)
	assert.Equal(t, int64(3), createErr.Index)
}

func TestUpsertNegativeArrayIndex(t *testing.T) {
	var node yaml.Node
	err := yaml.Unmarshal([]byte("{}"), &node)
	require.NoError(t, err)

	var ovl overlay.Overlay
	data := []byte(`
overlay: "1.0.0"
info:
  title: test
  version: "1.0"
actions:
  - target: $.items[-1].name
    update: value
    upsert: true
`)

	err = yaml.Unmarshal(data, &ovl)
	require.NoError(t, err)

	err = ovl.Validate()
	require.NoError(t, err)

	err = ovl.ApplyTo(node.Content[0])
	require.Error(t, err)

	var negErr *overlay.NegativeArrayIndexError
	require.ErrorAs(t, err, &negErr)
	assert.Equal(t, int64(-1), negErr.Index)
}

func TestUpsertArrayIndexOnScalar(t *testing.T) {
	var node yaml.Node
	err := yaml.Unmarshal([]byte("scalar: value"), &node)
	require.NoError(t, err)

	var ovl overlay.Overlay
	data := []byte(`
overlay: "1.0.0"
info:
  title: test
  version: "1.0"
actions:
  - target: $.scalar[0]
    update: newvalue
    upsert: true
`)

	err = yaml.Unmarshal(data, &ovl)
	require.NoError(t, err)

	err = ovl.Validate()
	require.NoError(t, err)

	err = ovl.ApplyTo(node.Content[0])
	require.Error(t, err)

	var typeErr *overlay.TypeMismatchError
	require.ErrorAs(t, err, &typeErr)
	assert.Equal(t, "sequence node", typeErr.Expected)
	assert.Equal(t, "scalar node", typeErr.Actual)
}

func TestUpsertArrayIndexOnMap(t *testing.T) {
	var node yaml.Node
	err := yaml.Unmarshal([]byte("mapkey: {nested: value}"), &node)
	require.NoError(t, err)

	var ovl overlay.Overlay
	data := []byte(`
overlay: "1.0.0"
info:
  title: test
  version: "1.0"
actions:
  - target: $.mapkey[0].field
    update: value
    upsert: true
`)

	err = yaml.Unmarshal(data, &ovl)
	require.NoError(t, err)

	err = ovl.Validate()
	require.NoError(t, err)

	err = ovl.ApplyTo(node.Content[0])
	require.Error(t, err)

	var typeErr *overlay.TypeMismatchError
	require.ErrorAs(t, err, &typeErr)
	assert.Equal(t, "sequence node", typeErr.Expected)
	assert.Equal(t, "mapping node", typeErr.Actual)
}

func TestUpsertNonZeroIndexOnNilArrayInPath(t *testing.T) {
	var node yaml.Node
	err := yaml.Unmarshal([]byte("root: {}"), &node)
	require.NoError(t, err)

	var ovl overlay.Overlay
	data := []byte(`
overlay: "1.0.0"
info:
  title: test
  version: "1.0"
actions:
  - target: $.root.items[1].field
    update: value
    upsert: true
`)

	err = yaml.Unmarshal(data, &ovl)
	require.NoError(t, err)

	err = ovl.Validate()
	require.NoError(t, err)

	err = ovl.ApplyTo(node.Content[0])
	require.Error(t, err)

	var createErr *overlay.CannotCreateArrayAtIndexError
	require.ErrorAs(t, err, &createErr)
	assert.Equal(t, int64(1), createErr.Index)
}

func TestUpsertNonZeroFinalIndexOnNilArray(t *testing.T) {
	var node yaml.Node
	err := yaml.Unmarshal([]byte("root: {}"), &node)
	require.NoError(t, err)

	var ovl overlay.Overlay
	data := []byte(`
overlay: "1.0.0"
info:
  title: test
  version: "1.0"
actions:
  - target: $.root.items[2]
    update: value
    upsert: true
`)

	err = yaml.Unmarshal(data, &ovl)
	require.NoError(t, err)

	err = ovl.Validate()
	require.NoError(t, err)

	err = ovl.ApplyTo(node.Content[0])
	require.Error(t, err)

	var createErr *overlay.CannotCreateArrayAtIndexError
	require.ErrorAs(t, err, &createErr)
	assert.Equal(t, int64(2), createErr.Index)
}

func TestUpsertAppendAtExactLength(t *testing.T) {
	var node yaml.Node
	err := yaml.Unmarshal([]byte("items:\n  - first\n  - second"), &node)
	require.NoError(t, err)

	var ovl overlay.Overlay
	data := []byte(`
overlay: "1.0.0"
info:
  title: test
  version: "1.0"
actions:
  - target: $.items[2]
    update: third
    upsert: true
`)

	err = yaml.Unmarshal(data, &ovl)
	require.NoError(t, err)

	err = ovl.Validate()
	require.NoError(t, err)

	err = ovl.ApplyTo(node.Content[0])
	require.NoError(t, err)

	var actualBuf bytes.Buffer
	enc := yaml.NewEncoder(&actualBuf)
	enc.SetIndent(2)
	err = enc.Encode(node.Content[0])
	require.NoError(t, err)

	assert.Contains(t, actualBuf.String(), "- first")
	assert.Contains(t, actualBuf.String(), "- second")
	assert.Contains(t, actualBuf.String(), "- third")
}

func TestUpsertNegativeIndexOnFinalSegment(t *testing.T) {
	var node yaml.Node
	err := yaml.Unmarshal([]byte("root: {}"), &node)
	require.NoError(t, err)

	var ovl overlay.Overlay
	data := []byte(`
overlay: "1.0.0"
info:
  title: test
  version: "1.0"
actions:
  - target: $.root.items[-1]
    update: value
    upsert: true
`)

	err = yaml.Unmarshal(data, &ovl)
	require.NoError(t, err)

	err = ovl.Validate()
	require.NoError(t, err)

	err = ovl.ApplyTo(node.Content[0])
	require.Error(t, err)

	var negErr *overlay.NegativeArrayIndexError
	require.ErrorAs(t, err, &negErr)
	assert.Equal(t, int64(-1), negErr.Index)
}

func TestUpsertErrorMessageFormatting(t *testing.T) {
	err1 := &overlay.UpsertError{Path: "$.items", Msg: "test error", Index: 5, Len: 3}
	assert.Contains(t, err1.Error(), "$.items")
	assert.Contains(t, err1.Error(), "test error")
	assert.Contains(t, err1.Error(), "index 5")
	assert.Contains(t, err1.Error(), "length 3")

	err2 := &overlay.UpsertError{Path: "$.items", Msg: "simple error"}
	assert.Contains(t, err2.Error(), "$.items")
	assert.Contains(t, err2.Error(), "simple error")

	err3 := &overlay.UpsertError{Msg: "no path error"}
	assert.Contains(t, err3.Error(), "no path error")

	err4 := &overlay.ArrayIndexOutOfBoundsError{Index: 10, Length: 5, CanAppend: true}
	assert.Contains(t, err4.Error(), "index 10")
	assert.Contains(t, err4.Error(), "length 5")
	assert.Contains(t, err4.Error(), "only appending")

	err5 := &overlay.ArrayIndexOutOfBoundsError{Index: 10, Length: 5, CanAppend: false}
	assert.Contains(t, err5.Error(), "index 10")
	assert.NotContains(t, err5.Error(), "only appending")

	err6 := &overlay.NegativeArrayIndexError{Index: -5}
	assert.Contains(t, err6.Error(), "-5")

	err7 := &overlay.UnknownSegmentKindError{Kind: 99}
	assert.Contains(t, err7.Error(), "99")

	err8 := &overlay.EmptyPathError{}
	assert.Contains(t, err8.Error(), "empty path")
}

func TestUpsertIndexOutOfBoundsOnExistingArrayFinal(t *testing.T) {
	var node yaml.Node
	err := yaml.Unmarshal([]byte("items:\n  - first"), &node)
	require.NoError(t, err)

	var ovl overlay.Overlay
	data := []byte(`
overlay: "1.0.0"
info:
  title: test
  version: "1.0"
actions:
  - target: $.items[5]
    update: value
    upsert: true
`)

	err = yaml.Unmarshal(data, &ovl)
	require.NoError(t, err)

	err = ovl.Validate()
	require.NoError(t, err)

	err = ovl.ApplyTo(node.Content[0])
	require.Error(t, err)

	var oobErr *overlay.ArrayIndexOutOfBoundsError
	require.ErrorAs(t, err, &oobErr)
	assert.Equal(t, int64(5), oobErr.Index)
	assert.Equal(t, 1, oobErr.Length)
	assert.False(t, oobErr.CanAppend)
}
