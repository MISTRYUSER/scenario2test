package strategy

import (
	"fmt"
	"strings"

	"scenario2test/internal/path"
)

const (
	HappyPath    = "happy_path"
	InvalidInput = "invalid_input"
	AuthFail     = "auth_fail"
	Boundary     = "boundary"
	RateLimit    = "rate_limit"
)

type TestCase struct {
	ID          string                 `json:"id"`
	Strategy    string                 `json:"strategy"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	PathID      string                 `json:"path_id"`
	Severity    string                 `json:"severity"`
	Tags        []string               `json:"tags,omitempty"`
	Inputs      map[string]string      `json:"inputs,omitempty"`
	Expected    []string               `json:"expected,omitempty"`
	Artifacts   map[string]interface{} `json:"artifacts,omitempty"`
}

type Strategy struct {
	Name  string
	Apply func(path.ExecutionPath) []TestCase
}

type Engine struct {
	strategies []Strategy
}

func NewDefaultEngine() Engine {
	return Engine{
		strategies: []Strategy{
			{Name: HappyPath, Apply: applyHappyPath},
			{Name: InvalidInput, Apply: applyInvalidInput},
			{Name: AuthFail, Apply: applyAuthFail},
			{Name: Boundary, Apply: applyBoundary},
			{Name: RateLimit, Apply: applyRateLimit},
		},
	}
}

func (e Engine) Generate(paths []path.ExecutionPath) []TestCase {
	var all []TestCase
	for _, p := range paths {
		for _, strategy := range e.strategies {
			all = append(all, strategy.Apply(p)...)
		}
	}
	return all
}

func applyHappyPath(p path.ExecutionPath) []TestCase {
	return []TestCase{newCase(p, HappyPath, "P1", "Valid scenario execution succeeds", nil, expectedFromPath(p))}
}

func applyInvalidInput(p path.ExecutionPath) []TestCase {
	inputs := map[string]string{}
	found := false
	for _, step := range p.Steps {
		field, _ := step.Payload["field"].(string)
		if field == "" {
			continue
		}
		inputs[field] = "invalid"
		found = true
	}
	if !found {
		return nil
	}

	return []TestCase{newCase(p, InvalidInput, "P1", "Invalid input is rejected with validation feedback", inputs, []string{
		"validation error is shown",
		"submission does not complete",
	})}
}

func applyAuthFail(p path.ExecutionPath) []TestCase {
	if !containsAction(p, "open_page") && !containsTarget(p, "/login") {
		return nil
	}
	return []TestCase{newCase(p, AuthFail, "P0", "Unauthorized or invalid credentials are handled safely", map[string]string{
		"auth_state": "invalid_credentials",
	}, []string{
		"user remains unauthenticated",
		"error message is shown",
		"protected resources remain inaccessible",
	})}
}

func applyBoundary(p path.ExecutionPath) []TestCase {
	if !containsAction(p, "input") {
		return nil
	}
	return []TestCase{newCase(p, Boundary, "P1", "Boundary values are handled without unexpected behavior", map[string]string{
		"boundary": "min,max,empty,oversized",
	}, []string{
		"system enforces field constraints",
		"no crash or silent truncation",
	})}
}

func applyRateLimit(p path.ExecutionPath) []TestCase {
	if !containsAction(p, "click") && !containsAction(p, "submit") {
		return nil
	}
	return []TestCase{newCase(p, RateLimit, "P2", "Repeated requests trigger rate limit or idempotent handling", map[string]string{
		"burst_requests": "20",
	}, []string{
		"server rejects or throttles excess requests",
		"duplicate effects do not occur",
	})}
}

func newCase(p path.ExecutionPath, strategyName, severity, desc string, inputs map[string]string, expected []string) TestCase {
	return TestCase{
		ID:          fmt.Sprintf("%s_%s", strategyName, p.ID),
		Strategy:    strategyName,
		Title:       title(strategyName, p.ID),
		Description: desc,
		PathID:      p.ID,
		Severity:    severity,
		Tags:        []string{strategyName, p.ID},
		Inputs:      inputs,
		Expected:    expected,
		Artifacts: map[string]interface{}{
			"path_signature": path.Signature(p),
		},
	}
}

func containsAction(p path.ExecutionPath, action string) bool {
	for _, step := range p.Steps {
		if value, _ := step.Payload["action"].(string); value == action {
			return true
		}
	}
	return false
}

func containsTarget(p path.ExecutionPath, target string) bool {
	for _, step := range p.Steps {
		if value, _ := step.Payload["target"].(string); value == target {
			return true
		}
	}
	return false
}

func expectedFromPath(p path.ExecutionPath) []string {
	var expected []string
	for _, step := range p.Steps {
		if value, _ := step.Payload["expected"].(string); value != "" {
			expected = append(expected, value)
		}
	}
	if len(expected) == 0 {
		expected = append(expected, "scenario completes successfully")
	}
	return expected
}

func title(strategyName, pathID string) string {
	return strings.ReplaceAll(strategyName, "_", " ") + " :: " + pathID
}
