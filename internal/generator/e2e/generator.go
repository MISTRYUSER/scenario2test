package e2e

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"scenario2test/internal/generator/shared"
	"scenario2test/internal/parser"
	"scenario2test/internal/path"
	"scenario2test/internal/strategy"
)

const (
	ModeMock = "mock"
	ModeCLI  = "cli"
)

type Config struct {
	Provider string   `json:"provider" yaml:"provider"`
	Mode     string   `json:"mode" yaml:"mode"`
	Command  []string `json:"command,omitempty" yaml:"command,omitempty"`
	Workdir  string   `json:"workdir,omitempty" yaml:"workdir,omitempty"`
	Timeout  int      `json:"timeout,omitempty" yaml:"timeout,omitempty"`
}

type Suite struct {
	Provider      string              `json:"provider"`
	PathID        string              `json:"path_id"`
	StartURL      string              `json:"start_url,omitempty"`
	Cases         []strategy.TestCase `json:"cases"`
	SeleniumDraft string              `json:"selenium_script"`
	Command       []string            `json:"command,omitempty"`
	CommandOutput string              `json:"command_output,omitempty"`
	Metadata      map[string]string   `json:"metadata,omitempty"`
}

type Generator struct {
	config Config
	runner shared.CommandRunner
}

func NewGenerator(cfg Config) Generator {
	return Generator{
		config: cfg,
		runner: shared.ExecRunner{},
	}
}

func (g Generator) Generate(scenario parser.Scenario, paths []path.ExecutionPath, tests []strategy.TestCase) []Suite {
	grouped := groupByPath(tests)
	suites := make([]Suite, 0, len(paths))
	type externalResult struct {
		command []string
		output  string
		err     error
	}
	cache := make(map[string]externalResult)

	for _, p := range paths {
		startURL := detectStartURL(p)
		resolvedURL := resolveStartURL(startURL, scenario)
		result, ok := cache[resolvedURL]
		if !ok {
			command, output, externalErr := g.runExternal(path.Signature(p), startURL, scenario)
			result = externalResult{command: command, output: output, err: externalErr}
			cache[resolvedURL] = result
		}
		metadata := map[string]string{
			"adapter_hint": "Translate scenario path into AUTOTEST crawl/test seed",
			"mode":         g.config.Mode,
		}
		if result.err != nil {
			metadata["external_error"] = result.err.Error()
		}

		suites = append(suites, Suite{
			Provider:      g.config.Provider,
			PathID:        p.ID,
			StartURL:      resolvedURL,
			Cases:         grouped[p.ID],
			SeleniumDraft: draftSelenium(p),
			Command:       result.command,
			CommandOutput: result.output,
			Metadata:      metadata,
		})
	}
	return suites
}

func (g Generator) runExternal(signature, url string, scenario parser.Scenario) ([]string, string, error) {
	if g.config.Mode != ModeCLI || len(g.config.Command) == 0 {
		return nil, "", nil
	}

	url = resolveStartURL(url, scenario)

	args := make([]string, len(g.config.Command))
	outputDir := filepath.Join(os.TempDir(), "scenario2test-autotest")
	_ = os.MkdirAll(outputDir, 0o755)
	for i, item := range g.config.Command {
		item = strings.ReplaceAll(item, "{{url}}", url)
		item = strings.ReplaceAll(item, "{{signature}}", signature)
		item = strings.ReplaceAll(item, "{{output_dir}}", outputDir)
		item = strings.ReplaceAll(item, "{{auth_data_file}}", scenario.Metadata["autotest_auth_data_file"])
		args[i] = item
	}

	timeout := time.Duration(g.config.Timeout) * time.Second
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	result, err := g.runner.Run(ctx, g.config.Workdir, args, nil)
	output := strings.TrimSpace(result.Stdout)
	if stderr := strings.TrimSpace(result.Stderr); stderr != "" {
		if output != "" {
			output += "\n"
		}
		output += stderr
	}
	return args, output, err
}

func resolveStartURL(raw string, scenario parser.Scenario) string {
	if raw == "" {
		return raw
	}
	if strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") {
		return raw
	}

	base := strings.TrimSpace(scenario.Metadata["base_url"])
	if base == "" {
		return raw
	}

	baseURL, err := url.Parse(base)
	if err != nil {
		return raw
	}
	ref, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	return baseURL.ResolveReference(ref).String()
}

func detectStartURL(p path.ExecutionPath) string {
	for _, step := range p.Steps {
		action, _ := step.Payload["action"].(string)
		target, _ := step.Payload["target"].(string)
		if action == "open_page" && target != "" {
			return target
		}
	}
	return ""
}

func draftSelenium(p path.ExecutionPath) string {
	lines := []string{
		"// generated draft for AUTOTEST integration",
		"driver := webdriver.New()",
	}
	for _, step := range p.Steps {
		action, _ := step.Payload["action"].(string)
		target, _ := step.Payload["target"].(string)
		field, _ := step.Payload["field"].(string)
		value, _ := step.Payload["value"].(string)

		switch action {
		case "open_page":
			lines = append(lines, fmt.Sprintf("driver.Get(%q)", target))
		case "input":
			lines = append(lines, fmt.Sprintf("driver.FindElement(%q).SendKeys(%q)", field, value))
		case "click":
			lines = append(lines, fmt.Sprintf("driver.FindElement(%q).Click()", target))
		}
	}
	return joinLines(lines)
}

func groupByPath(tests []strategy.TestCase) map[string][]strategy.TestCase {
	grouped := make(map[string][]strategy.TestCase)
	for _, test := range tests {
		grouped[test.PathID] = append(grouped[test.PathID], test)
	}
	return grouped
}

func joinLines(lines []string) string {
	result := ""
	for i, line := range lines {
		if i > 0 {
			result += "\n"
		}
		result += line
	}
	return result
}
