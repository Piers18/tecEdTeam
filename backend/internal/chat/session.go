package chat

import (
	"sync"

	"github.com/openai/openai-go"
)

// SessionStore holds in-memory conversation histories keyed by session ID.
type SessionStore struct {
	mu       sync.RWMutex
	sessions map[string][]openai.ChatCompletionMessageParamUnion
}

func NewSessionStore() *SessionStore {
	return &SessionStore{
		sessions: make(map[string][]openai.ChatCompletionMessageParamUnion),
	}
}

func (s *SessionStore) Get(id string) []openai.ChatCompletionMessageParamUnion {
	s.mu.RLock()
	defer s.mu.RUnlock()
	hist := s.sessions[id]
	cp := make([]openai.ChatCompletionMessageParamUnion, len(hist))
	copy(cp, hist)
	return cp
}

func (s *SessionStore) Append(id string, msgs ...openai.ChatCompletionMessageParamUnion) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[id] = append(s.sessions[id], msgs...)
}

func (s *SessionStore) Delete(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, id)
}

func (s *SessionStore) Exists(id string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.sessions[id]
	return ok
}
