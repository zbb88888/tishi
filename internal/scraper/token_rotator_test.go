package scraper

import (
	"sync"
	"testing"
)

func TestTokenRotator_RoundRobin(t *testing.T) {
	r := NewTokenRotator([]string{"a", "b", "c"})

	// Should cycle through tokens
	expected := []string{"a", "b", "c", "a", "b"}
	for i, want := range expected {
		got := r.Next()
		if got != want {
			t.Errorf("call %d: got %q, want %q", i, got, want)
		}
	}
}

func TestTokenRotator_Empty(t *testing.T) {
	r := NewTokenRotator(nil)
	if got := r.Next(); got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
	if r.Count() != 0 {
		t.Errorf("Count() = %d, want 0", r.Count())
	}
}

func TestTokenRotator_FilterEmpty(t *testing.T) {
	r := NewTokenRotator([]string{"", "a", "", "b", ""})
	if r.Count() != 2 {
		t.Errorf("Count() = %d, want 2", r.Count())
	}
	if got := r.Next(); got != "a" {
		t.Errorf("first = %q, want a", got)
	}
}

func TestTokenRotator_Concurrent(t *testing.T) {
	r := NewTokenRotator([]string{"tok1", "tok2", "tok3"})

	var wg sync.WaitGroup
	results := make([]string, 100)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			results[idx] = r.Next()
		}(i)
	}
	wg.Wait()

	// All results should be valid tokens
	valid := map[string]bool{"tok1": true, "tok2": true, "tok3": true}
	for i, tok := range results {
		if !valid[tok] {
			t.Errorf("result[%d] = %q, not a valid token", i, tok)
		}
	}
}
