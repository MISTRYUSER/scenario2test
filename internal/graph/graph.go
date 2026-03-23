package graph

import "fmt"

const (
	NodeTypeStart  = "start"
	NodeTypeAction = "action"
	NodeTypeEnd    = "end"
)

type Edge struct {
	To        *Node  `json:"-"`
	ToID      string `json:"to"`
	Condition string `json:"condition,omitempty"`
}

type Node struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"`
	Payload  map[string]interface{} `json:"payload,omitempty"`
	Next     []*Node                `json:"-"`
	Outgoing []Edge                 `json:"outgoing,omitempty"`
}

type Graph struct {
	Name  string           `json:"name"`
	Start *Node            `json:"-"`
	Nodes map[string]*Node `json:"nodes"`
}

func New(name string) *Graph {
	return &Graph{
		Name:  name,
		Nodes: make(map[string]*Node),
	}
}

func NewNode(id, kind string, payload map[string]interface{}) *Node {
	return &Node{
		ID:      id,
		Type:    kind,
		Payload: payload,
	}
}

func (g *Graph) SetStart(node *Node) {
	g.Start = node
}

func (g *Graph) AddNode(node *Node) {
	g.Nodes[node.ID] = node
	if g.Start == nil {
		g.Start = node
	}
}

func (g *Graph) HasNode(id string) bool {
	_, ok := g.Nodes[id]
	return ok
}

func (g *Graph) Connect(fromID, toID, condition string) error {
	from, ok := g.Nodes[fromID]
	if !ok {
		return fmt.Errorf("from node %q not found", fromID)
	}
	to, ok := g.Nodes[toID]
	if !ok {
		return fmt.Errorf("to node %q not found", toID)
	}

	from.Next = append(from.Next, to)
	from.Outgoing = append(from.Outgoing, Edge{
		To:        to,
		ToID:      to.ID,
		Condition: condition,
	})
	return nil
}
