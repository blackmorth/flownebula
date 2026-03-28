package internal

import "testing"

func TestNsToUS(t *testing.T) {
	tests := []struct {
		ns   uint64
		want int64
	}{
		{0, 0},
		{999, 1},
		{1000, 1},
		{1499, 1},
		{1500, 2},
		{10_000_000, 10_000},
	}

	for _, tt := range tests {
		if got := nsToUS(tt.ns); got != tt.want {
			t.Fatalf("nsToUS(%d) = %d, want %d", tt.ns, got, tt.want)
		}
	}
}

func TestAddEventConvertsTimingToMicroseconds(t *testing.T) {
	s := &Session{
		ID:    1,
		Root:  NewNode(0, 1),
		stack: []*Node{},
	}
	node := NewNode(1, 42)
	s.Push(node)

	s.AddEvent(CallEvent{
		Type:      EventExit,
		FuncID:    42,
		Inclusive: 10_000_000, // 10 ms in ns
		Exclusive: 3_000_000,  // 3 ms in ns
		CPUTime:   2_000_000,  // 2 ms in ns
	})

	if got := node.Metrics["ct"]; got != 1 {
		t.Fatalf("ct = %d, want 1", got)
	}
	if got := node.Metrics["wt"]; got != 10_000 {
		t.Fatalf("wt = %d, want 10000 (us)", got)
	}
	if got := node.Metrics["ewt"]; got != 3_000 {
		t.Fatalf("ewt = %d, want 3000 (us)", got)
	}
	if got := node.Metrics["cpu"]; got != 2_000 {
		t.Fatalf("cpu = %d, want 2000 (us)", got)
	}
}

func TestSessionEndCarriesDroppedEventsAndProtocol(t *testing.T) {
	s := &Session{
		ID:    2,
		Root:  NewNode(0, 1),
		stack: []*Node{},
	}

	s.AddEvent(CallEvent{
		Type:    EventSessionEnd,
		IOWait:  123,
		Network: 1,
	})

	if !s.Closed {
		t.Fatal("session should be closed after EventSessionEnd")
	}
	if s.Dropped != 123 {
		t.Fatalf("dropped = %d, want 123", s.Dropped)
	}
	if s.Protocol != 1 {
		t.Fatalf("protocol = %d, want 1", s.Protocol)
	}
}
