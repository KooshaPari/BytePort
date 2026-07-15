// Package a2a — tests for the A2A protocol server.
package a2a

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func newTestServer(t *testing.T) *Server {
	t.Helper()
	s := NewServer()
	s.Register("echo", func(ctx context.Context, task *Task) (map[string]any, error) {
		return map[string]any{"echo": task.Input["msg"]}, nil
	})
	s.Register("failer", func(ctx context.Context, task *Task) (map[string]any, error) {
		return nil, errors.New("synthetic failure")
	})
	return s
}

func TestServer_HandleCard(t *testing.T) {
	s := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/.well-known/agent.json", nil)
	rr := httptest.NewRecorder()

	s.HandleCard(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var card Card
	if err := json.Unmarshal(rr.Body.Bytes(), &card); err != nil {
		t.Fatalf("card decode failed: %v", err)
	}
	if card.Name != "BytePort" {
		t.Errorf("expected name=BytePort, got %q", card.Name)
	}
	if len(card.Capabilities.Tools) == 0 {
		t.Errorf("expected non-empty tools, got 0")
	}
	if !card.Security.TLSRequired {
		t.Errorf("expected TLSRequired true")
	}
}

func TestServer_HandleTaskSend_Blocking(t *testing.T) {
	s := newTestServer(t)
	body := `{"tool":"echo","input":{"msg":"hello"},"blocking":true}`
	req := httptest.NewRequest(http.MethodPost, "/a2a/default/tasks/send", strings.NewReader(body))
	rr := httptest.NewRecorder()

	s.HandleTaskSend(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var task Task
	if err := json.Unmarshal(rr.Body.Bytes(), &task); err != nil {
		t.Fatalf("task decode failed: %v", err)
	}
	if task.State != StateCompleted {
		t.Errorf("expected state=completed, got %s", task.State)
	}
	if task.Output["echo"] != "hello" {
		t.Errorf("expected echo=hello, got %v", task.Output["echo"])
	}
}

func TestServer_HandleTaskSend_NonBlocking(t *testing.T) {
	s := newTestServer(t)
	body := `{"tool":"echo","input":{"msg":"bye"},"blocking":false}`
	req := httptest.NewRequest(http.MethodPost, "/a2a/default/tasks/send", strings.NewReader(body))
	rr := httptest.NewRecorder()

	s.HandleTaskSend(rr, req)

	if rr.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", rr.Code)
	}

	// Give the goroutine a moment to complete
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		s.tasksMu.RLock()
		for _, task := range s.tasks {
			s.tasksMu.RUnlock()
			if task.State == StateCompleted {
				if task.Output["echo"] != "bye" {
					t.Errorf("expected echo=bye, got %v", task.Output["echo"])
				}
				return
			}
			s.tasksMu.RLock()
			continue
		}
		s.tasksMu.RUnlock()
		time.Sleep(50 * time.Millisecond)
	}
	t.Errorf("task did not complete within 2s")
}

func TestServer_HandleTaskSend_UnknownTool(t *testing.T) {
	s := newTestServer(t)
	body := `{"tool":"nonexistent","input":{}}`
	req := httptest.NewRequest(http.MethodPost, "/a2a/default/tasks/send", strings.NewReader(body))
	rr := httptest.NewRecorder()

	s.HandleTaskSend(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestServer_HandleTaskSend_Failer(t *testing.T) {
	s := newTestServer(t)
	body := `{"tool":"failer","input":{},"blocking":true}`
	req := httptest.NewRequest(http.MethodPost, "/a2a/default/tasks/send", strings.NewReader(body))
	rr := httptest.NewRecorder()

	s.HandleTaskSend(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var task Task
	if err := json.Unmarshal(rr.Body.Bytes(), &task); err != nil {
		t.Fatalf("task decode failed: %v", err)
	}
	if task.State != StateFailed {
		t.Errorf("expected state=failed, got %s", task.State)
	}
	if task.Error == "" {
		t.Errorf("expected non-empty error message")
	}
}

func TestServer_HandleTaskGet(t *testing.T) {
	s := newTestServer(t)
	// Seed a task
	s.tasksMu.Lock()
	id := newTaskID()
	s.tasks[id] = &Task{ID: id, Tool: "echo", State: StatePending, CreatedAt: time.Now()}
	s.tasksMu.Unlock()

	req := httptest.NewRequest(http.MethodGet, "/a2a/default/tasks/"+id, nil)
	rr := httptest.NewRecorder()
	s.HandleTaskGet(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var task Task
	if err := json.Unmarshal(rr.Body.Bytes(), &task); err != nil {
		t.Fatalf("task decode failed: %v", err)
	}
	if task.ID != id {
		t.Errorf("expected id=%s, got %s", id, task.ID)
	}
}

func TestServer_HandleTaskGet_NotFound(t *testing.T) {
	s := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/a2a/default/tasks/nonexistent-id", nil)
	rr := httptest.NewRecorder()
	s.HandleTaskGet(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestServer_HandleTaskCancel(t *testing.T) {
	s := newTestServer(t)
	// Seed a running task
	s.tasksMu.Lock()
	id := newTaskID()
	s.tasks[id] = &Task{ID: id, Tool: "echo", State: StateRunning, CreatedAt: time.Now()}
	s.tasksMu.Unlock()

	req := httptest.NewRequest(http.MethodPost, "/a2a/default/tasks/"+id+"/cancel", strings.NewReader(`{"reason":"user"}`))
	rr := httptest.NewRecorder()
	s.HandleTaskCancel(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	s.tasksMu.RLock()
	if s.tasks[id].State != StateCancelled {
		t.Errorf("expected state=cancelled, got %s", s.tasks[id].State)
	}
	s.tasksMu.RUnlock()
}

func TestExtractTaskID(t *testing.T) {
	tests := map[string]struct {
		path string
		want string
	}{
		"simple":        {path: "/a2a/default/tasks/abc123", want: "abc123"},
		"nested":        {path: "/a2a/tenant-1/tasks/task-456", want: "task-456"},
		"missing":       {path: "/a2a/default/somewhere", want: ""},
		"trailing":      {path: "/tasks/xyz", want: "xyz"},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := extractTaskID(tc.path)
			if got != tc.want {
				t.Errorf("got %q want %q", got, tc.want)
			}
		})
	}
}
