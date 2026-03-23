package strategy

import (
	"testing"

	"scenario2test/internal/path"
)

func TestDefaultEngineGeneratesCoreStrategies(t *testing.T) {
	engine := NewDefaultEngine()
	paths := []path.ExecutionPath{
		{
			ID: "path_01",
			Steps: []path.StepRef{
				{NodeID: "start", NodeType: "start", Payload: map[string]interface{}{"scenario": "login flow"}},
				{NodeID: "open_login", NodeType: "action", Payload: map[string]interface{}{"action": "open_page", "target": "/login"}},
				{NodeID: "input_user", NodeType: "action", Payload: map[string]interface{}{"action": "input", "field": "username"}},
				{NodeID: "click_login", NodeType: "action", Payload: map[string]interface{}{"action": "click", "target": "login_button"}},
			},
		},
	}

	testCases := engine.Generate(paths)
	if len(testCases) != 5 {
		t.Fatalf("expected 5 default strategy cases, got %d", len(testCases))
	}

	expected := map[string]bool{
		HappyPath:    false,
		InvalidInput: false,
		AuthFail:     false,
		Boundary:     false,
		RateLimit:    false,
	}

	for _, testCase := range testCases {
		expected[testCase.Strategy] = true
	}

	for name, seen := range expected {
		if !seen {
			t.Fatalf("expected strategy %q to be generated", name)
		}
	}
}
