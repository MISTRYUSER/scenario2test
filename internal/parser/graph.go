package parser

import (
	"fmt"
	"strings"

	"scenario2test/internal/graph"
)

func BuildGraph(s Scenario) (*graph.Graph, error) {
	g := graph.New(s.Name)
	start := graph.NewNode("start", graph.NodeTypeStart, map[string]interface{}{
		"scenario": s.Name,
	})
	g.SetStart(start)
	g.AddNode(start)

	previous := start
	for _, step := range s.Steps {
		nodeType := graph.NodeTypeAction
		if strings.EqualFold(step.Type, "end") {
			nodeType = graph.NodeTypeEnd
		}
		node := graph.NewNode(step.ID, graph.NodeTypeAction, map[string]interface{}{
			"action":   step.Action,
			"name":     step.Name,
			"target":   step.Target,
			"field":    step.Field,
			"value":    step.Value,
			"expected": step.Expected,
			"tags":     step.Tags,
			"params":   step.Params,
			"type":     step.Type,
		})
		node.Type = nodeType
		g.AddNode(node)
		if previous != nil {
			g.Connect(previous.ID, node.ID, "")
		}
		previous = node
	}

	end := graph.NewNode("end", graph.NodeTypeEnd, map[string]interface{}{
		"assertions": s.Assertions,
	})
	g.AddNode(end)
	g.Connect(previous.ID, end.ID, "")

	explicitTopology := false
	for _, step := range s.Steps {
		if step.Next != "" || len(step.Branches) > 0 {
			explicitTopology = true
			break
		}
	}

	if explicitTopology {
		start.Next = nil
		start.Outgoing = nil
		for _, node := range g.Nodes {
			if node.ID == end.ID {
				continue
			}
			node.Next = nil
			node.Outgoing = nil
		}
		if len(s.Steps) > 0 {
			_ = g.Connect(start.ID, s.Steps[0].ID, "")
		}
		for i, step := range s.Steps {
			if len(step.Branches) > 0 {
				for _, branch := range step.Branches {
					if targetID, ok := resolveStepReference(branch.Next, s.Steps, i); ok {
						_ = g.Connect(step.ID, targetID, branch.Condition)
					}
				}
			}
			if step.Next != "" {
				if targetID, ok := resolveStepReference(step.Next, s.Steps, i); ok {
					_ = g.Connect(step.ID, targetID, "")
				}
			} else if len(step.Branches) == 0 && i+1 < len(s.Steps) {
				_ = g.Connect(step.ID, s.Steps[i+1].ID, "")
			}
		}
	}

	for _, branch := range s.Branches {
		if !g.HasNode(branch.From) {
			return nil, fmt.Errorf("branch.from %q not found", branch.From)
		}
		if !g.HasNode(branch.To) {
			return nil, fmt.Errorf("branch.to %q not found", branch.To)
		}
		g.Connect(branch.From, branch.To, branch.Condition)
	}

	for _, step := range s.Steps {
		node := g.Nodes[step.ID]
		if node == nil {
			continue
		}
		if len(node.Outgoing) == 0 {
			_ = g.Connect(node.ID, end.ID, "")
		}
	}

	return g, nil
}

func resolveStepReference(ref string, steps []Step, index int) (string, bool) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		if index+1 < len(steps) {
			return steps[index+1].ID, true
		}
		return "", false
	}

	for _, step := range steps {
		if step.ID == ref {
			return step.ID, true
		}
	}

	normalizedRef := normalizeAction(ref)
	for _, step := range steps {
		if step.Action == normalizedRef {
			return step.ID, true
		}
	}

	if index+1 < len(steps) {
		return steps[index+1].ID, true
	}
	return "", false
}
