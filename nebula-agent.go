package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"sync"
	"time"
)

type CallEvent struct {
	Type      uint8
	FuncID    uint32
	Inclusive uint64
	Exclusive uint64
	MemDelta  int64
}

type Node struct {
	FuncID    uint32
	Calls     uint64
	Inclusive uint64
	Exclusive uint64
	MemDelta  int64
	Children  map[uint32]*Node
}

type Session struct {
	ID       string
	Root     *Node
	stack    []*Node
	LastSeen time.Time
	mu       sync.Mutex
}

var (
	sessions   = make(map[string]*Session)
	sessionsMu sync.Mutex
)

func newNode(funcID uint32) *Node {
	return &Node{
		FuncID:   funcID,
		Children: make(map[uint32]*Node),
	}
}

func getSession(id string) *Session {
	sessionsMu.Lock()
	defer sessionsMu.Unlock()

	if s, ok := sessions[id]; ok {
		s.LastSeen = time.Now()
		return s
	}

	s := &Session{
		ID:       id,
		Root:     newNode(0),
		stack:    []*Node{},
		LastSeen: time.Now(),
	}
	sessions[id] = s
	return s
}

func (s *Session) push(n *Node) {
	s.stack = append(s.stack, n)
}

func (s *Session) pop() *Node {
	if len(s.stack) == 0 {
		return s.Root
	}
	n := s.stack[len(s.stack)-1]
	s.stack = s.stack[:len(s.stack)-1]
	return n
}

func (s *Session) top() *Node {
	if len(s.stack) == 0 {
		return s.Root
	}
	return s.stack[len(s.stack)-1]
}

func (s *Session) addEvent(ev CallEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()

	switch ev.Type {
	case 0: // enter
		parent := s.top()
		child, ok := parent.Children[ev.FuncID]
		if !ok {
			child = newNode(ev.FuncID)
			parent.Children[ev.FuncID] = child
		}
		s.push(child)

	case 1: // exit
		node := s.pop()
		if node.FuncID != ev.FuncID {
			// désync possible, on ne panique pas
			node = s.Root
		}
		node.Calls++
		node.Inclusive += ev.Inclusive
		node.Exclusive += ev.Exclusive
		node.MemDelta += ev.MemDelta
	}
}

func (s *Session) printNode(n *Node, depth int) {
	prefix := ""
	for i := 0; i < depth; i++ {
		prefix += "  "
	}
	if n.FuncID != 0 {
		fmt.Printf("%sFunc %d | Calls=%d | Incl=%d | Excl=%d | Mem=%+d\n",
			prefix, n.FuncID, n.Calls, n.Inclusive, n.Exclusive, n.MemDelta)
	}
	for _, c := range n.Children {
		s.printNode(c, depth+1)
	}
}

func (s *Session) print() {
	s.mu.Lock()
	defer s.mu.Unlock()
	fmt.Printf("\n=== Session %s ===\n", s.ID)
	s.printNode(s.Root, 0)
}

func cleanupSessions() {
	for range time.Tick(30 * time.Second) {
		sessionsMu.Lock()
		for id, s := range sessions {
			if time.Since(s.LastSeen) > 2*time.Minute {
				delete(sessions, id)
			}
		}
		sessionsMu.Unlock()
	}
}

func main() {
	addr := net.UDPAddr{Port: 8135, IP: net.ParseIP("127.0.0.1")}
	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	fmt.Println("Nebula Agent listening on 127.0.0.1:8135")

	go cleanupSessions()

	go func() {
		for range time.Tick(2 * time.Second) {
			sessionsMu.Lock()
			for _, s := range sessions {
				s.print()
			}
			sessionsMu.Unlock()
		}
	}()

	buf := make([]byte, 64)

	for {
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil || n < 37 {
			continue
		}

		sessID := string(buf[:8])
		ev := CallEvent{
			Type:      buf[8],
			FuncID:    binary.LittleEndian.Uint32(buf[9:13]),
			Inclusive: binary.LittleEndian.Uint64(buf[13:21]),
			Exclusive: binary.LittleEndian.Uint64(buf[21:29]),
			MemDelta:  int64(binary.LittleEndian.Uint64(buf[29:37])),
		}

		s := getSession(sessID)
		s.addEvent(ev)
	}
}