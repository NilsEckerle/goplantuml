package goldentests

import (
	"bytes"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"testing"

	goplantuml "github.com/jfeliu007/goplantuml/parser"
)

// update regenerates the golden .puml files instead of comparing against them.
// Run: go test ./goldentests -run TestGolden -update
var update = flag.Bool("update", false, "update golden .puml files")

// caseConfig mirrors the CLI flags in cmd/goplantuml/main.go that affect
// rendering. JSON keys match the CLI flag names (hyphenated) so config.json
// reads like the command line. Every field is optional; the zero value
// reproduces a plain `goplantuml <dir>` invocation.
type caseConfig struct {
	Recursive               bool     `json:"recursive"`
	Ignore                  []string `json:"ignore"`
	MaxDepth                int      `json:"max-depth"`
	ShowAggregations        bool     `json:"show-aggregations"`
	HideFields              bool     `json:"hide-fields"`
	HideMethods             bool     `json:"hide-methods"`
	HideConnections         bool     `json:"hide-connections"`
	ShowCompositions        bool     `json:"show-compositions"`
	ShowImplementations     bool     `json:"show-implementations"`
	ShowAliases             bool     `json:"show-aliases"`
	ShowConnectionLabels    bool     `json:"show-connection-labels"`
	AggregatePrivateMembers bool     `json:"aggregate-private-members"`
	HidePrivateMembers      bool     `json:"hide-private-members"`
	Title                   string   `json:"title"`
}

// renderingOptions mirrors the option-building logic in main.go exactly,
// including the rule that compositions/implementations/aliases only take
// effect when connections are hidden.
func (c caseConfig) renderingOptions() map[goplantuml.RenderingOption]interface{} {
	ro := map[goplantuml.RenderingOption]interface{}{
		goplantuml.RenderConnectionLabels:  c.ShowConnectionLabels,
		goplantuml.RenderFields:            !c.HideFields,
		goplantuml.RenderMethods:           !c.HideMethods,
		goplantuml.RenderAggregations:      c.ShowAggregations,
		goplantuml.RenderTitle:             c.Title,
		goplantuml.AggregatePrivateMembers: c.AggregatePrivateMembers,
		goplantuml.RenderPrivateMembers:    !c.HidePrivateMembers,
	}
	if c.HideConnections {
		ro[goplantuml.RenderAliases] = c.ShowAliases
		ro[goplantuml.RenderCompositions] = c.ShowCompositions
		ro[goplantuml.RenderImplementations] = c.ShowImplementations
	}
	return ro
}

// loadConfig reads config.json from a case dir; missing file means defaults.
// Unknown keys are rejected so a typo'd flag fails loudly instead of being
// silently ignored.
func loadConfig(t *testing.T, caseDir string) caseConfig {
	t.Helper()
	cfg := caseConfig{}
	data, err := os.ReadFile(filepath.Join(caseDir, "config.json"))
	if os.IsNotExist(err) {
		return cfg
	}
	if err != nil {
		t.Fatalf("reading config.json: %v", err)
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&cfg); err != nil {
		t.Fatalf("parsing config.json (check for unknown/misspelled keys): %v", err)
	}
	return cfg
}

// TestGolden discovers every directory under cases/ that contains an `input`
// subfolder, parses input/, and compares the rendered diagram to
// expected.puml. To add an edge case: create cases/<name>/input/ with .go
// files, optionally a cases/<name>/config.json, then run with -update once.
func TestGolden(t *testing.T) {
	casesRoot := "cases"
	entries, err := os.ReadDir(casesRoot)
	if err != nil {
		t.Fatalf("reading cases dir: %v", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		caseDir := filepath.Join(casesRoot, name)
		inputDir := filepath.Join(caseDir, "input")
		if fi, err := os.Stat(inputDir); err != nil || !fi.IsDir() {
			continue // not a test case
		}

		t.Run(name, func(t *testing.T) {
			cfg := loadConfig(t, caseDir)

			absInput, err := filepath.Abs(inputDir)
			if err != nil {
				t.Fatalf("abs path: %v", err)
			}

			result, err := goplantuml.NewClassDiagramWithMaxDepth(
				[]string{absInput}, cfg.Ignore, cfg.Recursive, cfg.MaxDepth,
			)
			if err != nil {
				t.Fatalf("NewClassDiagram: %v", err)
			}
			result.SetRenderingOptions(cfg.renderingOptions())
			got := result.Render()

			goldenPath := filepath.Join(caseDir, "expected.puml")
			if *update {
				if err := os.WriteFile(goldenPath, []byte(got), 0o644); err != nil {
					t.Fatalf("writing golden: %v", err)
				}
				t.Logf("updated %s", goldenPath)
				return
			}

			want, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("reading golden (run with -update to create): %v", err)
			}
			if got != string(want) {
				t.Errorf("output mismatch for case %q.\nRun: go test ./goldentests -run TestGolden -update\nthen inspect the diff with git.", name)
			}
		})
	}
}
