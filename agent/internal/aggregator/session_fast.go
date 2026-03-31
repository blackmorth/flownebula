package aggregator

import (
	"flownebula/agent/internal/protocol"
)

const (
	maxFuncsPerSession = 65536
	maxStackDepth      = 1024
)

type funcMetrics struct {
	Count uint64
	WT    uint64 // wall time cumulé en ns
}

type stackFrame struct {
	FuncID uint32
	Enter  uint64 // ts_ns à l’entrée
}

type SessionFast struct {
	stack [maxStackDepth]stackFrame
	sp    int

	metrics [maxFuncsPerSession]funcMetrics

	firstTS uint64
	lastTS  uint64

	closed bool
}

func newSessionFast() *SessionFast {
	return &SessionFast{}
}

func (sf *SessionFast) addEvent(ev protocol.Event) {
	switch ev.Kind {

	case protocol.EventEnter:
		if sf.sp < maxStackDepth {
			sf.stack[sf.sp] = stackFrame{
				FuncID: ev.FuncID,
				Enter:  ev.TSNs,
			}
			sf.sp++
		}
		if ev.TSNs > 0 {
			if sf.firstTS == 0 || ev.TSNs < sf.firstTS {
				sf.firstTS = ev.TSNs
			}
			if ev.TSNs > sf.lastTS {
				sf.lastTS = ev.TSNs
			}
		}

	case protocol.EventExit:
		if sf.sp == 0 {
			return
		}
		sf.sp--
		frame := sf.stack[sf.sp]
		if frame.FuncID >= maxFuncsPerSession {
			return
		}
		m := &sf.metrics[frame.FuncID]
		m.Count++
		if ev.TSNs > frame.Enter {
			m.WT += ev.TSNs - frame.Enter
		}
		if ev.TSNs > 0 {
			if sf.firstTS == 0 || ev.TSNs < sf.firstTS {
				sf.firstTS = ev.TSNs
			}
			if ev.TSNs > sf.lastTS {
				sf.lastTS = ev.TSNs
			}
		}

	case protocol.EventSessionEnd:
		sf.closed = true
	}
}
