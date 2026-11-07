package models

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Transport safety ceilings, not provider performance acceptance targets.
const nanoVMSResponseLimit = 1 << 20
const nanoVMSRequestTimeout = 30 * time.Second

type NanoVMSHTTPTransport struct {
	BaseURL string
	Token   string
	Client  *http.Client
}

type nanoVMSHTTPDeployBody struct {
	Name        string            `json:"name"`
	Image       string            `json:"image"`
	SandboxType string            `json:"sandbox_type"`
	Labels      map[string]string `json:"labels,omitempty"`
}

type nanoVMSHTTPSandbox struct {
	ID     string            `json:"id"`
	Name   string            `json:"name"`
	Status string            `json:"status"`
	Labels map[string]string `json:"labels,omitempty"`
}

// The injected client's transport is reused, but the caller-owned client is
// never modified. Redirects cannot change provider identity or replay a POST.
func (t NanoVMSHTTPTransport) client() *http.Client {
	client := &http.Client{}
	if t.Client != nil {
		*client = *t.Client
	}
	if client.Timeout <= 0 || client.Timeout > nanoVMSRequestTimeout {
		client.Timeout = nanoVMSRequestTimeout
	}
	client.CheckRedirect = func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}
	return client
}

func (t NanoVMSHTTPTransport) endpoint(path string) (string, error) {
	base, err := url.Parse(t.BaseURL)
	if err != nil || base == nil || (base.Scheme != "http" && base.Scheme != "https") ||
		base.Hostname() == "" || base.User != nil || base.RawQuery != "" ||
		base.ForceQuery || base.Fragment != "" || base.Opaque != "" {
		return "", errors.New("NanoVMS requires an absolute HTTP(S) base URL without credentials, query or fragment")
	}
	return strings.TrimRight(base.String(), "/") + path, nil
}

// A local preflight failure is known not to have sent a request. All failures
// after Client.Do begins remain ambiguous at the mutation boundary.
type nanoVMSRequestNotSent struct{ cause error }

func (e *nanoVMSRequestNotSent) Error() string { return e.cause.Error() }
func (e *nanoVMSRequestNotSent) Unwrap() error { return e.cause }

func (t NanoVMSHTTPTransport) request(ctx context.Context, method, path string, body []byte) (*http.Response, error) {
	if ctx == nil {
		return nil, &nanoVMSRequestNotSent{errors.New("NanoVMS request context is required")}
	}
	if err := ctx.Err(); err != nil {
		return nil, &nanoVMSRequestNotSent{err}
	}
	endpoint, err := t.endpoint(path)
	if err != nil {
		return nil, &nanoVMSRequestNotSent{err}
	}
	if strings.ContainsAny(t.Token, "\r\n") {
		return nil, &nanoVMSRequestNotSent{errors.New("invalid NanoVMS authorization header")}
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, &nanoVMSRequestNotSent{errors.New("invalid NanoVMS request")}
	}
	if t.Token != "" {
		req.Header.Set("Authorization", "Bearer "+t.Token)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return t.client().Do(req)
}

func readNanoVMSBody(resp *http.Response) ([]byte, error) {
	if resp == nil || resp.Body == nil {
		return nil, errors.New("NanoVMS response body missing")
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, nanoVMSResponseLimit+1))
	if err != nil {
		return nil, fmt.Errorf("read NanoVMS response: %w", err)
	}
	if len(body) > nanoVMSResponseLimit {
		return nil, errors.New("NanoVMS response exceeds safety limit")
	}
	return body, nil
}

func nanoVMSMutationError(operation string, err error) error {
	var notSent *nanoVMSRequestNotSent
	if errors.As(err, &notSent) {
		return err
	}
	return &NanoVMSOutcomeUnknownError{Operation: operation, Cause: err}
}

func validNanoVMSIdentity(id string) bool {
	return id != "" && strings.TrimSpace(id) == id && id != "." && id != ".." &&
		!strings.ContainsAny(id, "/\\\x00\r\n\t?#")
}

func (t NanoVMSHTTPTransport) Deploy(ctx context.Context, req NanoVMSDeployRequest) (NanoVMSSandbox, error) {
	if strings.TrimSpace(req.OperationID) == "" || strings.TrimSpace(req.ResourceID) == "" || strings.TrimSpace(req.Image) == "" {
		return NanoVMSSandbox{}, errors.New("NanoVMS deploy requires operation, resource and image identity")
	}
	body, err := json.Marshal(nanoVMSHTTPDeployBody{
		Name: req.Name, Image: req.Image, SandboxType: "native",
		Labels: map[string]string{
			"byteport-operation-id":  req.OperationID,
			"byteport-resource-id":   req.ResourceID,
			"byteport-config-digest": req.ConfigDigest,
		},
	})
	if err != nil {
		return NanoVMSSandbox{}, err
	}
	resp, err := t.request(ctx, http.MethodPost, "/v1/deploy", body)
	if err != nil {
		return NanoVMSSandbox{}, nanoVMSMutationError(req.OperationID, err)
	}
	raw, err := readNanoVMSBody(resp)
	if err != nil {
		return NanoVMSSandbox{}, nanoVMSMutationError(req.OperationID, err)
	}
	// Until a provider contract proves a status implies no mutation, no HTTP
	// status authorizes blind retry. Provider bodies are not copied into errors.
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return NanoVMSSandbox{}, nanoVMSMutationError(req.OperationID,
			fmt.Errorf("NanoVMS deploy outcome unresolved: HTTP %d", resp.StatusCode))
	}
	var sandbox nanoVMSHTTPSandbox
	if err := json.Unmarshal(raw, &sandbox); err != nil || !validNanoVMSIdentity(sandbox.ID) || strings.TrimSpace(sandbox.Status) == "" {
		return NanoVMSSandbox{}, nanoVMSMutationError(req.OperationID,
			errors.New("successful NanoVMS deploy omitted a valid identity/state response"))
	}
	config := sandbox.Labels["byteport-config-digest"]
	if config != "" && config != req.ConfigDigest {
		return NanoVMSSandbox{}, nanoVMSMutationError(req.OperationID,
			errors.New("NanoVMS returned contradictory configuration identity"))
	}
	return NanoVMSSandbox{ID: sandbox.ID, Name: sandbox.Name, Status: sandbox.Status, ConfigDigest: config}, nil
}

func (t NanoVMSHTTPTransport) Observe(ctx context.Context, sandboxID string) (NanoVMSSandbox, bool, error) {
	if !validNanoVMSIdentity(sandboxID) {
		return NanoVMSSandbox{}, false, errors.New("NanoVMS observe requires safe exact sandbox identity")
	}
	resp, err := t.request(ctx, http.MethodGet, "/v1/sandboxes/"+url.PathEscape(sandboxID), nil)
	if err != nil {
		return NanoVMSSandbox{}, false, err
	}
	raw, err := readNanoVMSBody(resp)
	if err != nil {
		return NanoVMSSandbox{}, false, err
	}
	if resp.StatusCode == http.StatusNotFound {
		return NanoVMSSandbox{}, false, nil
	}
	if resp.StatusCode != http.StatusOK {
		return NanoVMSSandbox{}, false, fmt.Errorf("NanoVMS observation unavailable: HTTP %d", resp.StatusCode)
	}
	var sandbox nanoVMSHTTPSandbox
	if err := json.Unmarshal(raw, &sandbox); err != nil || sandbox.ID != sandboxID || strings.TrimSpace(sandbox.Status) == "" {
		return NanoVMSSandbox{}, false, errors.New("NanoVMS observation identity/state invalid")
	}
	return NanoVMSSandbox{ID: sandbox.ID, Name: sandbox.Name, Status: sandbox.Status,
		ConfigDigest: sandbox.Labels["byteport-config-digest"]}, true, nil
}

func (t NanoVMSHTTPTransport) Stop(ctx context.Context, sandboxID string) error {
	if !validNanoVMSIdentity(sandboxID) {
		return errors.New("NanoVMS stop requires safe exact sandbox identity")
	}
	operation := "stop:" + sandboxID
	resp, err := t.request(ctx, http.MethodPost, "/v1/stop?id="+url.QueryEscape(sandboxID), nil)
	if err != nil {
		return nanoVMSMutationError(operation, err)
	}
	raw, err := readNanoVMSBody(resp)
	if err != nil {
		return nanoVMSMutationError(operation, err)
	}
	if resp.StatusCode != http.StatusOK {
		return nanoVMSMutationError(operation, fmt.Errorf("NanoVMS stop outcome unresolved: HTTP %d", resp.StatusCode))
	}
	var ack nanoVMSHTTPSandbox
	if err := json.Unmarshal(raw, &ack); err != nil || ack.Status != "stopped" || (ack.ID != "" && ack.ID != sandboxID) {
		return nanoVMSMutationError(operation, errors.New("NanoVMS stop acknowledgement missing or contradictory"))
	}
	return nil
}
