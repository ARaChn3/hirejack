// Package canary provides canary token generation and callback tracking.
package canary

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// Token represents a canary token.
type Token struct {
	ID          string     `json:"id"`
	CreatedAt   time.Time  `json:"created_at"`
	CallbackURL string     `json:"callback_url"`
	Target      string     `json:"target,omitempty"`
	Callbacks   []Callback `json:"callbacks"`
}

// Callback represents a callback hit.
type Callback struct {
	Timestamp time.Time         `json:"timestamp"`
	IP        string            `json:"ip"`
	UserAgent string            `json:"user_agent"`
	Headers   map[string]string `json:"headers,omitempty"`
}

// Store holds tokens in memory.
type Store struct {
	mu     sync.RWMutex
	tokens map[string]*Token
}

// NewStore creates a new token store.
func NewStore() *Store {
	return &Store{
		tokens: make(map[string]*Token),
	}
}

// Generate creates a new canary token.
func (s *Store) Generate(baseURL, target string) *Token {
	id := generateTokenID()
	callbackURL := fmt.Sprintf("%s/c/%s", baseURL, id)

	token := &Token{
		ID:          id,
		CreatedAt:   time.Now(),
		CallbackURL: callbackURL,
		Target:      target,
		Callbacks:   []Callback{},
	}

	s.mu.Lock()
	s.tokens[id] = token
	s.mu.Unlock()

	return token
}

// Get retrieves a token by ID.
func (s *Store) Get(id string) *Token {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.tokens[id]
}

// GetAll returns all tokens.
func (s *Store) GetAll() []*Token {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tokens := make([]*Token, 0, len(s.tokens))
	for _, t := range s.tokens {
		tokens = append(tokens, t)
	}
	return tokens
}

// RecordCallback records a callback hit.
func (s *Store) RecordCallback(id string, cb Callback) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	token, ok := s.tokens[id]
	if !ok {
		return false
	}

	if cb.Timestamp.IsZero() {
		cb.Timestamp = time.Now()
	}

	token.Callbacks = append(token.Callbacks, cb)
	return true
}

// Delete removes a token.
func (s *Store) Delete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.tokens[id]; !ok {
		return false
	}
	delete(s.tokens, id)
	return true
}

// Count returns the number of tokens.
func (s *Store) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.tokens)
}

// GenerateXXE returns XXE DOCTYPE with callback URL.
func (t *Token) GenerateXXE() string {
	return fmt.Sprintf(`<!DOCTYPE foo [
  <!ENTITY canary SYSTEM "%s">
]>`, t.CallbackURL)
}

// GenerateImageURL returns a tracking pixel URL.
func (t *Token) GenerateImageURL() string {
	return t.CallbackURL + ".png"
}

// HasCallbacks returns true if the token has received callbacks.
func (t *Token) HasCallbacks() bool {
	return len(t.Callbacks) > 0
}

// LastCallback returns the most recent callback, or nil.
func (t *Token) LastCallback() *Callback {
	if len(t.Callbacks) == 0 {
		return nil
	}
	return &t.Callbacks[len(t.Callbacks)-1]
}

func generateTokenID() string {
	b := make([]byte, 4)
	rand.Read(b)
	return "hj_" + hex.EncodeToString(b)
}
