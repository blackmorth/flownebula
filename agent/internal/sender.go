package internal

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Sender handles authenticated HTTP communication with the Nebula server.
type Sender struct {
	serverURL string
	mu        sync.RWMutex
	jwt       string
	client    *http.Client
	walPath   string
	opts      SenderOptions
}

type SenderOptions struct {
	MaxRetries       int
	InitialBackoffMs int
}

// GlobalSender is the package-level sender, set at startup.
var GlobalSender *Sender

// NewSender authenticates against the server using the agent token and returns a ready Sender.
func NewSender(serverURL, agentToken, walPath string, opts SenderOptions) (*Sender, error) {
	if opts.MaxRetries <= 0 {
		opts.MaxRetries = 3
	}
	if opts.InitialBackoffMs <= 0 {
		opts.InitialBackoffMs = 500
	}
	s := &Sender{
		serverURL: serverURL,
		client:    &http.Client{Timeout: 10 * time.Second},
		walPath:   walPath,
		opts:      opts,
	}
	if err := s.loginWithRetry(agentToken); err != nil {
		return nil, err
	}
	if walPath != "" {
		if err := os.MkdirAll(filepath.Dir(walPath), 0o755); err != nil {
			return nil, fmt.Errorf("failed to create WAL dir: %w", err)
		}
	}
	go s.replayWALLoop()
	return s, nil
}

func (s *Sender) loginWithRetry(agentToken string) error {
	var lastErr error
	backoff := time.Duration(s.opts.InitialBackoffMs) * time.Millisecond
	for attempt := 1; attempt <= s.opts.MaxRetries; attempt++ {
		if attempt > 1 {
			AgentMetrics.RetryAttempts.Add(1)
			time.Sleep(backoff)
			if backoff < 5*time.Second {
				backoff *= 2
			}
		}
		if err := s.login(agentToken); err != nil {
			lastErr = err
			continue
		}
		return nil
	}
	return fmt.Errorf("agent login failed after retries: %w", lastErr)
}

func (s *Sender) login(agentToken string) error {
	body, _ := json.Marshal(map[string]string{"agent_token": agentToken})

	resp, err := s.client.Post(s.serverURL+"/auth/agent-login", "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("agent login request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("agent login returned status %d: %s", resp.StatusCode, string(raw))
	}

	var result struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode login response: %w", err)
	}
	if result.Token == "" {
		return fmt.Errorf("empty token in login response")
	}

	s.mu.Lock()
	s.jwt = result.Token
	s.mu.Unlock()
	return nil
}

func (s *Sender) SendSession(payload []byte) error {
	err := s.sendWithRetry(payload, s.opts.MaxRetries)
	if err == nil {
		return nil
	}
	if walErr := s.appendToWAL(payload); walErr != nil {
		return fmt.Errorf("send failed: %w (WAL append failed: %v)", err, walErr)
	}
	AgentMetrics.SendFailures.Add(1)
	return err
}

func (s *Sender) sendWithRetry(payload []byte, maxAttempts int) error {
	var lastErr error
	backoff := time.Duration(s.opts.InitialBackoffMs) * time.Millisecond
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if attempt > 1 {
			AgentMetrics.RetryAttempts.Add(1)
			time.Sleep(backoff)
			if backoff < 5*time.Second {
				backoff *= 2
			}
		}
		if err := s.sendOnce(payload); err != nil {
			lastErr = err
			continue
		}
		return nil
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("upload failed without details")
	}
	return lastErr
}

func (s *Sender) sendOnce(payload []byte) error {
	s.mu.RLock()
	jwt := s.jwt
	s.mu.RUnlock()

	req, err := http.NewRequest(http.MethodPost, s.serverURL+"/agent/session-upload", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("failed to build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+jwt)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("session upload request failed: %w", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}
	return nil
}

func (s *Sender) appendToWAL(payload []byte) error {
	if s.walPath == "" {
		return nil
	}
	f, err := os.OpenFile(s.walPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.Write(append(payload, '\n')); err != nil {
		return err
	}
	AgentMetrics.WALQueued.Add(1)
	return nil
}

func (s *Sender) replayWALLoop() {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		if s.walPath == "" {
			continue
		}
		s.replayWALOnce()
	}
}

func (s *Sender) replayWALOnce() {
	f, err := os.Open(s.walPath)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	var remaining [][]byte
	for scanner.Scan() {
		line := append([]byte(nil), scanner.Bytes()...)
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		if err := s.sendWithRetry(line, s.opts.MaxRetries); err != nil {
			remaining = append(remaining, line)
			continue
		}
		AgentMetrics.WALReplayed.Add(1)
	}

	tmp := s.walPath + ".tmp"
	out, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return
	}
	for _, line := range remaining {
		_, _ = out.Write(append(line, '\n'))
	}
	_ = out.Close()
	_ = os.Rename(tmp, s.walPath)
}
