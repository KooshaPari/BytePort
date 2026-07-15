// Package resilience provides circuit-breaker pattern for external HTTP/DB calls.
//
// Pillar L87 — breaker pattern. Closed → Open after N consecutive failures.
// Open → HalfOpen after cool-down window. HalfOpen → Closed on success, Open on failure.
package resilience

import (
	"context"
	"errors"
	"sync"
	"time"
)

// State of a circuit breaker.
type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	}
	return "unknown"
}

// ErrBreakerOpen is returned when the breaker refuses a call.
var ErrBreakerOpen = errors.New("circuit breaker is open")

// Config configures a Breaker.
type Config struct {
	// FailureThreshold is the number of consecutive failures before opening.
	FailureThreshold int
	// SuccessThreshold is the number of consecutive successes in half-open before closing.
	SuccessThreshold int
	// CoolDown is how long the breaker stays open before transitioning to half-open.
	CoolDown time.Duration
	// Now is injected for testability; defaults to time.Now if nil.
	Now func() time.Time
}

// Breaker is a sequential circuit breaker; safe for concurrent use.
type Breaker struct {
	mu sync.Mutex

	cfg    Config
	state  State
	fails  int
	passes int

	openedAt time.Time
}

// New constructs a Breaker.
func New(cfg Config) *Breaker {
	if cfg.FailureThreshold == 0 {
		cfg.FailureThreshold = 5
	}
	if cfg.SuccessThreshold == 0 {
		cfg.SuccessThreshold = 2
	}
	if cfg.CoolDown == 0 {
		cfg.CoolDown = 30 * time.Second
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	return &Breaker{cfg: cfg, state: StateClosed}
}

// Allow returns nil if the call may proceed, ErrBreakerOpen otherwise.
// Side effect: transitions Open → HalfOpen if cool-down elapsed.
func (b *Breaker) Allow() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.state == StateOpen {
		if b.cfg.Now().Sub(b.openedAt) >= b.cfg.CoolDown {
			b.state = StateHalfOpen
			b.passes = 0
		}
	}

	if b.state == StateOpen {
		return ErrBreakerOpen
	}
	return nil
}

// RecordSuccess notes that the call succeeded.
func (b *Breaker) RecordSuccess() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.state == StateHalfOpen {
		b.passes++
		if b.passes >= b.cfg.SuccessThreshold {
			b.state = StateClosed
			b.fails = 0
			b.passes = 0
		}
		return
	}
	if b.state == StateClosed {
		b.fails = 0
	}
}

// RecordFailure notes that the call failed. May transition to Open.
func (b *Breaker) RecordFailure() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.state == StateHalfOpen {
		b.state = StateOpen
		b.openedAt = b.cfg.Now()
		b.passes = 0
		return
	}
	if b.state == StateClosed {
		b.fails++
		if b.fails >= b.cfg.FailureThreshold {
			b.state = StateOpen
			b.openedAt = b.cfg.Now()
		}
	}
}

// State returns the current state (for metrics/dashboards).
func (b *Breaker) State() State {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.state == StateOpen && b.cfg.Now().Sub(b.openedAt) >= b.cfg.CoolDown {
		// Promote lazily to HalfOpen so callers seeing stale state will not be surprised.
		b.state = StateHalfOpen
		b.passes = 0
	}
	return b.state
}

// Do runs fn under breaker protection. Returns ErrBreakerOpen if breaker is open.
func (b *Breaker) Do(ctx context.Context, fn func(context.Context) error) error {
	if err := b.Allow(); err != nil {
		return err
	}
	err := fn(ctx)
	if err != nil {
		b.RecordFailure()
		return err
	}
	b.RecordSuccess()
	return nil
}
