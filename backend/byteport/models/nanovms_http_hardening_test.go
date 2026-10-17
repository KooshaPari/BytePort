package models

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

type hardeningRT func(*http.Request) (*http.Response, error)

func (f hardeningRT) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func hardeningResponse(status int, body string) *http.Response {
	return &http.Response{StatusCode: status, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}
}
func hardeningRequest() NanoVMSDeployRequest {
	return NanoVMSDeployRequest{Name: "svc", Image: "sha256:artifact", OperationID: "op", ResourceID: "svc", ConfigDigest: "cfg"}
}
func hardeningTransport(status int, body string) NanoVMSHTTPTransport {
	return NanoVMSHTTPTransport{BaseURL: "http://fixture.invalid", Client: &http.Client{Transport: hardeningRT(func(*http.Request) (*http.Response, error) { return hardeningResponse(status, body), nil })}}
}
func hardeningUnknown(t *testing.T, err error) {
	t.Helper()
	var u *NanoVMSOutcomeUnknownError
	if !errors.As(err, &u) {
		t.Fatalf("mutation outcome was not UNKNOWN: %v", err)
	}
}
func TestNanoVMSHTTPTransportHardeningNoRedirectMutationReplay(t *testing.T) {
	var forwarded atomic.Int32
	sink := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		forwarded.Add(1)
		w.WriteHeader(201)
		_, _ = w.Write([]byte(`{"id":"other","status":"running"}`))
	}))
	defer sink.Close()
	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", sink.URL+"/v1/deploy")
		w.WriteHeader(307)
	}))
	defer origin.Close()
	tr := NanoVMSHTTPTransport{BaseURL: origin.URL, Token: "test-token", Client: origin.Client()}
	_, err := tr.Deploy(context.Background(), hardeningRequest())
	if forwarded.Load() != 0 {
		t.Fatalf("redirect replayed mutation %d times", forwarded.Load())
	}
	hardeningUnknown(t, err)
}
func TestNanoVMSHTTPTransportHardeningNoRedirectedObservation(t *testing.T) {
	var forwarded atomic.Int32
	sink := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		forwarded.Add(1)
		_, _ = w.Write([]byte(`{"id":"r","status":"running"}`))
	}))
	defer sink.Close()
	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, sink.URL+"/r", 302) }))
	defer origin.Close()
	tr := NanoVMSHTTPTransport{BaseURL: origin.URL, Token: "test-token", Client: origin.Client()}
	_, found, err := tr.Observe(context.Background(), "r")
	if err == nil || found || forwarded.Load() != 0 {
		t.Fatalf("redirect observation accepted: found=%v calls=%d err=%v", found, forwarded.Load(), err)
	}
}
func TestNanoVMSHTTPTransportHardeningClientTimeoutAndOwnership(t *testing.T) {
	custom := &http.Client{}
	tr := NanoVMSHTTPTransport{Client: custom}
	secured := tr.client()
	if secured.Timeout <= 0 {
		t.Fatal("zero timeout permits indefinite wait")
	}
	if secured == custom || custom.Timeout != 0 || custom.CheckRedirect != nil {
		t.Fatal("shared client mutated")
	}
	if secured.CheckRedirect == nil {
		t.Fatal("redirect barrier missing")
	}
	if tr.client().Timeout <= 0 || (NanoVMSHTTPTransport{}).client().Timeout <= 0 {
		t.Fatal("default deadline missing")
	}
}
func TestNanoVMSHTTPTransportHardeningPreservesTighterTimeout(t *testing.T) {
	custom := &http.Client{Timeout: time.Millisecond}
	tr := NanoVMSHTTPTransport{Client: custom}
	if tr.client().Timeout != time.Millisecond {
		t.Fatal("shorter caller timeout weakened")
	}
}
func TestNanoVMSHTTPTransportHardeningRejectsInvalidEndpointBeforeSending(t *testing.T) {
	for _, base := range []string{"", "relative", "ftp://host", "http://user:password@host", "http://host?x=y", "http://host#fragment"} {
		t.Run(base, func(t *testing.T) {
			calls := 0
			tr := NanoVMSHTTPTransport{BaseURL: base, Client: &http.Client{Transport: hardeningRT(func(*http.Request) (*http.Response, error) {
				calls++
				return hardeningResponse(201, `{"id":"x","status":"running"}`), nil
			})}}
			_, err := tr.Deploy(context.Background(), hardeningRequest())
			var u *NanoVMSOutcomeUnknownError
			if err == nil || errors.As(err, &u) || calls != 0 {
				t.Fatalf("preflight=%v calls=%d", err, calls)
			}
		})
	}
}
func TestNanoVMSHTTPTransportHardeningRejectsCancelledContextBeforeSending(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	calls := 0
	tr := NanoVMSHTTPTransport{BaseURL: "http://fixture.invalid", Client: &http.Client{Transport: hardeningRT(func(*http.Request) (*http.Response, error) { calls++; return nil, context.Canceled })}}
	_, err := tr.Deploy(ctx, hardeningRequest())
	var u *NanoVMSOutcomeUnknownError
	if err == nil || errors.As(err, &u) || calls != 0 {
		t.Fatalf("cancelled request sent or misclassified: calls=%d err=%v", calls, err)
	}
}
func TestNanoVMSHTTPTransportHardeningRejectsUnsafeResourceIDsBeforeSending(t *testing.T) {
	for _, id := range []string{"", " ", ".", "..", "a/b", "a\\b", "a\nb"} {
		t.Run(id, func(t *testing.T) {
			calls := 0
			tr := NanoVMSHTTPTransport{BaseURL: "http://fixture.invalid", Client: &http.Client{Transport: hardeningRT(func(*http.Request) (*http.Response, error) {
				calls++
				return hardeningResponse(200, `{"status":"stopped"}`), nil
			})}}
			if _, _, err := tr.Observe(context.Background(), id); err == nil {
				t.Error("observe accepted unsafe ID")
			}
			err := tr.Stop(context.Background(), id)
			var u *NanoVMSOutcomeUnknownError
			if err == nil || errors.As(err, &u) || calls != 0 {
				t.Fatalf("unsafe ID reached transport: calls=%d err=%v", calls, err)
			}
		})
	}
}
func TestNanoVMSHTTPTransportHardeningMutationHTTPStatusDoesNotProveNoSideEffect(t *testing.T) {
	for _, status := range []int{202, 302, 400, 401, 409, 429, 500, 503, 504} {
		tr := hardeningTransport(status, `{"message":"rejection is not provider-proven non-mutation"}`)
		_, err := tr.Deploy(context.Background(), hardeningRequest())
		hardeningUnknown(t, err)
		hardeningUnknown(t, tr.Stop(context.Background(), "r"))
	}
}
func TestNanoVMSHTTPTransportHardeningNoRawProviderBodyInErrors(t *testing.T) {
	const marker = "sensitive-provider-echo"
	tr := hardeningTransport(500, marker)
	_, err := tr.Deploy(context.Background(), hardeningRequest())
	if err == nil || strings.Contains(err.Error(), marker) {
		t.Fatalf("provider body leaked: %v", err)
	}
	_, _, err = tr.Observe(context.Background(), "r")
	if err == nil || strings.Contains(err.Error(), marker) {
		t.Fatalf("observation body leaked: %v", err)
	}
}
func TestNanoVMSHTTPTransportHardeningBoundedResponse(t *testing.T) {
	tr := hardeningTransport(201, `{"id":"r","status":"running","padding":"`+strings.Repeat("x", (1<<20)+1)+`"}`)
	_, err := tr.Deploy(context.Background(), hardeningRequest())
	hardeningUnknown(t, err)
}
func TestNanoVMSHTTPTransportHardeningNilBodyIsErrorNotPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("nil response body panicked: %v", r)
		}
	}()
	_, err := readNanoVMSBody(&http.Response{StatusCode: 200})
	if err == nil {
		t.Fatal("nil body accepted")
	}
}
func TestNanoVMSHTTPTransportHardeningDoesNotInventObservedConfig(t *testing.T) {
	tr := hardeningTransport(201, `{"id":"r","status":"running"}`)
	s, err := tr.Deploy(context.Background(), hardeningRequest())
	if err != nil {
		t.Fatal(err)
	}
	if s.ConfigDigest != "" {
		t.Fatalf("request config reclassified as provider observation: %q", s.ConfigDigest)
	}
}
func TestNanoVMSHTTPTransportHardeningRequiresStopAcknowledgement(t *testing.T) {
	for _, body := range []string{"", `{bad`, `{}`, `{"status":"running"}`, `{"status":"stopped","id":"other"}`} {
		hardeningUnknown(t, hardeningTransport(200, body).Stop(context.Background(), "r"))
	}
	if err := hardeningTransport(200, `{"status":"stopped"}`).Stop(context.Background(), "r"); err != nil {
		t.Fatal(err)
	}
}
func TestNanoVMSHTTPTransportHardeningUnknownSendFailure(t *testing.T) {
	calls := 0
	tr := NanoVMSHTTPTransport{BaseURL: "http://fixture.invalid", Client: &http.Client{Transport: hardeningRT(func(*http.Request) (*http.Response, error) { calls++; return nil, io.ErrUnexpectedEOF })}}
	_, err := tr.Deploy(context.Background(), hardeningRequest())
	hardeningUnknown(t, err)
	if calls != 1 {
		t.Fatalf("implicit retry count=%d", calls)
	}
}
func TestNanoVMSHTTPTransportHardeningObserveAbsenceAndIdentity(t *testing.T) {
	_, found, err := hardeningTransport(404, `{}`).Observe(context.Background(), "r")
	if found || err != nil {
		t.Fatalf("404: %v %v", found, err)
	}
	_, found, err = hardeningTransport(200, `{"id":"other","status":"running"}`).Observe(context.Background(), "r")
	if found || err == nil {
		t.Fatal("wrong identity accepted")
	}
}
func TestNanoVMSHTTPTransportHardeningSuccessRetainsProviderLabels(t *testing.T) {
	tr := hardeningTransport(201, `{"id":"r","status":"running","labels":{"byteport-config-digest":"cfg"}}`)
	s, err := tr.Deploy(context.Background(), hardeningRequest())
	if err != nil || s.ID != "r" || s.ConfigDigest != "cfg" {
		t.Fatalf("success=%+v err=%v", s, err)
	}
}
