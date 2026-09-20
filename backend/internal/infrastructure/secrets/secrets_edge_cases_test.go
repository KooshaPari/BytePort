package secrets

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestProviderFunctionSignatures verifies that each Provider implementation
// exposes the four interface methods with the expected signatures.
func TestProviderFunctionSignatures(t *testing.T) {
	tests := []struct {
		name     string
		provider Provider
	}{
		{"VaultProvider", &VaultProvider{path: "secret/test"}},
		{"AWSSecretsProvider", &AWSSecretsProvider{region: "us-east-1"}},
		{"EnvironmentProvider", &EnvironmentProvider{}},
	}

	for _, tc := range tests {
		t.Run(tc.name+" function signatures", func(t *testing.T) {
			assert.NotNil(t, tc.provider.GetSecret)
			assert.NotNil(t, tc.provider.SetSecret)
			assert.NotNil(t, tc.provider.DeleteSecret)
			assert.NotNil(t, tc.provider.ListSecrets)

			// Compile-time signature checks: assigning a method value to a
			// typed func var forces the compiler to verify the signature.
			var getSecretFunc func(context.Context, string) (string, error) = tc.provider.GetSecret
			var setSecretFunc func(context.Context, string, string) error = tc.provider.SetSecret
			var deleteSecretFunc func(context.Context, string) error = tc.provider.DeleteSecret
			var listSecretsFunc func(context.Context) ([]string, error) = tc.provider.ListSecrets

			assert.NotNil(t, getSecretFunc)
			assert.NotNil(t, setSecretFunc)
			assert.NotNil(t, deleteSecretFunc)
			assert.NotNil(t, listSecretsFunc)
		})
	}
}

// TestManagerFunctionSignatures verifies that Manager exposes its lifecycle
// methods with the expected signatures.
func TestManagerFunctionSignatures(t *testing.T) {
	t.Run("Manager function signatures", func(t *testing.T) {
		manager := &Manager{}

		assert.NotNil(t, manager.GetSecret)
		assert.NotNil(t, manager.SetSecret)
		assert.NotNil(t, manager.InvalidateCache)
		assert.NotNil(t, manager.ClearCache)
		assert.NotNil(t, manager.RegisterProvider)

		var getSecretFunc func(context.Context, string) (string, error) = manager.GetSecret
		var setSecretFunc func(context.Context, string, string) error = manager.SetSecret
		var invalidateCacheFunc func(string) = manager.InvalidateCache
		var clearCacheFunc func() = manager.ClearCache
		var registerProviderFunc func(string, Provider) = manager.RegisterProvider

		assert.NotNil(t, getSecretFunc)
		assert.NotNil(t, setSecretFunc)
		assert.NotNil(t, invalidateCacheFunc)
		assert.NotNil(t, clearCacheFunc)
		assert.NotNil(t, registerProviderFunc)
	})
}

// TestProviderInterfaceCompliance verifies that all provider types satisfy
// the Provider interface.
func TestProviderInterfaceCompliance(t *testing.T) {
	tests := []struct {
		name     string
		provider Provider
	}{
		{"VaultProvider implements Provider interface", &VaultProvider{}},
		{"AWSSecretsProvider implements Provider interface", &AWSSecretsProvider{}},
		{"EnvironmentProvider implements Provider interface", &EnvironmentProvider{}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var provider Provider = tc.provider
			assert.NotNil(t, provider)
		})
	}
}

// TestNewFunctions verifies that each constructor exists and returns the
// documented type.
func TestNewFunctions(t *testing.T) {
	t.Run("New function exists and returns correct type", func(t *testing.T) {
		config := Config{}
		manager := New(config)
		assert.NotNil(t, manager)
		assert.IsType(t, &Manager{}, manager)
	})

	t.Run("NewAWSSecretsProvider function exists", func(t *testing.T) {
		assert.NotNil(t, NewAWSSecretsProvider)
		funcType := func(context.Context, string) (*AWSSecretsProvider, error) { return nil, nil }
		assert.IsType(t, funcType, NewAWSSecretsProvider)
	})

	t.Run("NewVaultProvider function exists", func(t *testing.T) {
		assert.NotNil(t, NewVaultProvider)
		funcType := func(string, string, string) (*VaultProvider, error) { return nil, nil }
		assert.IsType(t, funcType, NewVaultProvider)
	})

	t.Run("NewEnvironmentProvider function exists", func(t *testing.T) {
		provider := NewEnvironmentProvider()
		assert.NotNil(t, provider)
		assert.IsType(t, &EnvironmentProvider{}, provider)
	})
}

// TestEdgeCases exercises Manager with empty providers and edge-case keys.
func TestEdgeCases(t *testing.T) {
	keyTests := []struct {
		name string
		key  string
	}{
		{"Manager with nil provider", "nonexistent:key"},
		{"Manager with empty key", ""},
		{"Manager with invalid key format", "invalid-key"},
	}
	for _, tc := range keyTests {
		t.Run(tc.name, func(t *testing.T) {
			manager := newEmptyManager()

			_, err := manager.GetSecret(context.Background(), tc.key)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "not found in any provider")

			err = manager.SetSecret(context.Background(), tc.key, "value")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "failed to store secret")
		})
	}

	t.Run("Manager cache operations", func(t *testing.T) {
		manager := &Manager{
			providers: make(map[string]Provider),
			cache:     make(map[string]*cachedSecret),
		}

		assert.NotPanics(t, func() {
			manager.InvalidateCache("")
		})
		assert.NotPanics(t, func() {
			manager.InvalidateCache("nonexistent:key")
		})
		assert.NotPanics(t, func() {
			manager.ClearCache()
		})
	})

	t.Run("Manager RegisterProvider", func(t *testing.T) {
		manager := newEmptyManager()

		assert.NotPanics(t, func() {
			manager.RegisterProvider("test", nil)
		})

		envProvider := &EnvironmentProvider{}
		assert.NotPanics(t, func() {
			manager.RegisterProvider("env", envProvider)
		})
		assert.NotNil(t, manager.providers["env"])
	})
}

// TestContextHandling verifies that Manager respects context cancellation
// and timeouts when no providers are registered.
func TestContextHandling(t *testing.T) {
	t.Run("Context cancellation", func(t *testing.T) {
		manager := newEmptyManager()

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := manager.GetSecret(ctx, "test:key")
		assert.Error(t, err)

		err = manager.SetSecret(ctx, "test:key", "value")
		assert.Error(t, err)
	})

	t.Run("Context timeout", func(t *testing.T) {
		manager := newEmptyManager()

		ctx, cancel := context.WithTimeout(context.Background(), 0)
		defer cancel()

		_, err := manager.GetSecret(ctx, "test:key")
		assert.Error(t, err)

		err = manager.SetSecret(ctx, "test:key", "value")
		assert.Error(t, err)
	})
}

// newEmptyManager returns a Manager with an empty providers map. Used by
// edge-case tests that need to assert error paths without any backing
// providers.
func newEmptyManager() *Manager {
	return &Manager{providers: make(map[string]Provider)}
}
