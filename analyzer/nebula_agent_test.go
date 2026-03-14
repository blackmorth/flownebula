package main

import (
	"encoding/json"
	"net"
	"os"
	"testing"
)

func TestHandleConnAndWriteGraph(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	edges := make(map[string]*Edge)

	// écrivons quelques lignes côté client
	go func() {
		defer client.Close()
		client.Write([]byte("main a 100 10\n"))
		client.Write([]byte("a b 50 5\n"))
		client.Write([]byte("main a 200 20\n"))
	}()

	if err := handleConn(server, edges); err != nil {
		t.Fatalf("handleConn error: %v", err)
	}

	tmp, err := os.CreateTemp("", "nebula_agent_out_*.json")
	if err != nil {
		t.Fatalf("CreateTemp error: %v", err)
	}
	defer os.Remove(tmp.Name())
	tmp.Close()

	if err := writeGraph(tmp.Name(), edges); err != nil {
		t.Fatalf("writeGraph error: %v", err)
	}

	data, err := os.ReadFile(tmp.Name())
	if err != nil {
		t.Fatalf("read output error: %v", err)
	}

	var out []Edge
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("json unmarshal error: %v", err)
	}

	if len(out) != 2 {
		t.Fatalf("expected 2 edges, got %d", len(out))
	}

	m := map[string]Edge{}
	for _, e := range out {
		m[e.Caller+"->"+e.Callee] = e
	}

	assertEdgeGo(t, m, "main->a", 2, 300, 30)
	assertEdgeGo(t, m, "a->b", 1, 50, 5)
}

func assertEdgeGo(t *testing.T, edges map[string]Edge, key string, calls int, time int64, mem int64) {
	e, ok := edges[key]
	if !ok {
		t.Fatalf("missing edge %s", key)
	}

	if e.Calls != calls {
		t.Fatalf("edge %s: expected calls=%d, got %d", key, calls, e.Calls)
	}

	if e.Time != time {
		t.Fatalf("edge %s: expected time=%d, got %d", key, time, e.Time)
	}

	if e.MemTotal != mem {
		t.Fatalf("edge %s: expected mem=%d, got %d", key, mem, e.MemTotal)
	}
}

