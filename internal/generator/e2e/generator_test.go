package e2e

import (
	"context"
	"testing"

	"scenario2test/internal/generator/shared"
	"scenario2test/internal/parser"
	"scenario2test/internal/path"
)

type fakeRunner struct {
	command []string
	workdir string
	calls   int
}

func (f *fakeRunner) Run(_ context.Context, workdir string, command []string, _ []string) (shared.CommandResult, error) {
	f.calls++
	f.command = append([]string(nil), command...)
	f.workdir = workdir
	return shared.CommandResult{Stdout: "ok"}, nil
}

func TestRunExternalBuildsAUTOTESTCommand(t *testing.T) {
	runner := &fakeRunner{}
	generator := Generator{
		config: Config{
			Provider: "AUTOTEST",
			Mode:     ModeCLI,
			Command:  []string{"python", "autotest.py", "--url", "{{url}}", "--signature", "{{signature}}"},
			Workdir:  "/tmp/autotest",
			Timeout:  10,
		},
		runner: runner,
	}

	p := path.ExecutionPath{
		ID: "path_01",
		Steps: []path.StepRef{
			{NodeID: "start", NodeType: "start"},
			{NodeID: "open_login", NodeType: "action", Payload: map[string]interface{}{"action": "open_page", "target": "/login"}},
		},
	}

	command, output, err := generator.runExternal(path.Signature(p), "/login", parser.Scenario{})
	if err != nil {
		t.Fatalf("runExternal returned error: %v", err)
	}
	if output != "ok" {
		t.Fatalf("expected output %q, got %q", "ok", output)
	}

	expected := []string{"python", "autotest.py", "--url", "/login", "--signature", "start -> open_login"}
	if len(command) != len(expected) {
		t.Fatalf("expected %d args, got %d", len(expected), len(command))
	}
	for i := range expected {
		if command[i] != expected[i] {
			t.Fatalf("expected arg %d to be %q, got %q", i, expected[i], command[i])
		}
	}
}

func TestRunExternalBuildsAbsoluteURLFromScenarioBaseURL(t *testing.T) {
	runner := &fakeRunner{}
	generator := Generator{
		config: Config{
			Provider: "AUTOTEST",
			Mode:     ModeCLI,
			Command:  []string{"python", "autotest.py", "--url", "{{url}}"},
			Workdir:  "/tmp/autotest",
			Timeout:  10,
		},
		runner: runner,
	}

	command, _, err := generator.runExternal(
		"path",
		"/login",
		parser.Scenario{Metadata: map[string]string{"base_url": "http://127.0.0.1:8080"}},
	)
	if err != nil {
		t.Fatalf("runExternal returned error: %v", err)
	}

	expected := []string{"python", "autotest.py", "--url", "http://127.0.0.1:8080/login"}
	for i := range expected {
		if command[i] != expected[i] {
			t.Fatalf("expected arg %d to be %q, got %q", i, expected[i], command[i])
		}
	}
}

func TestGenerateReusesExternalResultForSameResolvedURL(t *testing.T) {
	runner := &fakeRunner{}
	generator := Generator{
		config: Config{
			Provider: "AUTOTEST",
			Mode:     ModeCLI,
			Command:  []string{"python", "autotest.py", "--url", "{{url}}"},
			Workdir:  "/tmp/autotest",
			Timeout:  10,
		},
		runner: runner,
	}

	paths := []path.ExecutionPath{
		{
			ID: "path_01",
			Steps: []path.StepRef{
				{NodeID: "open_login", NodeType: "action", Payload: map[string]interface{}{"action": "open_page", "target": "/login"}},
			},
		},
		{
			ID: "path_02",
			Steps: []path.StepRef{
				{NodeID: "open_login", NodeType: "action", Payload: map[string]interface{}{"action": "open_page", "target": "/login"}},
			},
		},
	}

	suites := generator.Generate(
		parser.Scenario{Metadata: map[string]string{"base_url": "http://127.0.0.1:8080"}},
		paths,
		nil,
	)

	if len(suites) != 2 {
		t.Fatalf("expected 2 suites, got %d", len(suites))
	}
	if runner.calls != 1 {
		t.Fatalf("expected runner to be called once, got %d", runner.calls)
	}
	if suites[0].CommandOutput != "ok" || suites[1].CommandOutput != "ok" {
		t.Fatalf("expected cached command output to be reused, got %q and %q", suites[0].CommandOutput, suites[1].CommandOutput)
	}
	if suites[0].StartURL != "http://127.0.0.1:8080/login" || suites[1].StartURL != "http://127.0.0.1:8080/login" {
		t.Fatalf("expected suites to expose resolved start url, got %q and %q", suites[0].StartURL, suites[1].StartURL)
	}
}
