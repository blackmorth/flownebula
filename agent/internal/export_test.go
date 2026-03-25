package internal

import (
	"encoding/json"
	"testing"
)

func TestParseNodeNameVariants(t *testing.T) {
	tests := []struct {
		name       string
		raw        string
		nodeID     string
		display    string
		calledClas string
	}{
		{
			name:       "closure",
			raw:        "{closure}::index.php/12",
			nodeID:     "{closure}::index.php/12",
			display:    "",
			calledClas: "",
		},
		{
			name:       "class method",
			raw:        "index.php::Service::Run",
			nodeID:     "Service::Run",
			display:    "Service::Run",
			calledClas: "Service",
		},
		{
			name:       "script level unknown function",
			raw:        "path/to/index.php::Main::unknown",
			nodeID:     "run_init::index.php",
			display:    "",
			calledClas: "",
		},
		{
			name:       "global function",
			raw:        "index.php::::strtolower",
			nodeID:     "strtolower",
			display:    "",
			calledClas: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nodeID, display, cls := parseNodeName(tt.raw)
			if nodeID != tt.nodeID || display != tt.display || cls != tt.calledClas {
				t.Fatalf("parseNodeName(%q) = (%q,%q,%q), expected (%q,%q,%q)", tt.raw, nodeID, display, cls, tt.nodeID, tt.display, tt.calledClas)
			}
		})
	}
}

func TestRound3(t *testing.T) {
	cases := []struct {
		in   float64
		want float64
	}{
		{0, 0},
		{12.3456, 12.3},
		{9999.9, 10000},
		{0.012345, 0.0123},
	}

	for _, tc := range cases {
		got := round3(tc.in)
		if got != tc.want {
			t.Fatalf("round3(%v) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

func TestExportToDetailedJSONKeepsChildrenWhenRootSelfEdgeIsSkipped(t *testing.T) {
	FuncNamesMu.Lock()
	FuncNames = map[uint32]string{
		1: "test.php",
		2: "foo",
	}
	FuncNamesMu.Unlock()

	s := &Session{
		ID:   0x42,
		Root: NewNode(0, 1),
	}

	// Reproduce shape created by RINIT + first ENTER:
	// root(1) -> child(1) -> foo(2)
	dupRoot := NewNode(1, 1)
	foo := NewNode(2, 2)
	foo.Metrics["ct"] = 1
	foo.Metrics["wt"] = 10

	dupRoot.Children[2] = foo
	s.Root.Children[1] = dupRoot

	raw, err := s.ExportToDetailedJSON()
	if err != nil {
		t.Fatalf("ExportToDetailedJSON failed: %v", err)
	}

	var out DetailedJSON
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if _, ok := out.Nodes["foo"]; !ok {
		t.Fatalf("expected foo node to be present, got nodes: %#v", out.Nodes)
	}

	foundEdge := false
	for _, e := range out.Edges {
		if e.Caller == "test.php" && e.Callee == "foo" {
			foundEdge = true
			break
		}
	}
	if !foundEdge {
		t.Fatalf("expected edge test.php -> foo, got edges: %#v", out.Edges)
	}
}

func TestExportToDetailedJSONKeepsRecursiveSelfEdge(t *testing.T) {
	FuncNamesMu.Lock()
	FuncNames = map[uint32]string{
		1: "test.php",
		2: "bar",
	}
	FuncNamesMu.Unlock()

	s := &Session{
		ID:   0x43,
		Root: NewNode(0, 1),
	}

	// test.php -> bar -> bar -> bar
	bar1 := NewNode(1, 2)
	bar1.Metrics["ct"] = 1
	bar1.Metrics["wt"] = 10

	bar2 := NewNode(2, 2)
	bar2.Metrics["ct"] = 1
	bar2.Metrics["wt"] = 20

	bar3 := NewNode(3, 2)
	bar3.Metrics["ct"] = 1
	bar3.Metrics["wt"] = 30

	bar2.Children[2] = bar3
	bar1.Children[2] = bar2
	s.Root.Children[2] = bar1

	raw, err := s.ExportToDetailedJSON()
	if err != nil {
		t.Fatalf("ExportToDetailedJSON failed: %v", err)
	}

	var out DetailedJSON
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	var rootToBarCt, barToBarCt int64
	for _, e := range out.Edges {
		if e.Caller == "test.php" && e.Callee == "bar" {
			rootToBarCt += e.Cost["ct"]
		}
		if e.Caller == "bar" && e.Callee == "bar" {
			barToBarCt += e.Cost["ct"]
		}
	}

	if rootToBarCt != 1 {
		t.Fatalf("expected edge test.php -> bar ct=1, got %d (edges: %#v)", rootToBarCt, out.Edges)
	}
	if barToBarCt != 2 {
		t.Fatalf("expected recursive edge bar -> bar ct=2, got %d (edges: %#v)", barToBarCt, out.Edges)
	}
}

func TestExportToDetailedJSONSkipsSyntheticRootSelfEdgeWithDifferentFuncID(t *testing.T) {
	FuncNamesMu.Lock()
	FuncNames = map[uint32]string{
		1: "test.php",
		2: "test.php",
		3: "foo",
	}
	FuncNamesMu.Unlock()

	s := &Session{
		ID:   0x44,
		Root: NewNode(0, 1),
	}

	// Synthetic request entry and script-level frame share the same rendered name.
	entryFrame := NewNode(1, 2)
	entryFrame.Metrics["ct"] = 1
	entryFrame.Metrics["wt"] = 11

	foo := NewNode(2, 3)
	foo.Metrics["ct"] = 1
	foo.Metrics["wt"] = 7

	entryFrame.Children[3] = foo
	s.Root.Children[2] = entryFrame

	raw, err := s.ExportToDetailedJSON()
	if err != nil {
		t.Fatalf("ExportToDetailedJSON failed: %v", err)
	}

	var out DetailedJSON
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	for _, e := range out.Edges {
		if e.Caller == "test.php" && e.Callee == "test.php" {
			t.Fatalf("unexpected synthetic self edge test.php -> test.php in edges: %#v", out.Edges)
		}
	}

	foundFooEdge := false
	for _, e := range out.Edges {
		if e.Caller == "test.php" && e.Callee == "foo" {
			foundFooEdge = true
			break
		}
	}
	if !foundFooEdge {
		t.Fatalf("expected edge test.php -> foo, got edges: %#v", out.Edges)
	}
}

func TestExportToDetailedJSONKeepsEdgeToRootNamedNodeWhenNotDirectChildOfRoot(t *testing.T) {
	FuncNamesMu.Lock()
	FuncNames = map[uint32]string{
		1: "test.php",
		2: "worker.php",
		3: "test.php",
	}
	FuncNamesMu.Unlock()

	s := &Session{
		ID:   0x45,
		Root: NewNode(0, 1),
	}

	worker := NewNode(1, 2)
	worker.Metrics["ct"] = 1
	worker.Metrics["wt"] = 5

	backToRootName := NewNode(2, 3)
	backToRootName.Metrics["ct"] = 1
	backToRootName.Metrics["wt"] = 6

	worker.Children[3] = backToRootName
	s.Root.Children[2] = worker

	raw, err := s.ExportToDetailedJSON()
	if err != nil {
		t.Fatalf("ExportToDetailedJSON failed: %v", err)
	}

	var out DetailedJSON
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	foundEdge := false
	for _, e := range out.Edges {
		if e.Caller == "worker.php" && e.Callee == "test.php" {
			foundEdge = true
			break
		}
	}
	if !foundEdge {
		t.Fatalf("expected edge worker.php -> test.php, got edges: %#v", out.Edges)
	}
}
