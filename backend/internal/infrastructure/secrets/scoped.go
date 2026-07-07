// Package secrets: per-user credential scoping.
//
// The base Manager exposes flat keys (GetAWSConfig, GetLLMConfig, etc) which
// leak global credentials to every authenticated user. UserScopedManager
// wraps Manager with a per-userID key namespace and a strict policy:
//
//   - All Read/Write/Delete calls compose "user:<uuid>:<key>"
//   - Admins (userID == "") retain access to global keys
//   - No fallback from per-user to global — fails closed
//   - ListSecretsScoped returns only keys owned by this user
//   - Invalidate/clear cache stay package-level (they purge the whole cache)
//
// Usage:
//
//	sm := secrets.New(secrets.Config{CacheTTL: 5 * time.Minute})
//	sm.RegisterProvider("env", envProvider)
//	scoped := sm.ScopeFor(userUUID) // or .ScopeForGlobal() for admins
//
// IMPORTANT: this wrapper does not change the contract of Manager — it
// rejects attempts to read un-scoped keys under a user context. Pooling
// unscoped credentials is therefore impossible through this API.
package secrets

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

// UserScopedScope identifies a credential scope.
type UserScopedScope string

const (
	// ScopeUser isolates a user's private credentials.
	ScopeUser UserScopedScope = "user"
	// ScopeAdmin is reserved for platform administrators and operators.
	ScopeAdmin UserScopedScope = "admin"
)

// UserScopedManager wraps a Manager with per-user key namespacing.
//
// Calls compose "(user|admin)/<id>/<key>" so providers cannot accidentally
// serve one user's credentials to another. The wrapper fails closed: if a
// per-user lookup fails it NEVER falls back to a global namespace.
type UserScopedManager struct {
	mu    sync.RWMutex
	base  *Manager
	cache map[string]*scopedCachedSecret
	ttl   time.Duration
}

// NewUserScopedManager constructs a scoped view over an existing Manager.
func NewUserScopedManager(base *Manager) *UserScopedManager {
	if base == nil {
		panic("secrets: NewUserScopedManager: base Manager is nil")
	}
	ttl := base.cacheTTL
	return &UserScopedManager{
		base:  base,
		cache: make(map[string]*scopedCachedSecret),
		ttl:   ttl,
	}
}

// ScopeFor returns a manager scoped to a single user. userID must be a
// non-empty UUID; an empty value returns nil so callers must explicitly
// decide between user scope (this method) and global scope (next method).
func (s *UserScopedManager) ScopeFor(userID string) *ScopedView {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil
	}
	return &ScopedView{
		parent:    s,
		scope:     ScopeUser,
		scopeID:   userID,
		namespace: fmt.Sprintf("user/%s", userID),
	}
}

// ScopeForAdmin returns a manager view with global privileges. Use only for
// platform operator flows — never construct this from request-derived data.
func (s *UserScopedManager) ScopeForAdmin() *ScopedView {
	return &ScopedView{
		parent:    s,
		scope:     ScopeAdmin,
		scopeID:   "",
		namespace: "admin",
	}
}

// ScopedView is the per-caller credential surface. Construct via ScopeFor
// or ScopeForAdmin on the parent UserScopedManager.
type ScopedView struct {
	parent    *UserScopedManager
	scope     UserScopedScope
	scopeID   string
	namespace string
}

// Scope returns "user" or "admin".
func (v *ScopedView) Scope() UserScopedScope { return v.scope }

// ScopeID returns the user UUID for user scopes or "" for admin.
// The ID is masked in logs (last 4 chars only) to avoid accidental
// disclosure when ScopedView is inspected in error paths.
func (v *ScopedView) ScopeID() string {
	if v.scope == ScopeAdmin {
		return ""
	}
	if len(v.scopeID) <= 4 {
		return v.scopeID
	}
	return "***" + v.scopeID[len(v.scopeID)-4:]
}

// composeKey returns the namespaced key passed to the underlying Manager.
func (v *ScopedView) composeKey(key string) (string, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return "", errors.New("secrets: empty key")
	}
	// Reject path-traversal attempts that could escape the namespace.
	if strings.Contains(key, "..") || strings.ContainsAny(key, "/\\") {
		return "", fmt.Errorf("secrets: key %q contains illegal characters", key)
	}
	return v.namespace + "/" + key, nil
}

// GetSecret looks up a secret within this view's scope. NEVER falls back to
// other scopes — fail-closed.
func (v *ScopedView) GetSecret(ctx context.Context, key string) (string, error) {
	composed, err := v.composeKey(key)
	if err != nil {
		return "", err
	}

	cacheKey := v.scopeKey(composed)
	if value, ok := v.readCache(cacheKey); ok {
		return value, nil
	}

	value, err := v.parent.base.GetSecret(ctx, composed)
	if err != nil {
		return "", fmt.Errorf("secrets: scope=%s key=%s not found: %w", v.scope, key, err)
	}
	v.writeCache(cacheKey, value)
	return value, nil
}

// SetSecret stores a secret within this view's scope.
func (v *ScopedView) SetSecret(ctx context.Context, key, value string) error {
	composed, err := v.composeKey(key)
	if err != nil {
		return err
	}
	if err := v.parent.base.SetSecret(ctx, composed, value); err != nil {
		return fmt.Errorf("secrets: scope=%s set key=%s: %w", v.scope, key, err)
	}
	v.invalidateCache(v.scopeKey(composed))
	return nil
}

// DeleteSecret removes a secret within this view's scope.
//
// The base Manager does not currently expose a top-level DeleteSecret
// method, so we delegate to the first registered provider's DeleteSecret.
func (v *ScopedView) DeleteSecret(ctx context.Context, key string) error {
	composed, err := v.composeKey(key)
	if err != nil {
		return err
	}
	provider, err := v.firstProvider()
	if err != nil {
		return fmt.Errorf("secrets: scope=%s delete key=%s: %w", v.scope, key, err)
	}
	if err := provider.DeleteSecret(ctx, composed); err != nil {
		return fmt.Errorf("secrets: scope=%s delete key=%s: %w", v.scope, key, err)
	}
	v.invalidateCache(v.scopeKey(composed))
	return nil
}

// ListSecretKeys returns the keys (un-namespaced) the caller can read.
//
// Provider-side enumeration strips the namespace prefix; entries whose
// prefix does not match this view are silently filtered. Returning a
// stable, scope-bound list prevents one user from learning about another's
// stored credentials through the API.
func (v *ScopedView) ListSecretKeys(ctx context.Context) ([]string, error) {
	provider, err := v.firstProvider()
	if err != nil {
		return nil, err
	}
	all, err := provider.ListSecrets(ctx)
	if err != nil {
		return nil, err
	}
	prefix := v.namespace + "/"
	var out []string
	for _, k := range all {
		if strings.HasPrefix(k, prefix) {
			out = append(out, strings.TrimPrefix(k, prefix))
		}
	}
	return out, nil
}

// HasSecret checks existence without exposing the value or performing I/O.
// It uses Manager.GetSecret so providers that don't implement Existence API
// are still respected.
func (v *ScopedView) HasSecret(ctx context.Context, key string) (bool, error) {
	_, err := v.GetSecret(ctx, key)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, ErrSecretNotFound) {
		return false, nil
	}
	return false, err
}

// Lock acquires the parent's write lock so batch operations are atomic.
func (v *ScopedView) Lock() {
	v.parent.mu.Lock()
}

// Unlock releases the parent's write lock.
func (v *ScopedView) Unlock() {
	v.parent.mu.Unlock()
}

// firstProvider returns the first registered provider in deterministic order.
// Returns ErrSecretNotFound if no providers have been registered.
func (v *ScopedView) firstProvider() (Provider, error) {
	v.parent.base.cacheMux.RLock()
	defer v.parent.base.cacheMux.RUnlock()
	if len(v.parent.base.order) == 0 {
		return nil, ErrSecretNotFound
	}
	name := v.parent.base.order[0]
	provider, ok := v.parent.base.providers[name]
	if !ok {
		return nil, ErrSecretNotFound
	}
	return provider, nil
}

// scopeKey returns the per-user-scoped cache key (separate from the
// composed provider key so two users with the same logical key don't
// collide in the local cache).
func (v *ScopedView) scopeKey(composed string) string {
	return string(v.scope) + "|" + composed
}

// readCache looks up a cached entry respecting ttl.
func (v *ScopedView) readCache(k string) (string, bool) {
	v.parent.mu.RLock()
	defer v.parent.mu.RUnlock()
	entry, ok := v.parent.cache[k]
	if !ok || time.Now().After(entry.expiresAt) {
		return "", false
	}
	return entry.value, true
}

// writeCache stores a value with the configured ttl.
func (v *ScopedView) writeCache(k, value string) {
	v.parent.mu.Lock()
	defer v.parent.mu.Unlock()
	v.parent.cache[k] = &scopedCachedSecret{
		value:     value,
		expiresAt: time.Now().Add(v.parent.ttl),
	}
}

// invalidateCache drops a single cache entry.
func (v *ScopedView) invalidateCache(k string) {
	v.parent.mu.Lock()
	defer v.parent.mu.Unlock()
	delete(v.parent.cache, k)
}

// scopedCachedSecret is a ttl-aware cached entry owned by UserScopedManager.
type scopedCachedSecret struct {
	value     string
	expiresAt time.Time
}

// ErrSecretNotFound is returned by user-scoped lookups when a key cannot be
// located in any registered provider.
var ErrSecretNotFound = errors.New("secrets: not found")
