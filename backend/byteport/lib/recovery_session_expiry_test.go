package lib

import (
    "testing"
    "time"

    "aidanwoods.dev/go-paseto"
)

// Mature-recovery adversarial oracle for session-token lifetime enforcement.
//
// GenerateToken writes an expiration claim. This test asks whether the actual
// ValidateToken path enforces that claim. No production remediation is
// included.
func TestRecoverySessionTokenRejectsExpiredToken(t *testing.T) {
    key := paseto.NewV4SymmetricKey()
    if err := keyringSet(tokenKeyService, keyringUser, key.ExportHex()); err != nil {
        t.Skipf("keyring unavailable for native auth oracle: %v", err)
    }

    token := paseto.NewToken()
    token.SetAudience("expired@example.invalid")
    token.SetSubject("session")
    token.SetIssuer("BytePort")
    token.SetIssuedAt(time.Now().Add(-2 * time.Hour))
    token.SetNotBefore(time.Now().Add(-2 * time.Hour))
    token.SetExpiration(time.Now().Add(-1 * time.Hour))
    token.SetString(claimUserID, "expired-user")

    encrypted := token.V4Encrypt(key, nil)
    valid, _, err := ValidateToken(encrypted)

    if err == nil && valid {
        t.Fatal("expired session token was accepted; ValidateToken must enforce token lifetime")
    }
}
