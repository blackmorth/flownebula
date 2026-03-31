package protocol

// Event représente un événement brut lu depuis le socket.
// Mappé exactement sur nebula_event_t (48 bytes, padding inclus).
type Event struct {
	SessionID    uint64
	Kind         uint8
	TSNs         uint64
	Depth        uint32
	FuncID       uint32
	FuncType     uint8
	Flags        uint8
	ArgCount     uint8
	HasException uint8
	JITFlag      uint8
	UnixTimeNs   uint64
}

// EventSize Taille exacte d’un événement binaire envoyé par le probe.
const EventSize = 48
