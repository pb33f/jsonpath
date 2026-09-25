package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/pb33f/go-yaml"
	"github.com/pb33f/jsonpath/pkg/jsonpath"
	"github.com/pb33f/jsonpath/pkg/jsonpath/config"
)

type publicSource struct {
	Name       string `json:"name"`
	Repository string `json:"repository"`
	License    string `json:"license"`
	Commit     string `json:"commit"`
	Path       string `json:"path"`
}

type publicSelector struct {
	Expression string   `json:"expression"`
	Sources    []string `json:"sources"`
	Status     string   `json:"status"`
	Reason     string   `json:"reason,omitempty"`
}

type publicInventory struct {
	GeneratedBy string           `json:"generatedBy"`
	Sources     []publicSource   `json:"sources"`
	Selectors   []publicSelector `json:"selectors"`
}

var publicSources = []publicSource{
	{
		Name:       "Harbor",
		Repository: "https://github.com/goharbor/harbor",
		License:    "Apache-2.0",
		Commit:     "691eb2fe36f4043f040c5d18ef365511042a917b",
		Path:       ".spectral.yaml",
	},
	{
		Name:       "Kibana",
		Repository: "https://github.com/elastic/kibana",
		License:    "NOASSERTION",
		Commit:     "aa6146fd7a49bf4291921c51b35bce3a7417dc01",
		Path:       "oas_docs/linters/.spectral.yaml",
	},
	{
		Name:       "Postman Knowledge Base",
		Repository: "https://github.com/postman-open-technologies/knowledge-base",
		License:    "Apache-2.0",
		Commit:     "2dd6098ab1d63dd9dc4ea52f9c6f6dc9c64fbab2",
		Path:       "spectral/postman/http-status-codes.spectral.yaml",
	},
	{
		Name:       "Cimpress API Style Guide",
		Repository: "https://github.com/Cimpress/cimpress-api-style-guide",
		License:    "NOASSERTION",
		Commit:     "873a6aa7d37933a3c7a0f8bfd7a6018b63c5b3a7",
		Path:       "spectral/cimpress.spectral.yaml",
	},
}

func main() {
	client := &http.Client{Timeout: 30 * time.Second}
	selectors := make(map[string]map[string]struct{})
	for _, source := range publicSources {
		document, err := fetchPublicRuleset(client, source)
		if err != nil {
			fatal(err)
		}
		var expressions []string
		collectGivenSelectors(&document, &expressions)
		aliases := collectAliasDefinitions(&document)
		if len(expressions) == 0 {
			fatal(fmt.Errorf("%s contains no given selectors", source.Name))
		}
		for _, expression := range expressions {
			for _, expanded := range expandAliasExpression(expression, aliases, nil) {
				if selectors[expanded] == nil {
					selectors[expanded] = make(map[string]struct{})
				}
				selectors[expanded][source.Name] = struct{}{}
			}
		}
	}

	inventory := publicInventory{
		GeneratedBy: "npm run collect-public",
		Sources:     publicSources,
		Selectors:   make([]publicSelector, 0, len(selectors)),
	}
	for expression, sourceNames := range selectors {
		entry := publicSelector{Expression: expression, Status: "supported"}
		for sourceName := range sourceNames {
			entry.Sources = append(entry.Sources, sourceName)
		}
		sort.Strings(entry.Sources)
		if _, err := jsonpath.NewPath(expression, config.WithSpectralCompatibility()); err != nil {
			entry.Status = "unsupported"
			entry.Reason = err.Error()
		}
		inventory.Selectors = append(inventory.Selectors, entry)
	}
	sort.Slice(inventory.Selectors, func(i, j int) bool {
		return inventory.Selectors[i].Expression < inventory.Selectors[j].Expression
	})

	encoded, err := json.MarshalIndent(inventory, "", "  ")
	if err != nil {
		fatal(fmt.Errorf("encode inventory: %w", err))
	}
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		fatal(fmt.Errorf("resolve collector path"))
	}
	output := filepath.Join(filepath.Dir(filename), "..", "public-selectors.json")
	if err := os.WriteFile(output, append(encoded, '\n'), 0o644); err != nil {
		fatal(fmt.Errorf("write inventory: %w", err))
	}
}

func fetchPublicRuleset(client *http.Client, source publicSource) (yaml.Node, error) {
	repository := strings.TrimPrefix(source.Repository, "https://github.com/")
	url := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s", repository, source.Commit, source.Path)
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return yaml.Node{}, fmt.Errorf("build %s request: %w", source.Name, err)
	}
	request.Header.Set("User-Agent", "pb33f-jsonpath-spectral-corpus")
	response, err := client.Do(request)
	if err != nil {
		return yaml.Node{}, fmt.Errorf("fetch %s: %w", source.Name, err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return yaml.Node{}, fmt.Errorf("fetch %s: %s", source.Name, response.Status)
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return yaml.Node{}, fmt.Errorf("read %s: %w", source.Name, err)
	}
	var document yaml.Node
	if err := yaml.Unmarshal(body, &document); err != nil {
		return yaml.Node{}, fmt.Errorf("parse %s: %w", source.Name, err)
	}
	return document, nil
}

func collectGivenSelectors(node *yaml.Node, selectors *[]string) {
	if node == nil {
		return
	}
	if node.Kind == yaml.MappingNode {
		for i := 0; i+1 < len(node.Content); i += 2 {
			key, value := node.Content[i], node.Content[i+1]
			if key.Value == "given" {
				collectGivenValue(value, selectors)
			}
			collectGivenSelectors(value, selectors)
		}
		return
	}
	for _, child := range node.Content {
		collectGivenSelectors(child, selectors)
	}
}

func collectGivenValue(node *yaml.Node, selectors *[]string) {
	switch node.Kind {
	case yaml.ScalarNode:
		if node.Value != "" {
			*selectors = append(*selectors, node.Value)
		}
	case yaml.SequenceNode:
		for _, child := range node.Content {
			if child.Kind == yaml.ScalarNode && child.Value != "" {
				*selectors = append(*selectors, child.Value)
			}
		}
	}
}

func collectAliasDefinitions(document *yaml.Node) map[string][]string {
	aliases := make(map[string][]string)
	root := document
	if root.Kind == yaml.DocumentNode && len(root.Content) == 1 {
		root = root.Content[0]
	}
	if root.Kind != yaml.MappingNode {
		return aliases
	}
	for i := 0; i+1 < len(root.Content); i += 2 {
		if root.Content[i].Value != "aliases" || root.Content[i+1].Kind != yaml.MappingNode {
			continue
		}
		definitions := root.Content[i+1]
		for j := 0; j+1 < len(definitions.Content); j += 2 {
			name, value := definitions.Content[j].Value, definitions.Content[j+1]
			var expressions []string
			collectGivenSelectors(value, &expressions)
			collectAliasScalarExpressions(value, &expressions)
			aliases[name] = uniqueStrings(expressions)
		}
	}
	return aliases
}

func collectAliasScalarExpressions(node *yaml.Node, expressions *[]string) {
	if node == nil {
		return
	}
	if node.Kind == yaml.ScalarNode && (strings.HasPrefix(node.Value, "$") || strings.HasPrefix(node.Value, "#")) {
		*expressions = append(*expressions, node.Value)
		return
	}
	for _, child := range node.Content {
		collectAliasScalarExpressions(child, expressions)
	}
}

func expandAliasExpression(expression string, aliases map[string][]string, active map[string]bool) []string {
	if !strings.HasPrefix(expression, "#") {
		return []string{expression}
	}
	end := strings.IndexAny(expression, ".[ ")
	if end < 0 {
		end = len(expression)
	}
	name := expression[1:end]
	definitions := aliases[name]
	if len(definitions) == 0 {
		return []string{expression}
	}
	if active == nil {
		active = make(map[string]bool)
	}
	if active[name] {
		return []string{expression}
	}
	active[name] = true
	defer delete(active, name)
	suffix := expression[end:]
	var expanded []string
	for _, definition := range definitions {
		for _, base := range expandAliasExpression(definition, aliases, active) {
			expanded = append(expanded, base+suffix)
		}
	}
	return uniqueStrings(expanded)
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
