package lib

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"os"

	"github.com/byteport/api/internal/infrastructure/auth"
	httpmiddleware "github.com/byteport/api/internal/infrastructure/http/middleware"
	"github.com/byteport/api/internal/infrastructure/secrets"
	"github.com/gin-gonic/gin"
)

const encryptionKeyEnv = "ENCRYPTION_KEY"

// InitializeEncryptionKey ensures the API process has an AES-compatible key.
func InitializeEncryptionKey() error {
	if os.Getenv(encryptionKeyEnv) != "" {
		return nil
	}

	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return fmt.Errorf("generate encryption key: %w", err)
	}

	if err := os.Setenv(encryptionKeyEnv, base64.StdEncoding.EncodeToString(key)); err != nil {
		return fmt.Errorf("set encryption key: %w", err)
	}

	return nil
}

// InitAuthSystem is retained for the legacy server startup path.
func InitAuthSystem() error {
	return nil
}

// AuthMiddleware returns the API's auth middleware.
// When WORKOS_API_KEY is set, it uses full WorkOS JWT validation.
// Otherwise it falls back to test-token-only validation for local dev.
func AuthMiddleware() gin.HandlerFunc {
	if os.Getenv("WORKOS_API_KEY") != "" {
		mgr := secrets.New(secrets.Config{})
		mgr.RegisterProvider("env", secrets.NewEnvironmentProvider())
		svc := auth.NewWorkOSAuthService(mgr)
		if err := svc.Initialize(context.Background()); err != nil {
			log.Printf("WARNING: WorkOS auth init failed, falling back to dev mode: %v", err)
		} else {
			return httpmiddleware.AuthMiddleware(svc)
		}
	}
	return httpmiddleware.AuthMiddlewareWithFallback(nil)
}
