package main

import (
	"flag"
	"flownebula/agent"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"time"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Printf("Starting Nebula Agent...")
	var (
		daemon  = flag.Bool("daemon", false, "Run as daemon (background process)")
		logFile = flag.String("log", "", "Path to log file")
		workers = flag.Int("workers", 4, "Number of worker goroutines for event processing")
	)
	flag.Parse()

	if *daemon {
		args := make([]string, 0, len(os.Args)-1)
		for i := 1; i < len(os.Args); i++ {
			arg := os.Args[i]
			if arg == "-daemon" || arg == "--daemon" || arg == "-daemon=true" || arg == "--daemon=true" {
				continue
			}
			args = append(args, arg)
		}
		cmd := exec.Command(os.Args[0], args...)
		if *logFile != "" {
			f, err := os.OpenFile(*logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				log.Fatal(err)
			}
			cmd.Stdout = f
			cmd.Stderr = f
		}
		if err := cmd.Start(); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Nebula Agent started in background with PID %d\n", cmd.Process.Pid)
		os.Exit(0)
	}

	if *logFile != "" {
		f, err := os.OpenFile(*logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}
		log.SetOutput(f)
		defer func(f *os.File) {
			err := f.Close()
			if err != nil {

			}
		}(f)
	} else {
		log.SetOutput(os.Stderr)
	}

	sockPath := "/var/run/nebula.sock"
	err := os.Remove(sockPath)
	if err != nil {
		return
	} // au cas où

	addr := net.UnixAddr{Name: sockPath, Net: "unixgram"}
	conn, err := net.ListenUnixgram("unixgram", &addr)
	if err != nil {
		log.Fatalf("Error listening on unix socket: %v", err)
	}
	defer func(conn *net.UnixConn) {
		err := conn.Close()
		if err != nil {

		}
	}(conn)

	// Permissions correctes pour PHP-FPM
	uid, gid := detectPHPUser()
	err = os.Chown(sockPath, uid, gid)
	if err != nil {
		return
	}
	err = os.Chmod(sockPath, 0660)
	if err != nil {
		return
	}

	log.Printf("Nebula Agent listening on %s", sockPath)

	go agent.CleanupSessions()

	go func() {
		for range time.Tick(2 * time.Second) {
			var activeSessions []*agent.Session
			for i := 0; i < agent.NumShards; i++ {
				shard := agent.SessionShards[i]
				shard.Mu.RLock()
				for _, s := range shard.Sessions {
					activeSessions = append(activeSessions, s)
				}
				shard.Mu.RUnlock()
			}

			for _, s := range activeSessions {
				s.Print()
			}
		}
	}()

	eventChan := make(chan agent.Packet, 10000)
	agent.StartWorkers(*workers, eventChan)

	agent.ListenUnixgram(conn, eventChan)
}

func detectPHPUser() (int, int) {
	candidates := []string{"www-data", "apache", "nginx", "php"}

	for _, name := range candidates {
		u, err := user.Lookup(name)
		if err == nil {
			uid, _ := strconv.Atoi(u.Uid)
			gid, _ := strconv.Atoi(u.Gid)
			return uid, gid
		}
	}

	// fallback: root
	return os.Getuid(), os.Getgid()
}
