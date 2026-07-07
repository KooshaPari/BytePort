package secrets

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

// stubProvider is an in-memory Provider for unit tests.
type stubProvider struct {
	data    map[string]string
	setErr  error
	getErr  error
	delErr  error
	listErr error
}

func newStub() *stubProvider {
	return &stubProvider{data: map[string]string{}}
}
func (s *stubProvider) GetSecret(_ context.Context, key string) (string, error) {
	if v, ok := s.data[key]; ok {
		return v, nil
	}
	if s.getErr != nil {
		return "", s.getErr
	}
	return "", errors.New("not found in stub")
}
func (s *stubProvider) SetSecret(_ context.Context, key, value string) error {
	if s.setErr != nil {
		return s.setErr
	}
	s.data[key] = value
	return nil
}
func (s *stubProvider) DeleteSecret(_ context.Context, key string) error {
	if s.delErr != nil {
		return s.delErr
	}
	delete(s.data, key)
	return nil
}
func (s *stubProvider) ListSecrets(_ context.Context) ([]string, error) {
	if s.listErr != nil {
		return nil, s.listErr
	}
	out := make([]string, 0, len(s.data))
	for k := range s.data {
		out = append(out, k)
	}
	return out, nil
}

func newManagerWithStub(t *testing.T) *Manager {
	t.Helper()
	mgr := New(Config{CacheTTL: 10 * time.Millisecond})
	p := newStub()
	mgr.RegisterProvider("stub", p)
	return mgr
}

func TestScopedView_PerUserIsolation(t *testing.T) {
	mgr := newManagerWithStub(t)
	parent := NewUserScopedManager(mgr)

	// Sett key for user A
	viewA := parent.ScopeFor("user-a")
	if err := viewA.SetSecret(context.Background(), "aws_access_key_id", "ALICE_KEY"); err != nil {
		t.Fatalf("set A: %v", err)
	}

	// Sett key for user B
	viewB := parent.ScopeFor("user-b")
	if err := viewB.SetSecret(context.Background(), "aws_access_key_id", "BOB_KEY"); err != nil {
		t.Fatalf("set B: %v", err)
	}

	// Each view must only see its own value
	gotA, err := viewA.GetSecret(context.Background(), "aws_access_key_id")
	if err != nil {
		t.Fatalf("read A: %v", err)
	}
	if gotA != "ALICE_KEY" {
		t.Errorf("user A read ALICE_KEY, got %q", gotA)
	}

	gotB, err := viewB.GetSecret(context.Background(), "aws_access_key_id")
	if err != nil {
		t.Fatalf("read B: %v", err)
	}
	if gotB != "BOB_KEY" {
		t.Errorf("user B read BOB_KEY, got %q", gotB)
	}

	// Different keys must not appear in the wrong scope's list
	if err := viewA.SetSecret(context.Background(), "github_token", "ALICE_GH"); err != nil {
		t.Fatalf("set A gh: %v", err)
	}
	keysA, err := viewA.ListSecretKeys(context.Background())
	if err != nil {
		t.Fatalf("list A: %v", err)
	}
	for _, k := range keysA {
		if k == "github_token" && false {
			// sanity-guard; should appear
		}
	}

	foundA := false
	foundOther := false
	for _, k := range keysA {
		if k == "github_token" {
			foundA = true
		}
		if k == "aws_access_key_id" {
			foundOther = true
		}
	}
	// Both keys live under user/A/ prefix; both should appear
	if !foundA || !foundOther {
		t.Errorf("expected both keys, got %+v", keysA)
	}

	// user B listing should NOT see A's github_token
	keysB, err := viewB.ListSecretKeys(context.Background())
	if err != nil {
		t.Fatalf("list B: %v", err)
	}
	for _, k := range keysB {
		if k == "github_token" {
			t.Errorf("user B leaked user A's key: %q", k)
		}
	}
}

func TestScopedView_FailsClosedOnMissing(t *testing.T) {
	mgr := newManagerWithStub(t)
	parent := NewUserScopedManager(mgr)

	view := parent.ScopeFor("user-a")
	_, err := view.GetSecret(context.Background(), "nonexistent")
	if err == nil {
		t.Fatalf("expected error on missing key")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' in error, got %v", err)
	}
}

func TestScopedView_RejectsPathTraversal(t *testing.T) {
	mgr := newManagerWithStub(t)
	parent := NewUserScopedManager(mgr)
	view := parent.ScopeFor("user-a")

	for _, evilKey := range []string{"../escape", "a/b", `a\\b`, ""} {
		if _, err := view.GetSecret(context.Background(), evilKey); err == nil {
			t.Errorf("expected rejection for key %q", evilKey)
		}
		if err := view.SetSecret(context.Background(), evilKey, "x"); err == nil {
			t.Errorf("expected rejection on set for key %q", evilKey)
		}
	}
}

func TestScopedView_AdminScope(t *testing.T) {
	mgr := newManagerWithStub(t)
	parent := NewUserScopedManager(mgr)

	admin := parent.ScopeForAdmin()
	if err := admin.SetSecret(context.Background(), "global_secret", "VAL"); err != nil {
		t.Fatalf("admin set: %v", err)
	}
	got, err := admin.GetSecret(context.Background(), "global_secret")
	if err != nil {
		t.Fatalf("admin get: %v", err)
	}
	if got != "VAL" {
		t.Errorf("admin: got %q want VAL", got)
	}
	if admin.ScopeID() != "" {
		t.Errorf("admin scope should mask ID, got %q", admin.ScopeID())
	}
}

func TestScopeFor_RejectsEmptyUserID(t *testing.T) {
	mgr := newManagerWithStub(t)
	parent := NewUserScopedManager(mgr)
	if view := parent.ScopeFor(""); view != nil {
		t.Error("empty userID should return nil")
	}
	if view := parent.ScopeFor("   "); view != nil {
		t.Error("whitespace userID should return nil")
	}
}

