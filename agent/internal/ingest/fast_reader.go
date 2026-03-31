package ingest

import (
	"flownebula/agent/internal/protocol"
	"net"
)

func FastReader(conn *net.UnixConn, ring *Ring) {
	buf := make([]byte, 65536-(65536%protocol.EventSize))

	for {
		n, err := conn.Read(buf)
		if err != nil {
			return
		}

		for off := 0; off+protocol.EventSize <= n; off += protocol.EventSize {
			ev := protocol.DecodeEvent(buf[off : off+protocol.EventSize])
			ring.Push(ev)
		}
	}
}
