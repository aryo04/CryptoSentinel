package network

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// CircuitState represents the state of the circuit breaker
type CircuitState int32

const (
	// CircuitClosed allows requests to pass through
	CircuitClosed CircuitState = iota
	// CircuitOpen blocks all requests
	CircuitOpen
	// CircuitHalfOpen allows limited requests for testing
	CircuitHalfOpen
)

// String returns string representation of circuit state
func (s CircuitState) String() string {
	switch s {
	case CircuitClosed:
		return "closed"
	case CircuitOpen:
		return "open"
	case CircuitHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// CircuitBreaker implements the circuit breaker pattern for connection failures
type CircuitBreaker struct {
	maxFailures      int
	resetTimeout     time.Duration
	halfOpenRequests int
	
	state            int32 // atomic CircuitState
	failures         int
	successCount     int
	lastFailTime     time.Time
	halfOpenAttempts int
	
	mu               sync.RWMutex
	onStateChange    func(from, to CircuitState)
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures:      maxFailures,
		resetTimeout:     resetTimeout,
		halfOpenRequests: 1, // Allow one request in half-open state
		state:           int32(CircuitClosed),
	}
}

// SetStateChangeHandler sets a callback for state changes
func (cb *CircuitBreaker) SetStateChangeHandler(handler func(from, to CircuitState)) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.onStateChange = handler
}

// Call executes the given function if the circuit allows it
func (cb *CircuitBreaker) Call(fn func() error) error {
	if !cb.CanAttempt() {
		return fmt.Errorf("circuit breaker is open")
	}
	
	err := fn()
	cb.RecordResult(err)
	return err
}

// CanAttempt returns whether a request can be attempted
func (cb *CircuitBreaker) CanAttempt() bool {
	state := CircuitState(atomic.LoadInt32(&cb.state))
	
	switch state {
	case CircuitClosed:
		return true
		
	case CircuitOpen:
		cb.mu.RLock()
		shouldReset := time.Since(cb.lastFailTime) > cb.resetTimeout
		cb.mu.RUnlock()
		
		if shouldReset {
			cb.transitionTo(CircuitHalfOpen)
			return true
		}
		return false
		
	case CircuitHalfOpen:
		cb.mu.Lock()
		defer cb.mu.Unlock()
		
		if cb.halfOpenAttempts < cb.halfOpenRequests {
			cb.halfOpenAttempts++
			return true
		}
		return false
		
	default:
		return false
	}
}

// RecordResult records the result of an attempt
func (cb *CircuitBreaker) RecordResult(err error) {
	state := CircuitState(atomic.LoadInt32(&cb.state))
	
	if err != nil {
		cb.recordFailure(state)
	} else {
		cb.recordSuccess(state)
	}
}

// recordFailure records a failure and potentially opens the circuit
func (cb *CircuitBreaker) recordFailure(currentState CircuitState) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	
	cb.failures++
	cb.lastFailTime = time.Now()
	
	switch currentState {
	case CircuitClosed:
		if cb.failures >= cb.maxFailures {
			cb.transitionToLocked(CircuitOpen)
		}
		
	case CircuitHalfOpen:
		// Single failure in half-open state opens the circuit
		cb.transitionToLocked(CircuitOpen)
		cb.halfOpenAttempts = 0
	}
}

// recordSuccess records a success and potentially closes the circuit
func (cb *CircuitBreaker) recordSuccess(currentState CircuitState) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	
	switch currentState {
	case CircuitHalfOpen:
		cb.successCount++
		if cb.successCount >= cb.halfOpenRequests {
			// Successfully tested, close the circuit
			cb.failures = 0
			cb.successCount = 0
			cb.halfOpenAttempts = 0
			cb.transitionToLocked(CircuitClosed)
		}
		
	case CircuitClosed:
		// Reset failure count on success
		if cb.failures > 0 {
			cb.failures = 0
		}
	}
}

// transitionTo transitions to a new state (thread-safe)
func (cb *CircuitBreaker) transitionTo(newState CircuitState) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.transitionToLocked(newState)
}

// transitionToLocked transitions to a new state (must hold lock)
func (cb *CircuitBreaker) transitionToLocked(newState CircuitState) {
	oldState := CircuitState(atomic.LoadInt32(&cb.state))
	if oldState == newState {
		return
	}
	
	atomic.StoreInt32(&cb.state, int32(newState))
	
	// Reset state-specific counters
	if newState == CircuitHalfOpen {
		cb.halfOpenAttempts = 0
		cb.successCount = 0
	}
	
	// Notify state change
	if cb.onStateChange != nil {
		// Call handler without holding lock to prevent deadlock
		go cb.onStateChange(oldState, newState)
	}
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() CircuitState {
	return CircuitState(atomic.LoadInt32(&cb.state))
}

// Reset resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	
	cb.failures = 0
	cb.successCount = 0
	cb.halfOpenAttempts = 0
	cb.lastFailTime = time.Time{}
	atomic.StoreInt32(&cb.state, int32(CircuitClosed))
}

// GetStats returns current statistics
func (cb *CircuitBreaker) GetStats() CircuitBreakerStats {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	
	return CircuitBreakerStats{
		State:            CircuitState(atomic.LoadInt32(&cb.state)),
		Failures:         cb.failures,
		SuccessCount:     cb.successCount,
		LastFailTime:     cb.lastFailTime,
		HalfOpenAttempts: cb.halfOpenAttempts,
	}
}

// CircuitBreakerStats contains circuit breaker statistics
type CircuitBreakerStats struct {
	State            CircuitState
	Failures         int
	SuccessCount     int
	LastFailTime     time.Time
	HalfOpenAttempts int
}