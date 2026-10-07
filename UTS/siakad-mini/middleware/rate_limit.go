package middleware

import (
	"sync"
	"time"
)

type FailureTracker struct {
	mu       sync.Mutex
	attempts map[string][]time.Time
	window   time.Duration
	maxFails int
}

func NewFailureTracker(maxFails int, window time.Duration) *FailureTracker {
	tracker := &FailureTracker{
		attempts: make(map[string][]time.Time),
		window:   window,
		maxFails: maxFails,
	}

	// Rutin pembersih histori berkala tiap 2 menit
	go func() {
		ticker := time.NewTicker(2 * time.Minute)
		for range ticker.C {
			tracker.cleanup()
		}
	}()

	return tracker
}

func (ft *FailureTracker) IsBlocked(key string) bool {
	ft.mu.Lock()
	defer ft.mu.Unlock()

	cutoff := time.Now().Add(-ft.window)
	timestamps := ft.attempts[key]

	var valid []time.Time
	for _, t := range timestamps {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}
	ft.attempts[key] = valid

	return len(valid) >= ft.maxFails
}

func (ft *FailureTracker) RecordFailure(key string) {
	ft.mu.Lock()
	defer ft.mu.Unlock()

	cutoff := time.Now().Add(-ft.window)
	timestamps := ft.attempts[key]

	var valid []time.Time
	for _, t := range timestamps {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}
	valid = append(valid, time.Now())
	ft.attempts[key] = valid
}

func (ft *FailureTracker) Reset(key string) {
	ft.mu.Lock()
	defer ft.mu.Unlock()
	delete(ft.attempts, key)
}

func (ft *FailureTracker) cleanup() {
	ft.mu.Lock()
	defer ft.mu.Unlock()

	cutoff := time.Now().Add(-ft.window)
	for key, timestamps := range ft.attempts {
		var valid []time.Time
		for _, t := range timestamps {
			if t.After(cutoff) {
				valid = append(valid, t)
			}
		}
		if len(valid) == 0 {
			delete(ft.attempts, key)
		} else {
			ft.attempts[key] = valid
		}
	}
}
