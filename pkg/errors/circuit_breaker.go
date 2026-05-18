package errors

import (
	"fmt"
	"sync"
	"time"
)

// CircuitBreakerState rappresenta lo stato del circuit breaker
type CircuitBreakerState string

const (
	StateClosed   CircuitBreakerState = "closed"   // Normale, permette richieste
	StateOpen     CircuitBreakerState = "open"     // In errore, blocca richieste
	StateHalfOpen CircuitBreakerState = "half_open" // Tentativo di ripresa
)

// CircuitBreaker implementa il pattern circuit breaker per resilienza
type CircuitBreaker struct {
	mu                 sync.RWMutex
	state              CircuitBreakerState
	failureCount       int
	lastFailureTime    time.Time
	successCount       int
	maxFailures        int
	resetTimeout       time.Duration
	halfOpenThreshold  int
}

// NewCircuitBreaker crea un nuovo circuit breaker
func NewCircuitBreaker(maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:             StateClosed,
		maxFailures:       maxFailures,
		resetTimeout:      resetTimeout,
		halfOpenThreshold: 2, // Numero di successi richiesti in half-open
	}
}

// Call esegue un'operazione attraverso il circuit breaker
func (cb *CircuitBreaker) Call(fn func() error) error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	// Se aperto e il timeout è scaduto, passa a half-open
	if cb.state == StateOpen {
		if time.Since(cb.lastFailureTime) > cb.resetTimeout {
			cb.state = StateHalfOpen
			cb.successCount = 0
		} else {
			return fmt.Errorf("circuit breaker is open")
		}
	}

	// Se in half-open, esegui con test
	if cb.state == StateHalfOpen {
		if err := fn(); err != nil {
			// Errore in half-open, torna ad open
			cb.state = StateOpen
			cb.lastFailureTime = time.Now()
			cb.failureCount = 0
			return err
		}
		// Successo in half-open
		cb.successCount++
		if cb.successCount >= cb.halfOpenThreshold {
			// Abbastanza successi, torna a closed
			cb.state = StateClosed
			cb.failureCount = 0
			cb.successCount = 0
		}
		return nil
	}

	// Se closed, esegui normalmente
	if err := fn(); err != nil {
		cb.failureCount++
		cb.lastFailureTime = time.Now()
		if cb.failureCount >= cb.maxFailures {
			cb.state = StateOpen
		}
		return err
	}

	// Successo, reset counter
	cb.failureCount = 0
	return nil
}

// GetState ritorna lo stato attuale del circuit breaker
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// Reset resetta il circuit breaker
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.state = StateClosed
	cb.failureCount = 0
	cb.successCount = 0
	cb.lastFailureTime = time.Time{}
}
