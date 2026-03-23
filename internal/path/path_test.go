package path

import (
	"testing"

	"scenario2test/internal/graph"
)

func TestEnumerateDFSWithBranch(t *testing.T) {
	g := graph.New("branching")
	start := graph.NewNode("start", graph.NodeTypeStart, nil)
	first := graph.NewNode("first", graph.NodeTypeAction, map[string]interface{}{"action": "open_page"})
	ok := graph.NewNode("ok", graph.NodeTypeAction, map[string]interface{}{"action": "click"})
	fail := graph.NewNode("fail", graph.NodeTypeAction, map[string]interface{}{"action": "click"})
	end := graph.NewNode("end", graph.NodeTypeEnd, nil)

	g.SetStart(start)
	for _, node := range []*graph.Node{start, first, ok, fail, end} {
		g.AddNode(node)
	}

	_ = g.Connect("start", "first", "")
	_ = g.Connect("first", "ok", "approved")
	_ = g.Connect("first", "fail", "rejected")
	_ = g.Connect("ok", "end", "")
	_ = g.Connect("fail", "end", "")

	paths := EnumerateDFS(g)
	if len(paths) != 2 {
		t.Fatalf("expected 2 paths, got %d", len(paths))
	}

	signatures := []string{Signature(paths[0]), Signature(paths[1])}
	if signatures[0] == signatures[1] {
		t.Fatalf("expected unique path signatures, got %v", signatures)
	}
}
