package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
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

	daemon := flag.Bool("daemon", false, "run agent in background")
	flag.BoolVar(daemon, "d", false, "run agent in background (shorthand)")
	flag.Parse()

	if *daemon && os.Getenv("FLOWNEBULA_DAEMONIZED") == "" {
		daemonize()
		return
	}

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

func daemonize() {

	logFile, err := os.OpenFile("flownebula-agent.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("cannot open log file: %v", err)
	}

	cmd := exec.Command(os.Args[0], os.Args[1:]...)
	cmd.Env = append(os.Environ(), "FLOWNEBULA_DAEMONIZED=1")
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.Stdin = nil

	err = cmd.Start()
	if err != nil {
		log.Fatalf("failed to start daemon: %v", err)
	}

	fmt.Printf("FlowNebula agent started in background (PID %d)\n", cmd.Process.Pid)
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
			log.Printf("invalid line: %q", line)
			continue
		}

		caller := parts[0]
		callee := parts[1]

		dur, err := strconv.ParseInt(parts[2], 10, 64)
		if err != nil {
			continue
		}

		var memDelta int64
		if len(parts) >= 4 {
			memDelta, _ = strconv.ParseInt(parts[3], 10, 64)
		}

		key := caller + "->" + callee

		e, ok := edges[key]
		if !ok {
			e = &Edge{
				Caller: caller,
				Callee: callee,
			}
			edges[key] = e
		}

		e.Calls++
		e.Time += dur
		e.MemTotal += memDelta
	}

	return scanner.Err()
}

func writeGraph(path string, edges map[string]*Edge) error {

	values := make([]*Edge, 0, len(edges))
	for _, e := range edges {
		values = append(values, e)
	}

	data, err := json.MarshalIndent(values, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func getenvDefault(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}