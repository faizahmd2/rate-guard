package auth

import (
	"sync"
)

type TokenStore struct {
	mu     sync.RWMutex
	hashes map[string]struct{}
}

func NewTokenStore(tokens []APIToken) *TokenStore {
	hashes := make(map[string]struct{}, len(tokens))

	for _, token := range tokens {
		if token.Status == "active" {
			hashes[token.TokenHash] = struct{}{}
		}
	}

	return &TokenStore{
		hashes: hashes,
	}
}

func (s *TokenStore) Add(tokenHash string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.hashes[tokenHash] = struct{}{}
}

func (s *TokenStore) Remove(tokenHash string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.hashes, tokenHash)
}

func (s *TokenStore) Has(tokenHash string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, ok := s.hashes[tokenHash]
	return ok
}
