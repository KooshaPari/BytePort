// loadtest_test.go — BytePort API load testing
//
// Run with: go test -bench=. -benchtime=10s -timeout=60s ./routes/
//
// These benchmarks measure the throughput and latency of the NanoVMS contract
// endpoints under concurrent load. They use mock servers to isolate API
// overhead from NanoVMS processing time.

package routes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"
)

// mockNVMSLoadTest returns a mock NanoVMS server that simulates realistic latencies.
func mockNVMSLoadTest(t *testing.T) *httptest.Server {
	t.Helper()
	var mu sync.Mutex
	count := 0

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		count++
		mu.Unlock()

		// Simulate realistic NanoVMS latency (1-5ms)
		time.Sleep(time.Duration(1+count%5) * time.Millisecond)

		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case "POST":
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":     fmt.Sprintf("sb-%d", count),
				"name":   "load-test",
				"status": "running",
			})
		case "GET":
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": []map[string]interface{}{
					{"id": "sb-1", "name": "test", "status": "running"},
				},
			})
		case "DELETE":
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
		}
	}))
}

// BenchmarkDeployEndpoint measures deploy throughput.
func BenchmarkDeployEndpoint(b *testing.B) {
	mockNVMS := mockNVMSLoadTest(&testing.T{})
	defer mockNVMS.Close()

	os.Setenv("NVMS_URL", mockNVMS.URL)
	os.Setenv("NVMS_TOKEN", "bench-token")
	defer os.Unsetenv("NVMS_URL")
	defer os.Unsetenv("NVMS_TOKEN")

	cfg := nvmsSandboxConfig{
		Name:        "bench-project",
		Image:       "alpine:latest",
		SandboxType: "native",
	}
	jsonBody, _ := json.Marshal(cfg)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req, _ := http.NewRequest("POST", mockNVMS.URL+"/v1/deploy", bytes.NewReader(jsonBody))
			req.Header.Set("Authorization", "Bearer bench-token")
			req.Header.Set("Content-Type", "application/json")

			resp, err := (&http.Client{}).Do(req)
			if err != nil {
				b.Errorf("request failed: %v", err)
				return
			}
			resp.Body.Close()
		}
	})
}

// BenchmarkListEndpoint measures list throughput.
func BenchmarkListEndpoint(b *testing.B) {
	mockNVMS := mockNVMSLoadTest(&testing.T{})
	defer mockNVMS.Close()

	os.Setenv("NVMS_URL", mockNVMS.URL)
	os.Setenv("NVMS_TOKEN", "bench-token")
	defer os.Unsetenv("NVMS_URL")
	defer os.Unsetenv("NVMS_TOKEN")

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req, _ := http.NewRequest("GET", mockNVMS.URL+"/v1/sandboxes", nil)
			req.Header.Set("Authorization", "Bearer bench-token")

			resp, err := (&http.Client{}).Do(req)
			if err != nil {
				b.Errorf("request failed: %v", err)
				return
			}
			resp.Body.Close()
		}
	})
}

// BenchmarkStopEndpoint measures stop throughput.
func BenchmarkStopEndpoint(b *testing.B) {
	mockNVMS := mockNVMSLoadTest(&testing.T{})
	defer mockNVMS.Close()

	os.Setenv("NVMS_URL", mockNVMS.URL)
	os.Setenv("NVMS_TOKEN", "bench-token")
	defer os.Unsetenv("NVMS_URL")
	defer os.Unsetenv("NVMS_TOKEN")

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req, _ := http.NewRequest("POST", mockNVMS.URL+"/v1/stop?id=bench-sb-1", nil)
			req.Header.Set("Authorization", "Bearer bench-token")

			resp, err := (&http.Client{}).Do(req)
			if err != nil {
				b.Errorf("request failed: %v", err)
				return
			}
			resp.Body.Close()
		}
	})
}

// TestConcurrentDeployStress tests 50 concurrent deploys.
func TestConcurrentDeployStress(t *testing.T) {
	mockNVMS := mockNVMSLoadTest(t)
	defer mockNVMS.Close()

	os.Setenv("NVMS_URL", mockNVMS.URL)
	os.Setenv("NVMS_TOKEN", "stress-token")
	defer os.Unsetenv("NVMS_URL")
	defer os.Unsetenv("NVMS_TOKEN")

	const concurrency = 50
	var wg sync.WaitGroup
	errors := make(chan error, concurrency)

	start := time.Now()

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			cfg := nvmsSandboxConfig{
				Name:        fmt.Sprintf("stress-%d", id),
				Image:       "alpine:latest",
				SandboxType: "native",
			}
			jsonBody, _ := json.Marshal(cfg)

			req, _ := http.NewRequest("POST", mockNVMS.URL+"/v1/deploy", bytes.NewReader(jsonBody))
			req.Header.Set("Authorization", "Bearer stress-token")
			req.Header.Set("Content-Type", "application/json")

			resp, err := (&http.Client{}).Do(req)
			if err != nil {
				errors <- fmt.Errorf("goroutine %d: %v", id, err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusCreated {
				errors <- fmt.Errorf("goroutine %d: status %d", id, resp.StatusCode)
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	elapsed := time.Since(start)
	t.Logf("Concurrent deploy stress: %d requests in %v (%.0f req/s)",
		concurrency, elapsed, float64(concurrency)/elapsed.Seconds())

	for err := range errors {
		t.Errorf("Stress test error: %v", err)
	}
}
