package monitor

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewMonitor(t *testing.T) {
	m := NewMonitor("http://localhost:9999")
	if m == nil {
		t.Fatal("NewMonitor returned nil")
	}
	if m.nvmsURL != "http://localhost:9999" {
		t.Errorf("nvmsURL = %q, want %q", m.nvmsURL, "http://localhost:9999")
	}
	if m.client == nil {
		t.Fatal("HTTP client is nil")
	}
}

func TestGetStatusMockServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/sandboxes/sb-123/status" {
			t.Errorf("path = %s, want /sandboxes/sb-123/status", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(nvmsContainerResponse{
			ID:       "sb-123",
			Name:     "test-sandbox",
			Status:   "running",
			CPU:      45.2,
			MemoryMB: 256.0,
			Uptime:   3600,
			TierID:   2,
		})
	}))
	defer server.Close()

	m := NewMonitor(server.URL)
	status, err := m.GetStatus(context.Background(), "sb-123")
	if err != nil {
		t.Fatalf("GetStatus returned error: %v", err)
	}

	if status.ID != "sb-123" {
		t.Errorf("ID = %q, want %q", status.ID, "sb-123")
	}
	if status.Name != "test-sandbox" {
		t.Errorf("Name = %q, want %q", status.Name, "test-sandbox")
	}
	if status.Status != "running" {
		t.Errorf("Status = %q, want %q", status.Status, "running")
	}
	if status.CPU != 45.2 {
		t.Errorf("CPU = %f, want %f", status.CPU, 45.2)
	}
	if status.MemoryMB != 256.0 {
		t.Errorf("MemoryMB = %f, want %f", status.MemoryMB, 256.0)
	}
	if status.Uptime != 3600 {
		t.Errorf("Uptime = %d, want %d", status.Uptime, 3600)
	}
	if status.TierID != 2 {
		t.Errorf("TierID = %d, want %d", status.TierID, 2)
	}
}

func TestGetStatusInvalidURL(t *testing.T) {
	m := NewMonitor("http://127.0.0.1:1") // port not listening
	_, err := m.GetStatus(context.Background(), "sb-404")
	if err == nil {
		t.Fatal("expected error for invalid URL, got nil")
	}
}

func TestGetStatusNon200(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"not found"}`))
	}))
	defer server.Close()

	m := NewMonitor(server.URL)
	_, err := m.GetStatus(context.Background(), "sb-missing")
	if err == nil {
		t.Fatal("expected error for 404 response, got nil")
	}
}

func TestListAllMockServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/sandboxes" {
			t.Errorf("path = %s, want /sandboxes", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(nvmsListResponse{
			Containers: []nvmsContainerResponse{
				{ID: "sb-1", Name: "alpha", Status: "running", CPU: 10.0, MemoryMB: 128.0, Uptime: 600, TierID: 1},
				{ID: "sb-2", Name: "bravo", Status: "stopped", CPU: 0.0, MemoryMB: 0.0, Uptime: 0, TierID: 3},
			},
		})
	}))
	defer server.Close()

	m := NewMonitor(server.URL)
	list, err := m.ListAll(context.Background())
	if err != nil {
		t.Fatalf("ListAll returned error: %v", err)
	}

	if len(list) != 2 {
		t.Fatalf("len(list) = %d, want 2", len(list))
	}
	if list[0].ID != "sb-1" || list[0].Name != "alpha" || list[0].Status != "running" {
		t.Errorf("first container = %+v, want sb-1/alpha/running", list[0])
	}
	if list[1].ID != "sb-2" || list[1].Name != "bravo" || list[1].Status != "stopped" {
		t.Errorf("second container = %+v, want sb-2/bravo/stopped", list[1])
	}
}

func TestListAllEmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(nvmsListResponse{
			Containers: []nvmsContainerResponse{},
		})
	}))
	defer server.Close()

	m := NewMonitor(server.URL)
	list, err := m.ListAll(context.Background())
	if err != nil {
		t.Fatalf("ListAll returned error: %v", err)
	}

	if len(list) != 0 {
		t.Errorf("len(list) = %d, want 0", len(list))
	}
}

func TestListAllInvalidURL(t *testing.T) {
	m := NewMonitor("http://127.0.0.1:1")
	_, err := m.ListAll(context.Background())
	if err == nil {
		t.Fatal("expected error for invalid URL, got nil")
	}
}
