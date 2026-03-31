package server

import (
	ingest2 "flownebula/agent/internal/ingest"
	"log"
	"net"
	"os"
)

func StartUnixStreamServer(path string, ring *ingest2.Ring, readers int) {
	// Supprimer l'ancien socket
	os.Remove(path)

	addr := &net.UnixAddr{Name: path, Net: "unix"}
	ln, err := net.ListenUnix("unix", addr)
	if err != nil {
		log.Fatalf("listen unix stream: %v", err)
	}

	log.Printf("UNIX STREAM server listening on %s", path)

	for {
		conn, err := ln.AcceptUnix()
		if err != nil {
			log.Printf("accept unix: %v", err)
			continue
		}

		// Plusieurs readers par connexion
		for i := 0; i < readers; i++ {
			go ingest2.FastReader(conn, ring)
		}
	}
}
