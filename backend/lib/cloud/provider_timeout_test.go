package cloud

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"
)

type blockingRoundTripper struct{}

func (blockingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	<-req.Context().Done()
	return nil, req.Context().Err()
}

func TestVercelValidateCredentialsHonorsCallerDeadline(t *testing.T) {
	provider, err := NewVercelProvider(Credentials{Data: map[string]string{"token": "test-token"}})
	if err != nil {
		t.Fatalf("construct provider: %v", err)
	}
	vercel := provider.(*VercelProvider)
	vercel.httpClient = &http.Client{Transport: blockingRoundTripper{}}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	err = vercel.ValidateCredentials(ctx)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline exceeded, got %v", err)
	}
}
