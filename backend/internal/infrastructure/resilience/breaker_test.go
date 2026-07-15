package resilience_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/byteport/api/internal/infrastructure/resilience"
)

func TestBreaker_StartsClosed(t *testing.T) {
	b := resilience.New(resilience.Config{})
	if got := b.State(); got != resilience.StateClosed {
		t.Fatalf("want closed, got %s", got)
	}
	if err := b.Allow(); err != nil {
		t.Fatalf("Allow on closed breaker should succeed: %v", err)
	}
}

func TestBreaker_OpensAfterThresholdFailures(t *testing.T) {
	b := resilience.New(resilience.Config{FailureThreshold: 3})
	for i := 0; i < 3; i++ {
		b.RecordFailure()
	}
	if got := b.State(); got != resilience.StateOpen {
		t.Fatalf("want open after 3 failures, got %s", got)
	}
	if err := b.Allow(); !errors.Is(err, resilience.ErrBreakerOpen) {
		t.Fatalf("want ErrBreakerOpen, got %v", err)
	}
}

func TestBreaker_TransitionsToHalfOpenAfterCoolDown(t *testing.T) {
	now := time.Unix(0, 0)
	b := resilience.New(resilience.Config{
		FailureThreshold: 1,
		CoolDown:         time.Minute,
		Now:              func() time.Time { return now },
	})
	b.RecordFailure()
	if got := b.State(); got != resilience.StateOpen {
		t.Fatalf("want open, got %s", got)
	}
	now = now.Add(time.Minute + time.Second)
	if got := b.State(); got != resilience.StateHalfOpen {
		t.Fatalf("want half-open after cool-down, got %s", got)
	}
	if err := b.Allow(); err != nil {
		t.Fatalf("Allow in half-open should succeed: %v", err)
	}
}

func TestBreaker_HalfOpenClosesAfterSuccessThreshold(t *testing.T) {
	b := resilience.New(resilience.Config{
		FailureThreshold: 1,
		SuccessThreshold: 2,
		CoolDown:         time.Millisecond,
	})
	b.RecordFailure()
	time.Sleep(time.Millisecond)
	// Trigger lazy transition via Allow.
	if err := b.Allow(); err != nil {
		t.Fatalf("Allow after cool-down: %v", err)
	}
	b.RecordSuccess()
	b.RecordSuccess()
	if got := b.State(); got != resilience.StateClosed {
		t.Fatalf("want closed, got %s", got)
	}
}

func TestBreaker_HalfOpenReopensOnFailure(t *testing.T) {
	b := resilience.New(resilience.Config{
		FailureThreshold: 1,
		CoolDown:         time.Millisecond,
	})
	b.RecordFailure()
	time.Sleep(time.Millisecond)
	if err := b.Allow(); err != nil {
		t.Fatalf("Allow: %v", err)
	}
	b.RecordFailure()
	if got := b.State(); got != resilience.StateOpen {
		t.Fatalf("want open, got %s", got)
	}
}

func TestBreaker_DoWrapsFn(t *testing.T) {
	b := resilience.New(resilience.Config{})
	if err := b.Do(context.Background(), func(ctx context.Context) error { return nil }); err != nil {
		t.Fatalf("Do(nil fn): %v", err)
	}
	want := errors.New("boom")
	if err := b.Do(context.Background(), func(ctx context.Context) error { return want }); !errors.Is(err, want) {
		t.Fatalf("Do failing fn: want %v, got %v", want, err)
	}
}

func TestBreaker_DoReturnsOpenErrorWhenTripped(t *testing.T) {
	b := resilience.New(resilience.Config{FailureThreshold: 1})
	b.RecordFailure()
	if err := b.Do(context.Background(), func(ctx context.Context) error { return nil }); !errors.Is(err, resilience.ErrBreakerOpen) {
		t.Fatalf("want ErrBreakerOpen, got %v", err)
	}
}

func TestBreaker_SuccessResetsFailureCounter(t *testing.T) {
	b := resilience.New(resilience.Config{FailureThreshold: 3})
	b.RecordFailure()
	b.RecordFailure()
	b.RecordSuccess()
	b.RecordFailure()
	b.RecordFailure()
	if got := b.State(); got != resilience.StateClosed {
		t.Fatalf("want closed (success reset counter), got %s", got)
	}
}

func TestBreaker_ConcurrentSafe(t *testing.T) {
	b := resilience.New(resilience.Config{FailureThreshold: 100})
	done := make(chan struct{}, 200)
	for i := 0; i < 100; i++ {
		go func() { _ = b.Allow(); b.RecordSuccess(); done <- struct{}{} }()
	}
	for i := 0; i < 100; i++ {
		go func() { b.RecordFailure(); done <- struct{}{} }()
	}
	for i := 0; i < 200; i++ {
		<-done
	}
	if got := b.State(); got != resilience.StateClosed && got != resilience.StateOpen {
		t.Fatalf("invalid state under concurrency: %s", got)
	}
}
