package parser

type ScenarioFile struct {
	Scenario Scenario `yaml:"scenario"`
}

type Scenario struct {
	Name        string                 `yaml:"name" json:"name"`
	Description string                 `yaml:"description,omitempty" json:"description,omitempty"`
	Metadata    map[string]string      `yaml:"metadata,omitempty" json:"metadata,omitempty"`
	Steps       []Step                 `yaml:"steps" json:"steps"`
	Branches    []Branch               `yaml:"branches,omitempty" json:"branches,omitempty"`
	Assertions  []Assertion            `yaml:"assertions,omitempty" json:"assertions,omitempty"`
	Targets     map[string]CodeTarget  `yaml:"targets,omitempty" json:"targets,omitempty"`
}

type Step struct {
	ID        string                 `yaml:"id,omitempty" json:"id,omitempty"`
	Name      string                 `yaml:"name,omitempty" json:"name,omitempty"`
	Type      string                 `yaml:"type,omitempty" json:"type,omitempty"`
	Action    string                 `yaml:"action" json:"action"`
	Target    string                 `yaml:"target,omitempty" json:"target,omitempty"`
	Field     string                 `yaml:"field,omitempty" json:"field,omitempty"`
	Value     string                 `yaml:"value,omitempty" json:"value,omitempty"`
	Params    map[string]interface{} `yaml:"params,omitempty" json:"params,omitempty"`
	Next      string                 `yaml:"next,omitempty" json:"next,omitempty"`
	Branches  []StepBranch           `yaml:"branches,omitempty" json:"branches,omitempty"`
	With      map[string]string      `yaml:"with,omitempty" json:"with,omitempty"`
	Expected  string                 `yaml:"expected,omitempty" json:"expected,omitempty"`
	Tags      []string               `yaml:"tags,omitempty" json:"tags,omitempty"`
	Meta      map[string]interface{} `yaml:"meta,omitempty" json:"meta,omitempty"`
}

type StepBranch struct {
	Condition string `yaml:"condition" json:"condition"`
	Next      string `yaml:"next" json:"next"`
}

type Branch struct {
	From      string `yaml:"from" json:"from"`
	Condition string `yaml:"condition" json:"condition"`
	To        string `yaml:"to" json:"to"`
}

type Assertion struct {
	Type   string `yaml:"type" json:"type"`
	Target string `yaml:"target,omitempty" json:"target,omitempty"`
	Value  string `yaml:"value,omitempty" json:"value,omitempty"`
}

type CodeTarget struct {
	Path     string `yaml:"path,omitempty" json:"path,omitempty"`
	Symbol   string `yaml:"symbol,omitempty" json:"symbol,omitempty"`
	Endpoint string `yaml:"endpoint,omitempty" json:"endpoint,omitempty"`
}
