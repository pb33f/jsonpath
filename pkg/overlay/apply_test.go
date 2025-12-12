package overlay_test

import (
    "bytes"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "go.yaml.in/yaml/v4"
    "os"
    "strings"
    "testing"
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
