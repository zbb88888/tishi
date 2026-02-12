package collector

import "sync"

// TokenRotator manages multiple GitHub tokens for round-robin usage.
type TokenRotator struct {
	tokens  []string
	current int
	mu      sync.Mutex
}

// NewTokenRotator creates a TokenRotator with the given tokens.
func NewTokenRotator(tokens []string) *TokenRotator {
	return &TokenRotator{
		tokens: tokens,
	}
}

// Next returns the next token in the rotation.
// Returns empty string if no tokens are configured.
func (r *TokenRotator) Next() string {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.tokens) == 0 {
		return ""
	}

	token := r.tokens[r.current]
	r.current = (r.current + 1) % len(r.tokens)
	return token
}

// Count returns the number of available tokens.
func (r *TokenRotator) Count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.tokens)
}
