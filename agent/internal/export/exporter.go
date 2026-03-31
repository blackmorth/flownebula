package export

import (
	"encoding/json"
	"flownebula/agent/internal/aggregator"
	"flownebula/agent/internal/metrics"
	"log"
	"time"
)

// ExportSessionsLoop parcourt périodiquement les sessions et exporte celles
// qui sont fermées ou inactives depuis un certain temps.
func ExportSessionsLoop(interval time.Duration, idleTimeout time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		exportReadySessions(idleTimeout)
	}
}

var backendToken string

func SetAgentToken(token string) {
	backendToken = token
}

var backendURL string

func SetServerURL(url string) {
	backendURL = url
}

func exportReadySessions(idleTimeout time.Duration) {
	now := time.Now()
	var activeCount int

	for i := 0; i < aggregator.NumShards; i++ {
		sh := aggregator.SessionShards[i]
		sh.Mu.Lock()
		for id, s := range sh.Sessions {
			s.Mu.Lock()

			// déjà exportée → on peut la supprimer
			if s.Exported {
				s.Mu.Unlock()
				delete(sh.Sessions, id)
				continue
			}

			inactive := now.Sub(s.LastSeen) > idleTimeout
			if s.Closed || inactive {
				// on exporte en dehors du lock de session
				s.Mu.Unlock()
				exportSession(s)
				s.Mu.Lock()
				s.Exported = true
				s.Mu.Unlock()
				delete(sh.Sessions, id)
				continue
			}

			s.Mu.Unlock()
			activeCount++
		}
		sh.Mu.Unlock()
	}

	metrics.MetricSessionsActive.Set(float64(activeCount))
}

func exportSession(s *aggregator.Session) {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	// Pour l’instant : export brut de la session en JSON.
	// Tu pourras remplacer ça par ton DetailedJSON + Sender HTTP.
	data, err := json.Marshal(s)
	if err != nil {
		log.Printf("exportSession: failed to marshal session %016x: %v", s.ID, err)
		return
	}

	log.Printf("exportSession: session %016x, size=%d bytes", s.ID, len(data))
	// TODO: brancher ici un Sender HTTP / WAL si tu veux.
}
