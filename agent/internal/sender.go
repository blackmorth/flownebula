package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// Sender handles authenticated HTTP communication with the Nebula server.
type Sender struct {
	serverURL string
	mu        sync.RWMutex
	jwt       string
	client    *http.Client
}

// GlobalSender is the package-level sender, set at startup.
var GlobalSender *Sender

// NewSender authenticates against the server using the agent token and returns a ready Sender.
func NewSender(serverURL, agentToken string) (*Sender, error) {
	s := &Sender{
		serverURL: serverURL,
		client:    &http.Client{Timeout: 10 * time.Second},
	}
	if err := s.login(agentToken); err != nil {
		return nil, err
	}
	return s, nil
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

// SendProfile sends the exported session JSON payload to the server.
// agentID is the hex representation of the session's uint64 ID.
func (s *Sender) SendProfile(agentID string, payload []byte) error {
	s.mu.RLock()
	jwt := s.jwt
	s.mu.RUnlock()

	reqBody, _ := json.Marshal(map[string]string{
		"agent_id": agentID,
		"payload":  string(payload),
	})

	req, err := http.NewRequest(http.MethodPost, s.serverURL+"/agent/session-upload", bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("failed to build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+jwt)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("send profile request failed: %w", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}
	return nil
}
