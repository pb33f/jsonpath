package jsonpath

import (
	"encoding/json"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/pb33f/jsonpath/pkg/jsonpath/config"
	"go.yaml.in/yaml/v4"
)

type spectralCorpus struct {
	Baseline struct {
		SpectralCommit      string `json:"spectralCommit"`
		SpectralCoreVersion string `json:"spectralCoreVersion"`
		NimmaVersion        string `json:"nimmaVersion"`
		JSONPathPlusVersion string `json:"jsonPathPlusVersion"`
		GeneratedBy         string `json:"generatedBy"`
		Unsafe              bool   `json:"unsafe"`
	} `json:"baseline"`
	Cases []struct {
		Name          string         `json:"name"`
		Source        string         `json:"source"`
		Expression    string         `json:"expression"`
		Input         map[string]any `json:"input"`
		Engine        string         `json:"engine"`
		ExpectedPaths []string       `json:"expectedPaths"`
		Extension     bool           `json:"extension"`
	} `json:"cases"`
}

func TestSpectralDifferentialCorpus(t *testing.T) {
	data, err := os.ReadFile("testdata/spectral/corpus.json")
	if err != nil {
		t.Fatalf("read corpus: %v", err)
	}
	var corpus spectralCorpus
	if err := json.Unmarshal(data, &corpus); err != nil {
		t.Fatalf("decode corpus: %v", err)
	}
	if corpus.Baseline.SpectralCommit == "" || corpus.Baseline.SpectralCoreVersion == "" ||
		corpus.Baseline.NimmaVersion == "" || corpus.Baseline.JSONPathPlusVersion == "" ||
		corpus.Baseline.GeneratedBy == "" || corpus.Baseline.Unsafe {
		t.Fatalf("incomplete or unsafe Spectral baseline: %+v", corpus.Baseline)
	}
	if len(corpus.Cases) == 0 {
		t.Fatal("Spectral corpus is empty")
	}

	for _, testCase := range corpus.Cases {
		t.Run(testCase.Name, func(t *testing.T) {
			if testCase.Source == "" || (testCase.Engine != "nimma" && testCase.Engine != "jsonpath-plus") {
				t.Fatalf("case is missing source or selected engine: %+v", testCase)
			}
			input, err := json.Marshal(testCase.Input)
			if err != nil {
				t.Fatalf("encode input: %v", err)
			}
			var document yaml.Node
			if err := yaml.Unmarshal(input, &document); err != nil {
				t.Fatalf("decode YAML node: %v", err)
			}
			compiled, err := NewPath(testCase.Expression, config.WithSpectralCompatibility())
			if err != nil {
				t.Fatalf("compile: %v", err)
			}
			got := normalizedResultPaths(&document, compiled.Query(&document))
			want := append([]string(nil), testCase.ExpectedPaths...)
			sort.Strings(want)
			if !equalStrings(got, want) {
				t.Fatalf("engine=%s paths=%v, want %v", testCase.Engine, got, want)
			}

			roundTrip, err := NewPath(compiled.String(), config.WithSpectralCompatibility())
			if err != nil {
				t.Fatalf("compile round trip %q: %v", compiled.String(), err)
			}
			if roundTripPaths := normalizedResultPaths(&document, roundTrip.Query(&document)); !equalStrings(roundTripPaths, want) {
				t.Fatalf("round-trip paths=%v, want %v", roundTripPaths, want)
			}

			if testCase.Extension {
				if _, err := NewPath(testCase.Expression, config.WithStrictRFC9535()); err == nil {
					t.Fatalf("strict RFC mode accepted Spectral-only expression")
				}
			}
		})
	}
}

func TestSpectralPinnedOfficialSelectorsCompile(t *testing.T) {
	data, err := os.ReadFile("testdata/spectral/official-selectors.json")
	if err != nil {
		t.Fatalf("read official selectors: %v", err)
	}
	var inventory struct {
		Repository string   `json:"repository"`
		License    string   `json:"license"`
		Commit     string   `json:"commit"`
		Files      []string `json:"files"`
		Selectors  []string `json:"selectors"`
	}
	if err := json.Unmarshal(data, &inventory); err != nil {
		t.Fatalf("decode official selectors: %v", err)
	}
	if inventory.Repository == "" || inventory.License == "" || inventory.Commit == "" || len(inventory.Files) != 2 || len(inventory.Selectors) < 100 {
		t.Fatalf("official selector inventory is incomplete: commit=%q files=%d selectors=%d", inventory.Commit, len(inventory.Files), len(inventory.Selectors))
	}
	for _, expression := range inventory.Selectors {
		compiled, err := NewPath(expression, config.WithSpectralCompatibility())
		if err != nil {
			t.Errorf("compile %s: %v", expression, err)
			continue
		}
		if _, err := NewPath(compiled.String(), config.WithSpectralCompatibility()); err != nil {
			t.Errorf("compile round trip %s as %s: %v", expression, compiled.String(), err)
		}
	}
}

func TestSpectralPinnedPublicSelectorsCompile(t *testing.T) {
	data, err := os.ReadFile("testdata/spectral/public-selectors.json")
	if err != nil {
		t.Fatalf("read public selectors: %v", err)
	}
	var inventory struct {
		GeneratedBy string `json:"generatedBy"`
		Sources     []struct {
			Name       string `json:"name"`
			Repository string `json:"repository"`
			License    string `json:"license"`
			Commit     string `json:"commit"`
			Path       string `json:"path"`
		} `json:"sources"`
		Selectors []struct {
			Expression string   `json:"expression"`
			Sources    []string `json:"sources"`
			Status     string   `json:"status"`
			Reason     string   `json:"reason"`
		} `json:"selectors"`
	}
	if err := json.Unmarshal(data, &inventory); err != nil {
		t.Fatalf("decode public selectors: %v", err)
	}
	if inventory.GeneratedBy == "" || len(inventory.Sources) != 4 || len(inventory.Selectors) < 50 {
		t.Fatalf("public selector inventory is incomplete: generatedBy=%q sources=%d selectors=%d", inventory.GeneratedBy, len(inventory.Sources), len(inventory.Selectors))
	}
	sourceNames := make(map[string]struct{}, len(inventory.Sources))
	for _, source := range inventory.Sources {
		if source.Name == "" || source.Repository == "" || source.License == "" || source.Commit == "" || source.Path == "" {
			t.Fatalf("public source metadata is incomplete: %+v", source)
		}
		sourceNames[source.Name] = struct{}{}
	}
	for _, entry := range inventory.Selectors {
		if entry.Status != "supported" || entry.Reason != "" {
			t.Errorf("public selector %q classified as %s: %s", entry.Expression, entry.Status, entry.Reason)
			continue
		}
		if entry.Expression == "" || strings.HasPrefix(entry.Expression, "#") || len(entry.Sources) == 0 {
			t.Errorf("invalid or unresolved public selector entry: %+v", entry)
			continue
		}
		for _, source := range entry.Sources {
			if _, ok := sourceNames[source]; !ok {
				t.Errorf("selector %q references unknown source %q", entry.Expression, source)
			}
		}
		compiled, err := NewPath(entry.Expression, config.WithSpectralCompatibility())
		if err != nil {
			t.Errorf("compile public selector %s: %v", entry.Expression, err)
			continue
		}
		if _, err := NewPath(compiled.String(), config.WithSpectralCompatibility()); err != nil {
			t.Errorf("compile public selector round trip %s as %s: %v", entry.Expression, compiled.String(), err)
		}
	}
}

func TestSpectralPinnedOfficialSelectorParity(t *testing.T) {
	data, err := os.ReadFile("testdata/spectral/official-parity.json")
	if err != nil {
		t.Fatalf("read official parity corpus: %v", err)
	}
	var parity struct {
		Baseline struct {
			SpectralCommit   string `json:"spectralCommit"`
			GeneratedBy      string `json:"generatedBy"`
			SourceRepository string `json:"sourceRepository"`
			SourceLicense    string `json:"sourceLicense"`
			Unsafe           bool   `json:"unsafe"`
		} `json:"baseline"`
		Input map[string]any `json:"input"`
		Cases []struct {
			Expression    string   `json:"expression"`
			Engine        string   `json:"engine"`
			ExpectedPaths []string `json:"expectedPaths"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(data, &parity); err != nil {
		t.Fatalf("decode official parity corpus: %v", err)
	}
	if parity.Baseline.SpectralCommit == "" || parity.Baseline.GeneratedBy == "" || parity.Baseline.SourceRepository == "" || parity.Baseline.SourceLicense == "" || parity.Baseline.Unsafe || len(parity.Cases) < 100 {
		t.Fatalf("official parity baseline is incomplete: commit=%q generatedBy=%q unsafe=%v cases=%d", parity.Baseline.SpectralCommit, parity.Baseline.GeneratedBy, parity.Baseline.Unsafe, len(parity.Cases))
	}
	input, err := json.Marshal(parity.Input)
	if err != nil {
		t.Fatalf("encode official input: %v", err)
	}
	var document yaml.Node
	if err := yaml.Unmarshal(input, &document); err != nil {
		t.Fatalf("decode official input node: %v", err)
	}
	for _, testCase := range parity.Cases {
		t.Run(testCase.Expression, func(t *testing.T) {
			compiled, err := NewPath(testCase.Expression, config.WithSpectralCompatibility())
			if err != nil {
				t.Fatalf("compile: %v", err)
			}
			got := normalizedResultPaths(&document, compiled.Query(&document))
			want := append([]string(nil), testCase.ExpectedPaths...)
			sort.Strings(want)
			if !equalStrings(got, want) {
				t.Fatalf("engine=%s paths=%v, want %v", testCase.Engine, got, want)
			}
		})
	}
}

func normalizedResultPaths(document *yaml.Node, results []*yaml.Node) []string {
	paths := make(map[*yaml.Node]string)
	collectNormalizedPaths(unwrapDocument(document), "$", paths)
	normalized := make([]string, 0, len(results))
	for _, result := range results {
		if path, ok := paths[result]; ok {
			normalized = append(normalized, path)
		} else {
			normalized = append(normalized, "<unknown>")
		}
	}
	sort.Strings(normalized)
	return normalized
}

func collectNormalizedPaths(node *yaml.Node, path string, paths map[*yaml.Node]string) {
	if node == nil {
		return
	}
	paths[node] = path
	switch node.Kind {
	case yaml.MappingNode:
		for i := 0; i+1 < len(node.Content); i += 2 {
			key := node.Content[i]
			value := node.Content[i+1]
			childPath := path + "['" + escapeNormalizedPathKey(key.Value) + "']"
			paths[key] = childPath
			collectNormalizedPaths(value, childPath, paths)
		}
	case yaml.SequenceNode:
		for i, child := range node.Content {
			collectNormalizedPaths(child, path+"["+strconv.Itoa(i)+"]", paths)
		}
	}
}

func escapeNormalizedPathKey(value string) string {
	return strings.ReplaceAll(strings.ReplaceAll(value, `\`, `\\`), `'`, `\'`)
}
