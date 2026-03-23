package main

import (
	"flag"
	"flownebula/agent/internal"
	"flownebula/agent/internal/sampler"
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
		daemon     = flag.Bool("daemon", false, "Run as daemon (background process)")
		logFile    = flag.String("log", "", "Path to log file")
		workers    = flag.Int("workers", 4, "Number of worker goroutines for event processing")
		samplePID  = flag.Int("sample-pid", 0, "PID to CPU-sample with perf_event_open")
		serverURL  = flag.String("server-url", "", "Nebula server URL (e.g. http://localhost:8080)")
		agentToken = flag.String("agent-token", "", "Agent token for server authentication")
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
	// --- sampler CPU (toujours actif si flag présent) ---
	if *samplePID > 0 {
		fd, err := sampler.StartCPUSampler(*samplePID, 99)
		if err != nil {
			log.Printf("perf sampler error: %v", err)
		} else {
			samples := make(chan sampler.Sample, 1024)
			go sampler.ReadSamples(fd, samples)

			go func() {
				for s := range samples {
					log.Printf("perf sample: %d frames", len(s.IPs))
				}
			}()
		}
	}

	sockPath := "/var/run/nebula.sock"
	os.MkdirAll("/var/run", 0755)

	if err := os.Remove(sockPath); err != nil && !os.IsNotExist(err) {
		log.Printf("Error removing old socket: %v", err)
	}

	addr := net.UnixAddr{Name: sockPath, Net: "unixgram"}

	var conn *net.UnixConn
	for {
		c, err := net.ListenUnixgram("unixgram", &addr)
		if err != nil {
			log.Printf("Error listening on unix socket: %v", err)
			time.Sleep(2 * time.Second)
			continue
		}
		conn = c
		break
	}
	defer conn.Close()

	uid, gid := detectPHPUser()
	_ = os.Chown(sockPath, uid, gid)
	_ = os.Chmod(sockPath, 0660)

	log.Printf("Nebula Agent listening on %s", sockPath)

	if *serverURL != "" && *agentToken != "" {
		sender, err := internal.NewSender(*serverURL, *agentToken)
		if err != nil {
			log.Printf("Warning: failed to connect to server: %v", err)
		} else {
			internal.GlobalSender = sender
			log.Printf("Connected to server: %s", *serverURL)
		}
	}

	go internal.ExportSessionsLoop()
	go internal.CleanupSessions()

	eventChan := make(chan internal.Packet, 10000)
	internal.StartWorkers(*workers, eventChan)

	internal.ListenUnixgram(conn, eventChan)
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
