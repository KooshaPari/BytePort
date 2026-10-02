package models

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type NanoVMSHTTPTransport struct {
	BaseURL string
	Token   string
	Client  *http.Client
}

type nanoVMSHTTPDeployBody struct {
	Name          string            `json:"name"`
	Image         string            `json:"image"`
	SandboxType   string            `json:"sandbox_type"`
	Labels        map[string]string `json:"labels,omitempty"`
}

type nanoVMSHTTPSandbox struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
	Labels map[string]string `json:"labels,omitempty"`
}

func (t NanoVMSHTTPTransport) client() *http.Client {
	if t.Client != nil {
		return t.Client
	}
	return http.DefaultClient
}

func (t NanoVMSHTTPTransport) endpoint(path string) (string, error) {
	if strings.TrimSpace(t.BaseURL) == "" {
		return "", fmt.Errorf("NanoVMS HTTP base URL is required")
	}
	return strings.TrimRight(t.BaseURL, "/") + path, nil
}

func (t NanoVMSHTTPTransport) request(
	ctx context.Context,
	method string,
	path string,
	body []byte,
) (*http.Response, error) {
	endpoint, err := t.endpoint(path)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
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
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read NanoVMS response: %w", err)
	}
	return body, nil
}

func (t NanoVMSHTTPTransport) Deploy(
	ctx context.Context,
	req NanoVMSDeployRequest,
) (NanoVMSSandbox, error) {
	if req.OperationID == "" || req.ResourceID == "" || req.Image == "" {
		return NanoVMSSandbox{}, fmt.Errorf("NanoVMS deploy requires operation, resource and image identity")
	}
	body, err := json.Marshal(nanoVMSHTTPDeployBody{
		Name:        req.Name,
		Image:       req.Image,
		SandboxType: "native",
		Labels: map[string]string{
			"byteport-operation-id": req.OperationID,
			"byteport-resource-id":  req.ResourceID,
			"byteport-config-digest": req.ConfigDigest,
		},
	})
	if err != nil {
		return NanoVMSSandbox{}, err
	}

	resp, err := t.request(ctx, http.MethodPost, "/v1/deploy", body)
	if err != nil {
		return NanoVMSSandbox{}, &NanoVMSOutcomeUnknownError{
			Operation: req.OperationID,
			Cause:     err,
		}
	}
	raw, err := readNanoVMSBody(resp)
	if err != nil {
		return NanoVMSSandbox{}, &NanoVMSOutcomeUnknownError{
			Operation: req.OperationID,
			Cause:     err,
		}
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return NanoVMSSandbox{}, fmt.Errorf(
			"NanoVMS deploy rejected status=%d body=%s",
			resp.StatusCode,
			string(raw),
		)
	}
	var sandbox nanoVMSHTTPSandbox
	if err := json.Unmarshal(raw, &sandbox); err != nil {
		return NanoVMSSandbox{}, &NanoVMSOutcomeUnknownError{
			Operation: req.OperationID,
			Cause:     fmt.Errorf("decode successful deploy response: %w", err),
		}
	}
	if sandbox.ID == "" {
		return NanoVMSSandbox{}, &NanoVMSOutcomeUnknownError{
			Operation: req.OperationID,
			Cause:     fmt.Errorf("successful deploy response omitted sandbox identity"),
		}
	}
	return NanoVMSSandbox{
		ID:           sandbox.ID,
		Name:         sandbox.Name,
		Status:       sandbox.Status,
		ConfigDigest: req.ConfigDigest,
	}, nil
}

func (t NanoVMSHTTPTransport) Observe(
	ctx context.Context,
	sandboxID string,
) (NanoVMSSandbox, bool, error) {
	if sandboxID == "" {
		return NanoVMSSandbox{}, false, fmt.Errorf("NanoVMS observe requires sandbox identity")
	}
	resp, err := t.request(
		ctx,
		http.MethodGet,
		"/v1/sandboxes/"+url.PathEscape(sandboxID),
		nil,
	)
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
		return NanoVMSSandbox{}, false, fmt.Errorf(
			"NanoVMS observe rejected status=%d body=%s",
			resp.StatusCode,
			string(raw),
		)
	}
	var sandbox nanoVMSHTTPSandbox
	if err := json.Unmarshal(raw, &sandbox); err != nil {
		return NanoVMSSandbox{}, false, fmt.Errorf("decode NanoVMS observation: %w", err)
	}
	if sandbox.ID != sandboxID {
		return NanoVMSSandbox{}, false, fmt.Errorf(
			"NanoVMS observation identity mismatch: got %q want %q",
			sandbox.ID,
			sandboxID,
		)
	}
	return NanoVMSSandbox{
		ID:           sandbox.ID,
		Name:         sandbox.Name,
		Status:       sandbox.Status,
		ConfigDigest: sandbox.Labels["byteport-config-digest"],
	}, true, nil
}

func (t NanoVMSHTTPTransport) Stop(ctx context.Context, sandboxID string) error {
	if sandboxID == "" {
		return fmt.Errorf("NanoVMS stop requires sandbox identity")
	}
	resp, err := t.request(
		ctx,
		http.MethodPost,
		"/v1/stop?id="+url.QueryEscape(sandboxID),
		nil,
	)
	if err != nil {
		return &NanoVMSOutcomeUnknownError{
			Operation: "stop:" + sandboxID,
			Cause:     err,
		}
	}
	raw, err := readNanoVMSBody(resp)
	if err != nil {
		return &NanoVMSOutcomeUnknownError{
			Operation: "stop:" + sandboxID,
			Cause:     err,
		}
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf(
			"NanoVMS stop rejected status=%d body=%s",
			resp.StatusCode,
			string(raw),
		)
	}
	return nil
}
