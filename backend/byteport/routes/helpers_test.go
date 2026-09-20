package routes

import (
	"bytes"
	"net/http"
	"os"
	"testing"
)

// setupBenchEnv starts the mock NanoVMS server and sets the NVMS_URL / NVMS_TOKEN
// env vars to point at it. It registers a cleanup via t.Cleanup that closes the
// server and unsets the env vars. Returns the server URL.
func setupBenchEnv(tb testing.TB, token string) string {
	tb.Helper()
	mockNVMS := mockNVMSLoadTest(tb)
	tb.Cleanup(mockNVMS.Close)

	os.Setenv("NVMS_URL", mockNVMS.URL)
	os.Setenv("NVMS_TOKEN", token)
	tb.Cleanup(func() {
		os.Unsetenv("NVMS_URL")
		os.Unsetenv("NVMS_TOKEN")
	})

	return mockNVMS.URL
}

// runBenchRequest performs a single request against the mock NVMS server.
// `extra` is an optional finalizer for headers/body that vary per benchmark.
func runBenchRequest(b *testing.B, method, url, token string, body []byte, extra func(*http.Request)) {
	b.Helper()
	req, _ := http.NewRequest(method, url, bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	if extra != nil {
		extra(req)
	}
	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		b.Errorf("request failed: %v", err)
		return
	}
	resp.Body.Close()
}
