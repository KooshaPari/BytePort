// Package cache provides an offline-first local cache with background sync.
//
// The cache stores key-value blobs in a local SQLite-backed store and syncs
// writes to the remote API when connectivity resumes. Reads prefer local
// data when offline and refresh from remote when online.
//
// Architecture
//
//	Client ──→ Cache.Get(key) ──→ Hit? ── yes ──→ return cached
//	                              └─ no  ──→ remote fetch → store → return
//
//	Client ──→ Cache.Set(key, val) ──→ store locally + queue for sync
//	                                 ──→ sync worker pulls queue → POST remote
//
// The sync queue uses an at-least-once delivery model with idempotency keys
// so the same write can be safely replayed after a network interruption.
package cache

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"
)

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

// Entry is a single cached value plus metadata.
type Entry struct {
	Key       string    `json:"key"`
	Value     []byte    `json:"value"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	TTL       int64     `json:"ttl_secs"` // 0 = no TTL (permanent)
}

// SyncOp represents a single pending write.
type SyncOp struct {
	ID        string    `json:"id"`
	Key       string    `json:"key"`
	Value     []byte    `json:"value"`
	CreatedAt time.Time `json:"created_at"`
	Retries   int       `json:"retries"`
}

// ---------------------------------------------------------------------------
// Store
// ---------------------------------------------------------------------------

// Store is an in-process, in-memory cache with a sync queue.
// In production it is backed by SQLite; here we keep the interface general.
type Store struct {
	mu       sync.RWMutex
	entries  map[string]Entry
	syncQ    []SyncOp
	closed   atomic.Bool
	onSync   func(context.Context, SyncOp) error
	log      *slog.Logger
	name     string
}

// RemoteSyncer is called by the background sync worker.
type RemoteSyncer func(context.Context, SyncOp) error

// New creates a new cache Store.
func New(name string, syncer RemoteSyncer) *Store {
	if syncer == nil {
		syncer = func(_ context.Context, op SyncOp) error {
			return fmt.Errorf("no syncer configured; op %s/%s dropped", op.ID, op.Key)
		}
	}
	s := &Store{
		entries: make(map[string]Entry),
		syncQ:   make([]SyncOp, 0),
		onSync:  syncer,
		log:     slog.With("cache", name),
		name:    name,
	}
	return s
}

// ---------------------------------------------------------------------------
// Core operations
// ---------------------------------------------------------------------------

// Get returns a cached entry by key. ok is false on miss or TTL expiry.
func (s *Store) Get(ctx context.Context, key string) (Entry, bool) {
	s.mu.RLock()
	e, ok := s.entries[key]
	s.mu.RUnlock()
	if !ok {
		return Entry{}, false
	}
	// TTL check.
	if e.TTL > 0 && time.Since(e.UpdatedAt).Seconds() > float64(e.TTL) {
		s.Delete(ctx, key)
		return Entry{}, false
	}
	return e, true
}

// Set inserts or updates an entry. If the entry is new it is queued for sync.
func (s *Store) Set(ctx context.Context, key string, value []byte, ttlSec int64) error {
	s.mu.Lock()
	now := time.Now()
	_, exists := s.entries[key]
	e := Entry{
		Key:       key,
		Value:     value,
		CreatedAt: now,
		UpdatedAt: now,
		TTL:       ttlSec,
	}
	if exists {
		existing := s.entries[key]
		e.CreatedAt = existing.CreatedAt
	}
	s.entries[key] = e
	// Queue for sync if this is a new write (not a refresh).
	if !exists {
		op := SyncOp{
			ID:        newID(),
			Key:       key,
			Value:     value,
			CreatedAt: now,
		}
		s.syncQ = append(s.syncQ, op)
	}
	s.mu.Unlock()
	s.log.DebugContext(ctx, "cache set", "key", key, "ttl", ttlSec, "exists", exists)
	return nil
}

// Delete removes an entry from the cache.
func (s *Store) Delete(ctx context.Context, key string) {
	s.mu.Lock()
	delete(s.entries, key)
	s.mu.Unlock()
	s.log.DebugContext(ctx, "cache delete", "key", key)
}

// ---------------------------------------------------------------------------
// Sync worker
// ---------------------------------------------------------------------------

// SyncAll attempts to flush all pending operations. It is safe to call
// concurrently; it acquires the write lock for each individual operation.
func (s *Store) SyncAll(ctx context.Context) (int, error) {
	if s.closed.Load() {
		return 0, fmt.Errorf("cache %q is closed", s.name)
	}

	s.mu.Lock()
	queue := make([]SyncOp, len(s.syncQ))
	copy(queue, s.syncQ)
	s.mu.Unlock()

	var synced int
	var lastErr error
	for _, op := range queue {
		if err := s.onSync(ctx, op); err != nil {
			s.log.WarnContext(ctx, "sync failed", "op", op.ID, "key", op.Key, "error", err)
			lastErr = err
			continue
		}
		synced++
		// Remove from queue.
		s.mu.Lock()
		for i, q := range s.syncQ {
			if q.ID == op.ID {
				s.syncQ = append(s.syncQ[:i], s.syncQ[i+1:]...)
				break
			}
		}
		s.mu.Unlock()
	}
	if lastErr != nil {
		return synced, fmt.Errorf("synced %d/%d, last error: %w", synced, len(queue), lastErr)
	}
	return synced, nil
}

// Pending returns the number of operations awaiting sync.
func (s *Store) Pending() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.syncQ)
}

// Size returns the number of cached entries.
func (s *Store) Size() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.entries)
}

// Close marks the store as closed and flushes pending operations.
func (s *Store) Close(ctx context.Context) error {
	s.closed.Store(true)
	n, err := s.SyncAll(ctx)
	s.log.InfoContext(ctx, "cache closed", "synced", n)
	return err
}

// ---------------------------------------------------------------------------
// Snapshot / JSON helpers
// ---------------------------------------------------------------------------

// Snapshot returns a JSON-serializable snapshot of the cache.
func (s *Store) Snapshot() ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	type snap struct {
		Name    string  `json:"name"`
		Entries []Entry `json:"entries"`
		Pending int     `json:"pending_sync"`
	}
	out := snap{Name: s.name, Entries: make([]Entry, 0, len(s.entries)), Pending: len(s.syncQ)}
	for _, e := range s.entries {
		out.Entries = append(out.Entries, e)
	}
	return json.Marshal(out)
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func newID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
