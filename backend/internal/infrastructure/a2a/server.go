// Package a2a implements the Agent-to-Agent (A2A) protocol handler for
// BytePort. A2A is a JSON-RPC-over-HTTP protocol that allows autonomous
// agents to delegate tasks to BytePort as a specialized deployment agent.
//
// Spec: https://github.com/a2a-spec/a2a/blob/main/spec.md (v0.3.0)
//
// Endpoints exposed:
//   GET  /.well-known/agent.json    → Agent card (capability advertisement)
//   POST /a2a/{tenant}/tasks/send   → Send a task (request/response)
//   POST /a2a/{tenant}/tasks/stream → Stream task updates via SSE
//   GET  /a2a/{tenant}/tasks/{id}   → Get task status
//   POST /a2a/{tenant}/tasks/{id}/cancel → Cancel a running task
//
// Authentication uses Bearer tokens validated by the global auth middleware.
package a2a

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/byteport/api/internal/infrastructure/observability"
)

// ──────────────────────────────────────────────────────────────────────────────
// Agent Card
// ──────────────────────────────────────────────────────────────────────────────

// Card is the public-facing capability manifest. Served at
// /.well-known/agent.json. See skills.yaml for the canonical list.
type Card struct {
	SchemaVersion string            `json:"schema_version"`
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	Version       string            `json:"version"`
	Description   string            `json:"description"`
	Author        string            `json:"author,omitempty"`
	License       string            `json:"license,omitempty"`
	Endpoints     map[string]string `json:"endpoints"`
	Auth          AuthPolicy        `json:"auth"`
	Security      SecurityPolicy    `json:"security"`
	Capabilities  Capabilities      `json:"capabilities"`
	Metadata      map[string]any    `json:"metadata,omitempty"`
}

// AuthPolicy declares the auth schemes the agent supports.
type AuthPolicy struct {
	Schemes []string `json:"schemes"` // e.g. ["bearer", "oauth2"]
	ServiceDocURL string `json:"service_doc_url,omitempty"`
}

// SecurityPolicy declares post-auth transport security guarantees.
type SecurityPolicy struct {
	TLSRequired bool `json:"tls_required"`
}

// Capabilities enumerates the agent's declared skills/inputs/outputs.
type Capabilities struct {
	Streaming   bool     `json:"streaming"`
	Push        bool     `json:"push_notifications"`
	Tools       []string `json:"tools"`
	InputModes  []string `json:"input_modes"`  // text, image, audio, file
	OutputModes []string `json:"output_modes"` // text, json
}

// ──────────────────────────────────────────────────────────────────────────────
// Task protocol
// ──────────────────────────────────────────────────────────────────────────────

// TaskState is the lifecycle state of an A2A task.
type TaskState string

const (
	StatePending   TaskState = "pending"
	StateRunning   TaskState = "running"
	StateCompleted TaskState = "completed"
	StateFailed    TaskState = "failed"
	StateCancelled TaskState = "cancelled"
)

// Task is a unit of work delegated to BytePort by another agent.
type Task struct {
	ID         string         `json:"id"`
	SessionID  string         `json:"session_id,omitempty"`
	Tool       string         `json:"tool"`     // which skill to invoke
	Input      map[string]any `json:"input"`    // skill-specific arguments
	State      TaskState      `json:"state"`
	Output     map[string]any `json:"output,omitempty"`
	Error      string         `json:"error,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	StreamID   string         `json:"stream_id,omitempty"`
}

// TaskSendRequest is the request body for POST /a2a/.../tasks/send.
type TaskSendRequest struct {
	Tool       string         `json:"tool"`
	Input      map[string]any `json:"input"`
	SessionID  string         `json:"session_id,omitempty"`
	Blocking   bool           `json:"blocking"`
	TimeoutSec int            `json:"timeout_seconds,omitempty"`
}

// TaskCancelRequest is the request body for POST /a2a/.../tasks/{id}/cancel.
type TaskCancelRequest struct {
	Reason string `json:"reason,omitempty"`
}

// TaskHandler is the user-supplied callback that produces a task's output.
// The handler MUST respect ctx cancellation; cancellable behavior is the
// contract A2A clients expect.
type TaskHandler func(ctx context.Context, task *Task) (output map[string]any, err error)

// ──────────────────────────────────────────────────────────────────────────────
// Server
// ──────────────────────────────────────────────────────────────────────────────

// Server hosts the A2A protocol routes on a http.Server (or a Gin server).
type Server struct {
	Card       *Card
	Handlers   map[string]TaskHandler
	tasksMu    sync.RWMutex
	tasks      map[string]*Task
	streamSubs map[string]map[chan *Task]struct{} // task_id → subscribers

	// Metrics
	mTasksTotal *observability.Counter
	mTasksLive  *observability.Gauge
	mTaskDur    *observability.Histogram
}

// NewServer constructs a Server with the canonical agent card.
func NewServer() *Server {
	s := &Server{
		Handlers: make(map[string]TaskHandler),
		tasks:    make(map[string]*Task),
		streamSubs: make(map[string]map[chan *Task]struct{}),
		mTasksTotal: observability.NewCounter(
			"byteport_a2a_tasks_total",
			"Total A2A tasks accepted, by tool and state",
		),
		mTasksLive: observability.NewGauge(
			"byteport_a2a_tasks_active",
			"Currently active A2A tasks",
		),
		mTaskDur: observability.NewHistogram(
			"byteport_a2a_task_duration_seconds",
			"A2A task duration in seconds",
			[]float64{.1, .5, 1, 5, 10, 30, 60, 300, 600},
		),
	}
	s.Card = s.defaultCard()
	return s
}

// Register binds a TaskHandler to a tool/skill name. Called at startup.
func (s *Server) Register(tool string, h TaskHandler) {
	s.Handlers[tool] = h
}

// defaultCard returns the canonical capability manifest. The values here
// MUST match public/.well-known/agent.json for cross-protocol consistency.
func (s *Server) defaultCard() *Card {
	return &Card{
		SchemaVersion: "0.3.0",
		ID:            "byteport",
		Name:          "BytePort",
		Version:       "0.1.0",
		Description:   "Multi-cloud deployment agent: ship apps to AWS, GCP, Azure, Fly, or local Docker via a single declarative interface.",
		Author:        "kooshapari",
		License:       "Apache-2.0",
		Endpoints: map[string]string{
			"agent_card":  "/.well-known/agent.json",
			"task_send":   "/a2a/{tenant}/tasks/send",
			"task_get":    "/a2a/{tenant}/tasks/{id}",
			"task_stream": "/a2a/{tenant}/tasks/stream",
			"task_cancel": "/a2a/{tenant}/tasks/{id}/cancel",
		},
		Auth: AuthPolicy{
			Schemes:       []string{"bearer", "oauth2"},
			ServiceDocURL: "https://docs.byteport.dev/auth",
		},
		Security: SecurityPolicy{TLSRequired: true},
		Capabilities: Capabilities{
			Streaming:   true,
			Push:        true,
			Tools:       []string{"byteport_deploy", "byteport_list_deployments", "byteport_get_deployment", "byteport_terminate_deployment", "byteport_deployment_status", "byteport_deployment_logs", "byteport_estimate_cost", "byteport_detect_app"},
			InputModes:  []string{"text", "json"},
			OutputModes: []string{"text", "json"},
		},
		Metadata: map[string]any{
			"homepage": "https://github.com/kooshapari/BytePort",
			"skills":   "skills.yaml",
			"mcp_card": "public/.well-known/agent.json",
		},
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// HTTP handlers
// ──────────────────────────────────────────────────────────────────────────────

// HandleCard returns the agent card as JSON. Mount on GET /.well-known/agent.json
// of the API server.
func (s *Server) HandleCard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=300")
	if err := json.NewEncoder(w).Encode(s.Card); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// HandleTaskSend accepts a new task and either runs it blocking-style or
// returns immediately with a pending task ID.
func (s *Server) HandleTaskSend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req TaskSendRequest
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid request: %v", err), http.StatusBadRequest)
		return
	}

	handler, ok := s.Handlers[req.Tool]
	if !ok {
		http.Error(w, fmt.Sprintf("unknown tool: %s", req.Tool), http.StatusBadRequest)
		return
	}

	task := &Task{
		ID:        newTaskID(),
		SessionID: req.SessionID,
		Tool:      req.Tool,
		Input:     req.Input,
		State:     StatePending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	s.tasksMu.Lock()
	s.tasks[task.ID] = task
	s.tasksMu.Unlock()

	s.mTasksTotal.Inc("tool", req.Tool, "state", "accepted")
	s.mTasksLive.Inc("tool", req.Tool)
	defer s.mTasksLive.Dec("tool", req.Tool)

	w.Header().Set("Content-Type", "application/json")
	if !req.Blocking {
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(task)
		go s.runTask(task, handler)
		return
	}

	// Blocking mode: run inline with timeout
	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(req.TimeoutSec)*time.Second)
	defer cancel()
	output, err := handler(ctx, task)
	s.finalizeTask(task, output, err)

	s.mTasksTotal.Inc("tool", req.Tool, "state", string(task.State))

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(task)
}

// HandleTaskGet returns the current state of a task.
func (s *Server) HandleTaskGet(w http.ResponseWriter, r *http.Request) {
	id := extractTaskID(r.URL.Path)
	if id == "" {
		http.Error(w, "missing task id", http.StatusBadRequest)
		return
	}
	s.tasksMu.RLock()
	task, ok := s.tasks[id]
	s.tasksMu.RUnlock()
	if !ok {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(task)
}

// HandleTaskStream streams task updates via Server-Sent Events.
// Subscribers receive an event whenever the task's state changes.
func (s *Server) HandleTaskStream(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	id := extractTaskID(r.URL.Path)
	if id == "" {
		http.Error(w, "missing task id", http.StatusBadRequest)
		return
	}

	s.tasksMu.RLock()
	task, ok := s.tasks[id]
	s.tasksMu.RUnlock()
	if !ok {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}

	ch := make(chan *Task, 8)
	s.subscribe(id, ch)
	defer s.unsubscribe(id, ch)

	// Send current state immediately
	writeSSE(w, flusher, "state", task)

	for {
		select {
		case <-r.Context().Done():
			return
		case t := <-ch:
			writeSSE(w, flusher, "state", t)
			if t.State == StateCompleted || t.State == StateFailed || t.State == StateCancelled {
				return
			}
		}
	}
}

// HandleTaskCancel best-effort cancels an in-progress task.
func (s *Server) HandleTaskCancel(w http.ResponseWriter, r *http.Request) {
	id := extractTaskID(r.URL.Path)
	if id == "" {
		http.Error(w, "missing task id", http.StatusBadRequest)
		return
	}
	var req TaskCancelRequest
	_ = json.NewDecoder(io.LimitReader(r.Body, 1<<16)).Decode(&req)

	s.tasksMu.Lock()
	task, ok := s.tasks[id]
	if !ok {
		s.tasksMu.Unlock()
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}
	if task.State == StateRunning {
		task.State = StateCancelled
		task.UpdatedAt = time.Now()
	}
	s.tasksMu.Unlock()

	s.mTasksTotal.Inc("tool="+task.Tool, "state=cancelled")
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(task)
}

// ──────────────────────────────────────────────────────────────────────────────
// Internal helpers
// ──────────────────────────────────────────────────────────────────────────────

// runTask executes a non-blocking task in the background and notifies subscribers.
func (s *Server) runTask(task *Task, handler TaskHandler) {
	s.tasksMu.Lock()
	task.State = StateRunning
	task.UpdatedAt = time.Now()
	s.tasksMu.Unlock()
	s.notify(task)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	start := time.Now()
	output, err := handler(ctx, task)
	s.mTaskDur.Observe(time.Since(start).Seconds(), "tool", task.Tool)

	s.tasksMu.Lock()
	s.finalizeTask(task, output, err)
	s.tasksMu.Unlock()

	s.mTasksTotal.Inc("tool", task.Tool, "state", string(task.State))
	s.notify(task)
}

// finalizeTask sets the final state on a task. caller MUST hold s.tasksMu.
func (s *Server) finalizeTask(task *Task, output map[string]any, err error) {
	task.UpdatedAt = time.Now()
	if err != nil {
		task.Error = err.Error()
		task.State = StateFailed
	} else {
		task.Output = output
		task.State = StateCompleted
	}
}

func (s *Server) subscribe(taskID string, ch chan *Task) {
	s.tasksMu.Lock()
	defer s.tasksMu.Unlock()
	subs, ok := s.streamSubs[taskID]
	if !ok {
		subs = make(map[chan *Task]struct{})
		s.streamSubs[taskID] = subs
	}
	subs[ch] = struct{}{}
}

func (s *Server) unsubscribe(taskID string, ch chan *Task) {
	s.tasksMu.Lock()
	defer s.tasksMu.Unlock()
	subs, ok := s.streamSubs[taskID]
	if !ok {
		return
	}
	delete(subs, ch)
	if len(subs) == 0 {
		delete(s.streamSubs, taskID)
	}
}

func (s *Server) notify(task *Task) {
	s.tasksMu.RLock()
	subs := s.streamSubs[task.ID]
	// Snapshot subscribers under the lock
	targets := make([]chan *Task, 0, len(subs))
	for ch := range subs {
		targets = append(targets, ch)
	}
	s.tasksMu.RUnlock()

	for _, ch := range targets {
		select {
		case ch <- task:
		default:
			// drop if subscriber is slow
		}
	}
}

// extractTaskID parses .../tasks/{id} from the request path.
func extractTaskID(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	for i, p := range parts {
		if p == "tasks" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

func newTaskID() string {
	return fmt.Sprintf("task-%d", time.Now().UnixNano())
}

func writeSSE(w http.ResponseWriter, f http.Flusher, event string, data any) {
	raw, _ := json.Marshal(data)
	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, raw)
	f.Flush()
}
