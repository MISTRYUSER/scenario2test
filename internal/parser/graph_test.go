package parser

import (
	"testing"

	"scenario2test/internal/path"
)

func TestBuildGraphSupportsFlexibleDSLBranches(t *testing.T) {
	scenario := Scenario{
		Name: "checkout",
		Steps: []Step{
			{ID: "open_home", Action: "open_page", Target: "https://mall.com", Next: "check_auth"},
			{
				ID:     "check_auth",
				Action: "conditional_branch",
				Branches: []StepBranch{
					{Condition: "auth == 'unlogged'", Next: "user_login"},
					{Condition: "auth == 'logged'", Next: "search_item"},
				},
			},
			{ID: "user_login", Action: "user_login", Next: "search_item"},
			{ID: "search_item", Action: "search_item", Next: "add_to_cart"},
			{ID: "add_to_cart", Action: "add_to_cart", Next: "terminal_confirm"},
			{ID: "terminal_confirm", Action: "terminal_confirm", Type: "end"},
		},
	}

	graph, err := BuildGraph(Normalize(scenario))
	if err != nil {
		t.Fatalf("BuildGraph returned error: %v", err)
	}

	paths := path.EnumerateDFS(graph)
	signatures := make([]string, 0, len(paths))
	for _, item := range paths {
		signatures = append(signatures, path.Signature(item))
	}
	if len(paths) != 2 {
		t.Fatalf("expected 2 execution paths, got %d: %v", len(paths), signatures)
	}
	if signatures[0] == signatures[1] {
		t.Fatalf("expected branch-specific signatures, got %v", signatures)
	}
}
