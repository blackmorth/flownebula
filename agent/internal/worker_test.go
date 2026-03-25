package internal

import (
	"encoding/binary"
	"testing"
)

func TestProcessPacketStoresFunctionName(t *testing.T) {
	FuncNamesMu.Lock()
	FuncNames = map[uint32]string{}
	FuncNamesMu.Unlock()

	buf := make([]byte, NameHeaderSize+4)
	binary.LittleEndian.PutUint64(buf[0:8], 42)
	buf[8] = EventFuncName
	binary.LittleEndian.PutUint32(buf[9:13], 77)
	binary.LittleEndian.PutUint32(buf[13:17], 4)
	copy(buf[17:], []byte("main"))

	ProcessPacket(Packet{Data: buf, N: len(buf)})

	FuncNamesMu.RLock()
	got, ok := FuncNames[77]
	FuncNamesMu.RUnlock()
	if !ok {
		t.Fatalf("expected function name to be stored")
	}
	if got != "main" {
		t.Fatalf("expected function name 'main', got %q", got)
	}
}

func TestProcessPacketAddsEventToSession(t *testing.T) {
	for i := 0; i < NumShards; i++ {
		SessionShards[i] = &SessionShard{Sessions: make(map[uint64]*Session)}
	}

	sessID := uint64(12345)
	buf := make([]byte, EventSize)
	binary.LittleEndian.PutUint64(buf[0:8], sessID)
	buf[8] = EventEnter
	binary.LittleEndian.PutUint32(buf[9:13], 9)

	ProcessPacket(Packet{Data: buf, N: len(buf)})

	s := GetSession(sessID)
	if len(s.stack) != 1 {
		t.Fatalf("expected one frame in stack, got %d", len(s.stack))
	}
	if s.stack[0].FuncID != 9 {
		t.Fatalf("expected func id 9 on stack, got %d", s.stack[0].FuncID)
	}
}
