package protocol

import (
	"encoding/binary"
)

const (
	ProtocolVersion = 2

	EventEnter      = 0
	EventExit       = 1
	EventFuncName   = 255
	EventSessionEnd = 0xFE

	SessionIDSize  = 8
	NameHeaderSize = 17
)

type CallEvent struct {
	FuncID       uint32
	Kind         uint8
	Depth        uint32
	TSNs         uint64
	UnixTimeNs   uint64
	FuncType     uint8
	Flags        uint8
	ArgCount     uint8
	HasException uint8
	JITFlag      uint8
}

// DecodeEvent lit un nebula_event_t (48 bytes, padding inclus).
func DecodeEvent(b []byte) Event {
	return Event{
		SessionID:    binary.LittleEndian.Uint64(b[0:8]),
		Kind:         b[8],
		TSNs:         binary.LittleEndian.Uint64(b[16:24]),
		Depth:        binary.LittleEndian.Uint32(b[24:28]),
		FuncID:       binary.LittleEndian.Uint32(b[28:32]),
		FuncType:     b[32],
		Flags:        b[33],
		ArgCount:     b[34],
		HasException: b[35],
		JITFlag:      b[36],
		UnixTimeNs:   binary.LittleEndian.Uint64(b[40:48]),
	}
}
