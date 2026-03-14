package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

type Edge struct {
	Caller   string `json:"caller"`
	Callee   string `json:"callee"`
	Calls    int    `json:"calls"`
	Time     int64  `json:"time"`
	MemTotal int64  `json:"mem_total"`
}

func main() {
	addr := getenvDefault("FLOWNEBULA_AGENT_ADDR", "127.0.0.1:8135")
	outPath := getenvDefault("FLOWNEBULA_AGENT_OUT", "nebula.json")

	log.Printf("FlowNebula agent listening on %s, writing to %s\n", addr, outPath)

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("cannot listen on %s: %v", addr, err)
	}
	defer ln.Close()

	edges := make(map[string]*Edge)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("accept error: %v", err)
			continue
		}

		log.Printf("new connection from %s", conn.RemoteAddr())
		if err := handleConn(conn, edges); err != nil {
			log.Printf("connection error: %v", err)
		}

		if err := writeGraph(outPath, edges); err != nil {
			log.Printf("writeGraph error: %v", err)
		}
	}
}

func handleConn(conn net.Conn, edges map[string]*Edge) error {
	defer conn.Close()

	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 3 {
			log.Printf("invalid line (expected at least 3 fields): %q", line)
			continue
		}

		caller := parts[0]
		callee := parts[1]
		durStr := parts[2]
		memStr := ""
		if len(parts) >= 4 {
			memStr = parts[3]
		}

		dur, err := strconv.ParseInt(durStr, 10, 64)
		if err != nil {
			log.Printf("invalid duration %q in line %q: %v", durStr, line, err)
			continue
		}

		var memDelta int64
		if memStr != "" {
			memDelta, err = strconv.ParseInt(memStr, 10, 64)
			if err != nil {
				log.Printf("invalid mem %q in line %q: %v", memStr, line, err)
				memDelta = 0
			}
		}

		key := caller + "->" + callee

		e, ok := edges[key]
		if !ok {
			e = &Edge{
				Caller:   caller,
				Callee:   callee,
				Calls:    0,
				Time:     0,
				MemTotal: 0,
			}
			edges[key] = e
		}

		e.Calls++
		e.Time += dur
		e.MemTotal += memDelta
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scanner error: %w", err)
	}

	return nil
}

func writeGraph(path string, edges map[string]*Edge) error {
	values := make([]*Edge, 0, len(edges))
	for _, e := range edges {
		values = append(values, e)
	}

	data, err := json.MarshalIndent(values, "", "  ")
	if err != nil {
		return fmt.Errorf("json marshal: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	log.Printf("graph written to %s (%d edges)", path, len(values))
	return nil
}

func getenvDefault(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

