package scraper

import "sync"

// TokenRotator manages multiple GitHub tokens for round-robin usage.
// Thread-safe for concurrent access.
type TokenRotator struct {
	tokens  []string
	current int
	mu      sync.Mutex
}

// NewTokenRotator creates a TokenRotator with the given tokens.
// Empty or nil tokens slice results in unauthenticated API usage.
func NewTokenRotator(tokens []string) *TokenRotator {
	// Filter out empty strings
	clean := make([]string, 0, len(tokens))
	for _, t := range tokens {
		if t != "" {
			clean = append(clean, t)
		}
	}
	return &TokenRotator{tokens: clean}
}

// Next returns the next token in round-robin order.
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
