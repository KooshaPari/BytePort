// Package middleware: CSRF protection via double-submit cookie.
//
// Why double-submit: stateless, works behind any reverse proxy, no server
// session store needed. The same token is set as an HttpOnly+SameSite=Strict
// cookie AND expected in the X-CSRF-Token header. A forged cross-site
// request cannot read the cookie value, so it cannot supply the header.
//
// Safe methods (GET/HEAD/OPTIONS/TRACE) are exempt. State-changing methods
// (POST/PUT/PATCH/DELETE) require a matching pair.
//
// Rollout: enforcement is gated behind BP_CSRF_ENFORCE=1. Default (dev/test)
// is permissive so existing integration tests keep passing; production sets
// the flag. This is intentional — flipping it on in prod is a one-line env
// change, not a code change.
package middleware

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

const (
	csrfCookieName = "bp_csrf"
	csrfHeaderName = "X-CSRF-Token"
)

// CSRFToken issues a fresh token and sets it as a cookie. Mount this on the
// routes that bootstrap the SPA or on a dedicated GET /csrf/token endpoint.
func CSRFToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		tok, err := newCSRFToken()
		if err != nil {
			c.Error(err)
			c.Next()
			return
		}
		c.SetSameSite(http.SameSiteStrictMode)
		c.SetCookie(csrfCookieName, tok, 3600, "/", "", true, true)
		c.Set("csrf_token", tok)
		c.Next()
	}
}

// CSRFProtect enforces double-submit matching for unsafe methods.
func CSRFProtect() gin.HandlerFunc {
	enforce := os.Getenv("BP_CSRF_ENFORCE") == "1"
	return func(c *gin.Context) {
		if !enforce {
			c.Next()
			return
		}
		if isSafeMethod(c.Request.Method) {
			c.Next()
			return
		}
		cookie, err := c.Cookie(csrfCookieName)
		header := c.GetHeader(csrfHeaderName)
		if err != nil || cookie == "" || header == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "missing CSRF token",
				"code":  "CSRF_MISSING",
			})
			return
		}
		if subtle.ConstantTimeCompare([]byte(cookie), []byte(header)) != 1 {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "CSRF token mismatch",
				"code":  "CSRF_MISMATCH",
			})
			return
		}
		c.Next()
	}
}

func isSafeMethod(m string) bool {
	switch m {
	case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
		return true
	}
	return false
}

func newCSRFToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
