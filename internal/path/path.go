package path

import (
	"fmt"
	"strings"

	"scenario2test/internal/graph"
)

type StepRef struct {
	NodeID    string                 `json:"node_id"`
	NodeType  string                 `json:"node_type"`
	Condition string                 `json:"condition,omitempty"`
	Payload   map[string]interface{} `json:"payload,omitempty"`
}

type ExecutionPath struct {
	ID    string    `json:"id"`
	Steps []StepRef `json:"steps"`
}

func EnumerateDFS(g *graph.Graph) []ExecutionPath {
	if g == nil || g.Start == nil {
		return nil
	}

	var paths []ExecutionPath
	var visit func(node *graph.Node, current []StepRef, seen map[string]int)

	visit = func(node *graph.Node, current []StepRef, seen map[string]int) {
		if seen[node.ID] >= 2 {
			return
		}
		seen[node.ID]++

		current = append(current, StepRef{
			NodeID:   node.ID,
			NodeType: node.Type,
			Payload:  node.Payload,
		})

		if len(node.Next) == 0 {
			paths = append(paths, ExecutionPath{
				ID:    fmt.Sprintf("path_%02d", len(paths)+1),
				Steps: append([]StepRef(nil), current...),
			})
			return
		}

		for _, edge := range node.Outgoing {
			next := append([]StepRef(nil), current...)
			if edge.Condition != "" && len(next) > 0 {
				next[len(next)-1].Condition = edge.Condition
			}
			nextSeen := cloneSeen(seen)
			visit(edge.To, next, nextSeen)
		}
	}

	visit(g.Start, nil, map[string]int{})
	return dedupe(paths)
}

func EnumerateBFS(g *graph.Graph) []ExecutionPath {
	if g == nil || g.Start == nil {
		return nil
	}

	type state struct {
		node  *graph.Node
		steps []StepRef
		seen  map[string]int
	}

	queue := []state{{node: g.Start, seen: map[string]int{}}}
	var paths []ExecutionPath

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current.seen[current.node.ID] >= 2 {
			continue
		}
		current.seen[current.node.ID]++

		steps := append(current.steps, StepRef{
			NodeID:   current.node.ID,
			NodeType: current.node.Type,
			Payload:  current.node.Payload,
		})

		if len(current.node.Next) == 0 {
			paths = append(paths, ExecutionPath{
				ID:    fmt.Sprintf("path_%02d", len(paths)+1),
				Steps: steps,
			})
			continue
		}

		for _, edge := range current.node.Outgoing {
			next := append([]StepRef(nil), steps...)
			if edge.Condition != "" && len(next) > 0 {
				next[len(next)-1].Condition = edge.Condition
			}
			queue = append(queue, state{
				node:  edge.To,
				steps: next,
				seen:  cloneSeen(current.seen),
			})
		}
	}

	return dedupe(paths)
}

func dedupe(paths []ExecutionPath) []ExecutionPath {
	seen := make(map[string]struct{}, len(paths))
	result := make([]ExecutionPath, 0, len(paths))

	for _, path := range paths {
		key := ""
		for _, step := range path.Steps {
			key += step.NodeID + "|" + step.Condition + ";"
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, path)
	}

	return result
}

func cloneSeen(input map[string]int) map[string]int {
	cloned := make(map[string]int, len(input))
	for key, value := range input {
		cloned[key] = value
	}
	return cloned
}

func Signature(path ExecutionPath) string {
	parts := make([]string, 0, len(path.Steps))
	for _, step := range path.Steps {
		label := step.NodeID
		if step.Condition != "" {
			label += "[" + step.Condition + "]"
		}
		parts = append(parts, label)
	}
	return strings.Join(parts, " -> ")
}
