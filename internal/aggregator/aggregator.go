package aggregator

import (
	"encoding/json"

	"scenario2test/internal/generator/e2e"
	"scenario2test/internal/graph"
	"scenario2test/internal/path"
	"scenario2test/internal/strategy"
)

type Result struct {
	Scenario        string               `json:"scenario"`
	Graph           *graph.Graph         `json:"graph,omitempty"`
	Paths           []path.ExecutionPath `json:"paths"`
	TestCases       []strategy.TestCase  `json:"test_cases"`
	SeleniumScripts []e2e.Suite          `json:"selenium_scripts"`
}

func (r Result) JSON() ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}

type Builder struct {
	result Result
}

func New() Builder {
	return Builder{}
}

func (b Builder) WithScenario(name string) Builder {
	b.result.Scenario = name
	return b
}

func (b Builder) WithGraph(graph *graph.Graph) Builder {
	b.result.Graph = graph
	return b
}

func (b Builder) WithPaths(paths []path.ExecutionPath) Builder {
	b.result.Paths = paths
	return b
}

func (b Builder) WithTestCases(testCases []strategy.TestCase) Builder {
	b.result.TestCases = testCases
	return b
}

func (b Builder) WithE2ESuites(suites []e2e.Suite) Builder {
	b.result.SeleniumScripts = suites
	return b
}

func (b Builder) Build() Result {
	return b.result
}
