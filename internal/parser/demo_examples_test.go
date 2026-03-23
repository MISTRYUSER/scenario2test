package parser

import (
	"path/filepath"
	"testing"

	"scenario2test/internal/path"
)

func TestDemoExamplesLoadAndEnumeratePaths(t *testing.T) {
	t.Parallel()

	examples := []string{
		"login.yaml",
		"signup.yaml",
		"checkout.yaml",
		"password-reset.yaml",
	}

	for _, name := range examples {
		name := name
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			scenario, err := LoadScenarioFile(filepath.Join("..", "..", "examples", name))
			if err != nil {
				t.Fatalf("LoadScenarioFile returned error: %v", err)
			}

			graph, err := BuildGraph(scenario)
			if err != nil {
				t.Fatalf("BuildGraph returned error: %v", err)
			}

			paths := path.EnumerateDFS(graph)
			if len(paths) == 0 {
				t.Fatalf("expected at least one execution path for %s", name)
			}
		})
	}
}
