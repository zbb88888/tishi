package collector

import (
	"sync"
	"testing"
)

func TestTokenRotator_Empty(t *testing.T) {
	r := NewTokenRotator(nil)
	if got := r.Next(); got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
	if got := r.Count(); got != 0 {
		t.Errorf("expected count 0, got %d", got)
	}
}

func TestTokenRotator_SingleToken(t *testing.T) {
	r := NewTokenRotator([]string{"tok-a"})
	for i := 0; i < 5; i++ {
		if got := r.Next(); got != "tok-a" {
			t.Errorf("iteration %d: expected tok-a, got %q", i, got)
		}
	}
}

func TestTokenRotator_RoundRobin(t *testing.T) {
	tokens := []string{"tok-a", "tok-b", "tok-c"}
	r := NewTokenRotator(tokens)

	for cycle := 0; cycle < 3; cycle++ {
		for i, expected := range tokens {
			got := r.Next()
			if got != expected {
				t.Errorf("cycle %d, index %d: expected %q, got %q", cycle, i, expected, got)
			}
		}
	}
}

func TestTokenRotator_Concurrent(t *testing.T) {
	tokens := []string{"a", "b", "c"}
	r := NewTokenRotator(tokens)

	var wg sync.WaitGroup
	results := make(chan string, 100)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results <- r.Next()
		}()
	}

	wg.Wait()
	close(results)

	counts := make(map[string]int)
	for tok := range results {
		counts[tok]++
	}

	// 确保所有 token 都被使用过
	for _, tok := range tokens {
		if counts[tok] == 0 {
			t.Errorf("token %q was never used", tok)
		}
	}
}
