package parser

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

func LoadScenarioFile(path string) (Scenario, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Scenario{}, err
	}

	return LoadScenarioBytes(raw)
}

func LoadScenarioBytes(raw []byte) (Scenario, error) {
	var file ScenarioFile
	if err := yaml.Unmarshal(raw, &file); err != nil {
		return Scenario{}, fmt.Errorf("unmarshal yaml: %w", err)
	}

	if err := Validate(file.Scenario); err != nil {
		return Scenario{}, err
	}

	return Normalize(file.Scenario), nil
}

func Normalize(s Scenario) Scenario {
	for i := range s.Steps {
		if s.Steps[i].ID == "" {
			s.Steps[i].ID = stepID(i)
		}
		s.Steps[i].Action = normalizeAction(s.Steps[i].Action)
		if s.Steps[i].Target == "" {
			if target := stringParam(s.Steps[i].Params, "url"); target != "" {
				s.Steps[i].Target = target
			}
		}
		if s.Steps[i].Field == "" {
			if field := stringParam(s.Steps[i].Params, "field"); field != "" {
				s.Steps[i].Field = field
			}
		}
		if s.Steps[i].Value == "" {
			if value := stringParam(s.Steps[i].Params, "value"); value != "" {
				s.Steps[i].Value = value
			}
		}
	}
	return s
}

func Validate(s Scenario) error {
	if s.Name == "" {
		return fmt.Errorf("scenario.name is required")
	}
	if len(s.Steps) == 0 {
		return fmt.Errorf("scenario.steps cannot be empty")
	}
	for i, step := range s.Steps {
		if step.Action == "" {
			return fmt.Errorf("scenario.steps[%d].action is required", i)
		}
	}
	return nil
}

func stepID(index int) string {
	return fmt.Sprintf("step_%02d", index+1)
}

func normalizeAction(action string) string {
	switch strings.TrimSpace(action) {
	case "open_url":
		return "open_page"
	default:
		return strings.TrimSpace(action)
	}
}

func stringParam(params map[string]interface{}, key string) string {
	if len(params) == 0 {
		return ""
	}
	value, ok := params[key]
	if !ok {
		return ""
	}
	text, _ := value.(string)
	return strings.TrimSpace(text)
}
