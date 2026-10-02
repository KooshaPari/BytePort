package models

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNanoVMSHTTPTransportDeployObserveStopExactIdentity(t *testing.T) {
	var deployed bool
	var stopped bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v1/deploy":
			deployed = true
			if got := r.Header.Get("Authorization"); got != "Bearer token" {
				t.Fatalf("authorization=%q", got)
			}
			raw, _ := io.ReadAll(r.Body)
			body := string(raw)
			for _, want := range []string{
				`"byteport-operation-id":"op-1"`,
				`"byteport-resource-id":"service"`,
				`"byteport-config-digest":"cfg-1"`,
				`"image":"sha256:artifact"`,
			} {
				if !strings.Contains(body, want) {
					t.Fatalf("deploy body missing %s: %s", want, body)
				}
			}
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":"sandbox-1","name":"service","status":"running"}`))
		case r.Method == http.MethodGet && r.URL.Path == "/v1/sandboxes/sandbox-1":
			_, _ = w.Write([]byte(
				`{"id":"sandbox-1","name":"service","status":"running","labels":{"byteport-config-digest":"cfg-1"}}`,
			))
		case r.Method == http.MethodPost && r.URL.Path == "/v1/stop":
			if r.URL.Query().Get("id") != "sandbox-1" {
				t.Fatalf("stop id=%q", r.URL.Query().Get("id"))
			}
			stopped = true
			_, _ = w.Write([]byte(`{"status":"stopped"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	transport := NanoVMSHTTPTransport{
		BaseURL: server.URL,
		Token:   "token",
		Client:  server.Client(),
	}
	sandbox, err := transport.Deploy(context.Background(), NanoVMSDeployRequest{
		Name: "service", Image: "sha256:artifact", OperationID: "op-1",
		ResourceID: "service", ConfigDigest: "cfg-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !deployed || sandbox.ID != "sandbox-1" || sandbox.ConfigDigest != "cfg-1" {
		t.Fatalf("deploy=%v sandbox=%+v", deployed, sandbox)
	}

	observed, found, err := transport.Observe(context.Background(), sandbox.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !found || observed.ID != sandbox.ID || observed.ConfigDigest != "cfg-1" {
		t.Fatalf("observed=%+v found=%v", observed, found)
	}

	if err := transport.Stop(context.Background(), sandbox.ID); err != nil {
		t.Fatal(err)
	}
	if !stopped {
		t.Fatal("exact sandbox stop was not issued")
	}
}

func TestNanoVMSHTTPTransportSuccessfulDeployWithoutIdentityIsUnknown(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"status":"running"}`))
	}))
	defer server.Close()

	transport := NanoVMSHTTPTransport{BaseURL: server.URL, Client: server.Client()}
	_, err := transport.Deploy(context.Background(), NanoVMSDeployRequest{
		Name: "service", Image: "sha256:artifact", OperationID: "op-unknown",
		ResourceID: "service", ConfigDigest: "cfg",
	})
	var unknown *NanoVMSOutcomeUnknownError
	if !errors.As(err, &unknown) {
		t.Fatalf("error=%v want NanoVMSOutcomeUnknownError", err)
	}
}

func TestNanoVMSHTTPTransportPostSendDeployFailureIsUnknown(t *testing.T) {
	transport := NanoVMSHTTPTransport{
		BaseURL: "http://nanovms.invalid",
		Client: &http.Client{Transport: recoveryHTTPRoundTripper(func(*http.Request) (*http.Response, error) {
			return nil, io.ErrUnexpectedEOF
		})},
	}
	_, err := transport.Deploy(context.Background(), NanoVMSDeployRequest{
		Name: "service", Image: "sha256:artifact", OperationID: "op-lost",
		ResourceID: "service", ConfigDigest: "cfg",
	})
	var unknown *NanoVMSOutcomeUnknownError
	if !errors.As(err, &unknown) {
		t.Fatalf("error=%v want NanoVMSOutcomeUnknownError", err)
	}
	if unknown.Operation != "op-lost" {
		t.Fatalf("unknown operation=%q", unknown.Operation)
	}
}

func TestNanoVMSHTTPTransportObserveNotFoundIsAbsenceNotMutationAuthority(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	defer server.Close()
	transport := NanoVMSHTTPTransport{BaseURL: server.URL, Client: server.Client()}
	sandbox, found, err := transport.Observe(context.Background(), "sandbox-missing")
	if err != nil {
		t.Fatal(err)
	}
	if found || sandbox.ID != "" {
		t.Fatalf("found=%v sandbox=%+v", found, sandbox)
	}
}

type recoveryHTTPRoundTripper func(*http.Request) (*http.Response, error)

func (f recoveryHTTPRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
