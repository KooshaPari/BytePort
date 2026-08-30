package auth

import (
	"testing"
	"time"

	"aidanwoods.dev/go-paseto"
)

// TestClaimsExtractionFromValidToken verifies that all standard claims and
// custom claims can be extracted from a valid PASETO v4 local token.
func TestClaimsExtractionFromValidToken(t *testing.T) {
	key := paseto.NewV4SymmetricKey()

	now := time.Now()
	token := paseto.NewToken()
	token.SetAudience("user@example.com")
	token.SetExpiration(now.Add(time.Hour))
	token.SetSubject("session")
	token.SetIssuer("BytePort")
	token.SetIssuedAt(now)
	token.SetNotBefore(now)
	token.SetString("user-id", "uuid-abc-123")
	token.SetString("project-id", "proj-456")

	encrypted := token.V4Encrypt(key, nil)

	parser := paseto.NewParser()
	decrypted, err := parser.ParseV4Local(key, encrypted, nil)
	if err != nil {
		t.Fatalf("ParseV4Local returned error: %v", err)
	}

	// Verify audience
	audience, err := decrypted.GetAudience()
	if err != nil {
		t.Fatalf("GetAudience returned error: %v", err)
	}
	if audience != "user@example.com" {
		t.Errorf("audience = %q, want %q", audience, "user@example.com")
	}

	// Verify subject
	subject, err := decrypted.GetSubject()
	if err != nil {
		t.Fatalf("GetSubject returned error: %v", err)
	}
	if subject != "session" {
		t.Errorf("subject = %q, want %q", subject, "session")
	}

	// Verify issuer
	issuer, err := decrypted.GetIssuer()
	if err != nil {
		t.Fatalf("GetIssuer returned error: %v", err)
	}
	if issuer != "BytePort" {
		t.Errorf("issuer = %q, want %q", issuer, "BytePort")
	}

	// Verify issued-at is within a reasonable window
	iat, err := decrypted.GetIssuedAt()
	if err != nil {
		t.Fatalf("GetIssuedAt returned error: %v", err)
	}
	if iat.Before(now.Add(-time.Second)) || iat.After(now.Add(time.Second)) {
		t.Errorf("issued-at = %v, want within 1s of %v", iat, now)
	}

	// Verify expiration is in the future
	exp, err := decrypted.GetExpiration()
	if err != nil {
		t.Fatalf("GetExpiration returned error: %v", err)
	}
	if exp.Before(now) {
		t.Errorf("expiration = %v, should be in the future", exp)
	}

	// Verify custom claims
	userID, err := decrypted.GetString("user-id")
	if err != nil {
		t.Fatalf("GetString(user-id) returned error: %v", err)
	}
	if userID != "uuid-abc-123" {
		t.Errorf("user-id = %q, want %q", userID, "uuid-abc-123")
	}

	projectID, err := decrypted.GetString("project-id")
	if err != nil {
		t.Fatalf("GetString(project-id) returned error: %v", err)
	}
	if projectID != "proj-456" {
		t.Errorf("project-id = %q, want %q", projectID, "proj-456")
	}
}

// TestMissingOrInvalidTokenReturnsError verifies that parsing a garbled or
// empty token string always produces an error.
func TestMissingOrInvalidTokenReturnsError(t *testing.T) {
	key := paseto.NewV4SymmetricKey()
	parser := paseto.NewParser()

	tests := []struct {
		name  string
		token string
	}{
		{"empty string", ""},
		{"random text", "not-a-paseto-token"},
		{"partial token", "v4.local."},
		{"base64 garbage", "v4.local.!!!invalid-base64!!!"},
		{"public token format", "v4.public.dGVzdA"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parser.ParseV4Local(key, tt.token, nil)
			if err == nil {
				t.Errorf("ParseV4Local should return error for %q", tt.name)
			}
		})
	}
}

// TestOwnerIDExtractionFromClaims verifies that the user-id custom claim
// (used as the owner identifier) can be reliably extracted.
func TestOwnerIDExtractionFromClaims(t *testing.T) {
	key := paseto.NewV4SymmetricKey()

	ownerIDs := []string{
		"550e8400-e29b-41d4-a716-446655440000",
		"simple-id",
		"owner-with-special_chars.123",
	}

	for _, ownerID := range ownerIDs {
		token := paseto.NewToken()
		token.SetExpiration(time.Now().Add(time.Hour))
		token.SetString("user-id", ownerID)

		encrypted := token.V4Encrypt(key, nil)

		parser := paseto.NewParser()
		decrypted, err := parser.ParseV4Local(key, encrypted, nil)
		if err != nil {
			t.Fatalf("ParseV4Local returned error for ownerID %q: %v", ownerID, err)
		}

		extractedID, err := decrypted.GetString("user-id")
		if err != nil {
			t.Fatalf("GetString(user-id) returned error: %v", err)
		}
		if extractedID != ownerID {
			t.Errorf("user-id = %q, want %q", extractedID, ownerID)
		}
	}
}

// TestOwnerIDMissingReturnsError verifies that requesting a user-id claim
// that does not exist returns an error.
func TestOwnerIDMissingReturnsError(t *testing.T) {
	key := paseto.NewV4SymmetricKey()

	token := paseto.NewToken()
	token.SetExpiration(time.Now().Add(time.Hour))
	// Deliberately not setting user-id

	encrypted := token.V4Encrypt(key, nil)

	parser := paseto.NewParser()
	decrypted, err := parser.ParseV4Local(key, encrypted, nil)
	if err != nil {
		t.Fatalf("ParseV4Local returned error: %v", err)
	}

	_, err = decrypted.GetString("user-id")
	if err == nil {
		t.Error("GetString(user-id) should return error when claim is absent")
	}
}

// TestTokenWithWrongKeyCannotBeDecrypted verifies that a token encrypted
// with key A cannot be decrypted with key B.
func TestTokenWithWrongKeyCannotBeDecrypted(t *testing.T) {
	keyA := paseto.NewV4SymmetricKey()
	keyB := paseto.NewV4SymmetricKey()

	token := paseto.NewToken()
	token.SetExpiration(time.Now().Add(time.Hour))
	token.SetString("user-id", "owner-123")

	encrypted := token.V4Encrypt(keyA, nil)

	parser := paseto.NewParser()
	_, err := parser.ParseV4Local(keyB, encrypted, nil)
	if err == nil {
		t.Error("ParseV4Local with wrong key should return error")
	}
}

// TestTokenAudienceValidation verifies that tokens can be filtered by
// audience claim, as the service token validator does.
func TestTokenAudienceValidation(t *testing.T) {
	key := paseto.NewV4SymmetricKey()

	// Create token with specific audience
	token := paseto.NewToken()
	token.SetExpiration(time.Now().Add(time.Hour))
	token.SetAudience("NVMService")
	token.SetString("user-id", "user-1")

	encrypted := token.V4Encrypt(key, nil)

	// Parse with audience check
	parser := paseto.NewParser()
	parser.AddRule(paseto.ForAudience("NVMService"))

	decrypted, err := parser.ParseV4Local(key, encrypted, nil)
	if err != nil {
		t.Fatalf("ParseV4Local with ForAudience should succeed: %v", err)
	}

	audience, err := decrypted.GetAudience()
	if err != nil {
		t.Fatalf("GetAudience returned error: %v", err)
	}
	if audience != "NVMService" {
		t.Errorf("audience = %q, want %q", audience, "NVMService")
	}

	// Parse with wrong audience should fail
	parser2 := paseto.NewParser()
	parser2.AddRule(paseto.ForAudience("DifferentService"))

	_, err = parser2.ParseV4Local(key, encrypted, nil)
	if err == nil {
		t.Error("ParseV4Local with wrong audience should fail")
	}
}

// TestExpiredTokenRejectedByParser verifies that a token whose expiration
// is in the past is rejected when NotExpired rule is applied.
func TestExpiredTokenRejectedByParser(t *testing.T) {
	key := paseto.NewV4SymmetricKey()

	token := paseto.NewToken()
	token.SetExpiration(time.Now().Add(-time.Hour))
	token.SetString("user-id", "expired-user")

	encrypted := token.V4Encrypt(key, nil)

	parser := paseto.NewParser()
	parser.AddRule(paseto.NotExpired())

	_, err := parser.ParseV4Local(key, encrypted, nil)
	if err == nil {
		t.Error("ParseV4Local with NotExpired should reject expired token")
	}
}

// TestServiceTokenClaims verifies that a token carrying both user-id and
// project-id claims (the NVMS service token pattern) can be parsed correctly.
func TestServiceTokenClaims(t *testing.T) {
	key := paseto.NewV4SymmetricKey()

	token := paseto.NewToken()
	token.SetAudience("NVMService")
	token.SetExpiration(time.Now().Add(10 * time.Minute))
	token.SetSubject("deployment")
	token.SetIssuer("BytePort")
	token.SetString("user-id", "user-uuid-999")
	token.SetString("project-id", "project-uuid-888")

	encrypted := token.V4Encrypt(key, nil)

	parser := paseto.NewParser()
	parser.AddRule(paseto.ForAudience("NVMService"))
	parser.AddRule(paseto.NotExpired())

	decrypted, err := parser.ParseV4Local(key, encrypted, nil)
	if err != nil {
		t.Fatalf("ParseV4Local returned error: %v", err)
	}

	userID, err := decrypted.GetString("user-id")
	if err != nil {
		t.Fatalf("GetString(user-id) returned error: %v", err)
	}
	if userID != "user-uuid-999" {
		t.Errorf("user-id = %q, want %q", userID, "user-uuid-999")
	}

	projectID, err := decrypted.GetString("project-id")
	if err != nil {
		t.Fatalf("GetString(project-id) returned error: %v", err)
	}
	if projectID != "project-uuid-888" {
		t.Errorf("project-id = %q, want %q", projectID, "project-uuid-888")
	}
}
