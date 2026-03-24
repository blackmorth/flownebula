package sessions

import (
	"errors"
	"sync"
	"time"
)

type Repository interface {
	Create(session *Session) error
	List() ([]*Session, error)
	Get(id int64) (*Session, error)
}

type inMemoryRepo struct {
	mu       sync.RWMutex
	sessions map[int64]*Session
	nextID   int64
}

func NewInMemoryRepo() Repository {
	return &inMemoryRepo{
		sessions: make(map[int64]*Session),
		nextID:   1,
	}
}

func (r *inMemoryRepo) Create(session *Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	session.ID = r.nextID
	session.CreatedAt = time.Now()

	r.sessions[session.ID] = session
	r.nextID++

	return nil
}

func (r *inMemoryRepo) List() ([]*Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]*Session, 0, len(r.sessions))
	for _, s := range r.sessions {
		out = append(out, s)
	}
	return out, nil
}

func (r *inMemoryRepo) Get(id int64) (*Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	s, ok := r.sessions[id]
	if !ok {
		return nil, errors.New("session not found")
	}
	return s, nil
}
