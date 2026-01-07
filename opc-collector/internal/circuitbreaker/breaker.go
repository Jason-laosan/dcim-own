package circuitbreaker

import (
	"errors"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/yourusername/opc-collector/pkg/config"
	"github.com/yourusername/opc-collector/pkg/logger"
)

var (
	// ErrCircuitOpen is returned when the circuit breaker is open
	ErrCircuitOpen = errors.New("circuit breaker is open")
)

// State represents the circuit breaker state
type State string

const (
	StateClosed   State = "closed"
	StateOpen     State = "open"
	StateHalfOpen State = "half_open"
)

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	mu                  sync.RWMutex
	state               State
	failureThreshold    int
	successThreshold    int
	timeout             time.Duration
	halfOpenMaxRequests int

	failureCount        int
	successCount        int
	lastFailureTime     time.Time
	halfOpenCount       int

	deviceID            string
	logger              *zap.Logger
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(deviceID string, cfg config.CircuitBreakerConfig) *CircuitBreaker {
	return &CircuitBreaker{
		state:               StateClosed,
		failureThreshold:    cfg.FailureThreshold,
		successThreshold:    cfg.SuccessThreshold,
		timeout:             time.Duration(cfg.Timeout) * time.Second,
		halfOpenMaxRequests: cfg.HalfOpenMaxRequests,
		deviceID:            deviceID,
		logger:              logger.Named("circuit-breaker").With(zap.String("device_id", deviceID)),
	}
}

// Execute executes a function with circuit breaker protection
func (cb *CircuitBreaker) Execute(fn func() error) error {
	if !cb.canExecute() {
		return ErrCircuitOpen
	}

	err := fn()

	if err != nil {
		cb.recordFailure()
		return err
	}

	cb.recordSuccess()
	return nil
}

// canExecute checks if the circuit breaker allows execution
func (cb *CircuitBreaker) canExecute() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		return true

	case StateOpen:
		// Check if timeout has passed
		if time.Since(cb.lastFailureTime) > cb.timeout {
			cb.logger.Info("circuit breaker transitioning to half-open")
			cb.state = StateHalfOpen
			cb.halfOpenCount = 0
			return true
		}
		return false

	case StateHalfOpen:
		// Allow limited requests in half-open state
		if cb.halfOpenCount < cb.halfOpenMaxRequests {
			cb.halfOpenCount++
			return true
		}
		return false
	}

	return false
}

// recordFailure records a failure
func (cb *CircuitBreaker) recordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount++
	cb.lastFailureTime = time.Now()
	cb.successCount = 0

	switch cb.state {
	case StateClosed:
		if cb.failureCount >= cb.failureThreshold {
			cb.logger.Warn("circuit breaker opening",
				zap.Int("failure_count", cb.failureCount),
				zap.Int("threshold", cb.failureThreshold))
			cb.state = StateOpen
			cb.failureCount = 0
		}

	case StateHalfOpen:
		cb.logger.Warn("circuit breaker reopening due to failure in half-open state")
		cb.state = StateOpen
		cb.failureCount = 0
		cb.halfOpenCount = 0
	}
}

// recordSuccess records a success
func (cb *CircuitBreaker) recordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.successCount++
	cb.failureCount = 0

	switch cb.state {
	case StateHalfOpen:
		if cb.successCount >= cb.successThreshold {
			cb.logger.Info("circuit breaker closing",
				zap.Int("success_count", cb.successCount),
				zap.Int("threshold", cb.successThreshold))
			cb.state = StateClosed
			cb.successCount = 0
			cb.halfOpenCount = 0
		}
	}
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// Reset resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state = StateClosed
	cb.failureCount = 0
	cb.successCount = 0
	cb.halfOpenCount = 0

	cb.logger.Info("circuit breaker reset")
}

// Stats returns circuit breaker statistics
type Stats struct {
	State        State
	FailureCount int
	SuccessCount int
}

// GetStats returns current statistics
func (cb *CircuitBreaker) GetStats() Stats {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return Stats{
		State:        cb.state,
		FailureCount: cb.failureCount,
		SuccessCount: cb.successCount,
	}
}
