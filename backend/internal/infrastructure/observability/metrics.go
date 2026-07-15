// Package observability provides runtime telemetry: Prometheus metrics,
// structured logging, and tracing hooks. The metrics here are exported by
// the API on GET /metrics in standard Prometheus exposition format.
//
// Design notes:
//   - We avoid prometheus/client_golang as a hard dependency by emitting the
//     exposition format from an in-process registry. This keeps the binary
//     small and lets us define typed business counters (deployments, costs,
//     MCP tool calls) without paying the dependency tax of the OpenMetrics
//     spec on the wire.
//   - For richer histograms (latency buckets), callers wrap handlers with
//     Observe() which records timing into the latency histogram.
//   - The HTTP middleware from internal/infrastructure/otel records span
//     timings via the OTel pipeline; metrics here are deliberately a
//     lightweight sibling export.
package observability

import (
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

// ──────────────────────────────────────────────────────────────────────────────
// Metric types
// ──────────────────────────────────────────────────────────────────────────────

// Counter is a monotonically increasing uint64 with optional labels.
type Counter struct {
	name   string
	help   string
	values sync.Map // map[string]*uint64 — keys are serialized label sets
}

// Gauge is an int64 that may increase or decrease with optional labels.
type Gauge struct {
	name   string
	help   string
	values sync.Map
}

// Histogram tracks bucketed observations with sum + count.
type Histogram struct {
	name      string
	help      string
	bucketUBs []float64 // upper bounds, sorted ascending; +Inf sentinel appended
	instances sync.Map  // map[string]*histogramInstance
}

type histogramInstance struct {
	mu           sync.Mutex
	bucketCounts []uint64
	sum          float64
	count        uint64
}

// ──────────────────────────────────────────────────────────────────────────────
// Global registry
// ──────────────────────────────────────────────────────────────────────────────

var (
	registryMu sync.RWMutex
	counters   = map[string]*Counter{}
	gauges     = map[string]*Gauge{}
	histograms = map[string]*Histogram{}
)

// ──────────────────────────────────────────────────────────────────────────────
// Registration
// ──────────────────────────────────────────────────────────────────────────────

// NewCounter registers a counter with the given name and HELP line.
func NewCounter(name, help string) *Counter {
	registryMu.Lock()
	defer registryMu.Unlock()
	if existing, ok := counters[name]; ok {
		return existing
	}
	c := &Counter{name: name, help: help}
	counters[name] = c
	return c
}

// NewGauge registers a gauge with the given name and HELP line.
func NewGauge(name, help string) *Gauge {
	registryMu.Lock()
	defer registryMu.Unlock()
	if existing, ok := gauges[name]; ok {
		return existing
	}
	g := &Gauge{name: name, help: help}
	gauges[name] = g
	return g
}

// NewHistogram registers a histogram with the given buckets (upper bounds).
// The +Inf bucket is appended automatically.
func NewHistogram(name, help string, buckets []float64) *Histogram {
	registryMu.Lock()
	defer registryMu.Unlock()
	if existing, ok := histograms[name]; ok {
		return existing
	}
	sorted := make([]float64, len(buckets))
	copy(sorted, buckets)
	sort.Float64s(sorted)
	h := &Histogram{
		name:      name,
		help:      help,
		bucketUBs: sorted,
	}
	histograms[name] = h
	return h
}

// ──────────────────────────────────────────────────────────────────────────────
// Counter operations
// ──────────────────────────────────────────────────────────────────────────────

// Inc atomically increments the counter.
func (c *Counter) Inc(labels ...string) { c.Add(1, labels...) }

// Add atomically adds v to the counter.
func (c *Counter) Add(v uint64, labels ...string) {
	key := labelKey(labels...)
	actual, _ := c.values.LoadOrStore(key, new(uint64))
	atomic.AddUint64(actual.(*uint64), v)
}

// Value returns the current value for the given labels.
func (c *Counter) Value(labels ...string) uint64 {
	key := labelKey(labels...)
	if v, ok := c.values.Load(key); ok {
		return atomic.LoadUint64(v.(*uint64))
	}
	return 0
}

// ──────────────────────────────────────────────────────────────────────────────
// Gauge operations
// ──────────────────────────────────────────────────────────────────────────────

// Set atomically stores v in the gauge.
func (g *Gauge) Set(v int64, labels ...string) {
	key := labelKey(labels...)
	g.values.Store(key, v)
}

// Inc atomically increments the gauge.
func (g *Gauge) Inc(labels ...string) {
	key := labelKey(labels...)
	actual, _ := g.values.LoadOrStore(key, new(int64))
	atomic.AddInt64(actual.(*int64), 1)
}

// Dec atomically decrements the gauge.
func (g *Gauge) Dec(labels ...string) {
	key := labelKey(labels...)
	actual, _ := g.values.LoadOrStore(key, new(int64))
	atomic.AddInt64(actual.(*int64), -1)
}

// Value returns the current value for the given labels.
func (g *Gauge) Value(labels ...string) int64 {
	key := labelKey(labels...)
	if v, ok := g.values.Load(key); ok {
		return atomic.LoadInt64(v.(*int64))
	}
	return 0
}

// ──────────────────────────────────────────────────────────────────────────────
// Histogram operations
// ──────────────────────────────────────────────────────────────────────────────

// Observe records a single observation v (seconds typically).
func (h *Histogram) Observe(v float64, labels ...string) {
	key := labelKey(labels...)
	actual, _ := h.instances.LoadOrStore(key, &histogramInstance{
		bucketCounts: make([]uint64, len(h.bucketUBs)+1), // +1 for +Inf bucket
	})
	inst := actual.(*histogramInstance)
	inst.mu.Lock()
	inst.count++
	inst.sum += v
	for i, ub := range h.bucketUBs {
		if v <= ub {
			inst.bucketCounts[i]++
		}
	}
	inst.bucketCounts[len(h.bucketUBs)]++ // +Inf
	inst.mu.Unlock()
}

// Snapshot returns bucket counts, sum, count for a given label set.
func (h *Histogram) Snapshot(labels ...string) (buckets []uint64, sum float64, count uint64) {
	key := labelKey(labels...)
	if v, ok := h.instances.Load(key); ok {
		inst := v.(*histogramInstance)
		inst.mu.Lock()
		defer inst.mu.Unlock()
		buckets = make([]uint64, len(inst.bucketCounts))
		copy(buckets, inst.bucketCounts)
		sum = inst.sum
		count = inst.count
	}
	return
}

// ──────────────────────────────────────────────────────────────────────────────
// HTTP handler
// ──────────────────────────────────────────────────────────────────────────────

// MetricsHandler returns an http.Handler serving the registry in Prometheus
// text exposition format. Mount at GET /metrics on the API server.
func MetricsHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		writeMetrics(w)
	})
}

func writeMetrics(w io.Writer) {
	registryMu.RLock()
	defer registryMu.RUnlock()

	// Walk in stable name order for deterministic output
	names := make([]string, 0, len(counters)+len(gauges)+len(histograms))
	for n := range counters {
		names = append(names, n)
	}
	for n := range gauges {
		names = append(names, n)
	}
	for n := range histograms {
		names = append(names, n)
	}
	sort.Strings(names)
	written := make(map[string]bool)

	for _, n := range names {
		if written[n] {
			continue
		}
		switch {
		case counters[n] != nil:
			c := counters[n]
			fmt.Fprintf(w, "# HELP %s %s\n", c.name, c.help)
			fmt.Fprintf(w, "# TYPE %s counter\n", c.name)
			c.values.Range(func(k, v any) bool {
				key := k.(string)
				val := atomic.LoadUint64(v.(*uint64))
				if lbl := renderLabels(key); lbl != "" {
					fmt.Fprintf(w, "%s{%s} %d\n", c.name, lbl, val)
				} else {
					fmt.Fprintf(w, "%s %d\n", c.name, val)
				}
				return true
			})
		case gauges[n] != nil:
			g := gauges[n]
			fmt.Fprintf(w, "# HELP %s %s\n", g.name, g.help)
			fmt.Fprintf(w, "# TYPE %s gauge\n", g.name)
			g.values.Range(func(k, v any) bool {
				key := k.(string)
				val := atomic.LoadInt64(v.(*int64))
				if lbl := renderLabels(key); lbl != "" {
					fmt.Fprintf(w, "%s{%s} %d\n", g.name, lbl, val)
				} else {
					fmt.Fprintf(w, "%s %d\n", g.name, val)
				}
				return true
			})
		case histograms[n] != nil:
			h := histograms[n]
			fmt.Fprintf(w, "# HELP %s %s\n", h.name, h.help)
			fmt.Fprintf(w, "# TYPE %s histogram\n", h.name)
			h.instances.Range(func(k, v any) bool {
				key := k.(string)
				inst := v.(*histogramInstance)
				inst.mu.Lock()
				buckets := make([]uint64, len(inst.bucketCounts))
				copy(buckets, inst.bucketCounts)
				sum := inst.sum
				count := inst.count
				inst.mu.Unlock()
				lbl := renderLabels(key)
				lblWithComma := lbl
				if lblWithComma != "" {
					lblWithComma += ","
				}
				for i, ub := range h.bucketUBs {
					le := strconv.FormatFloat(ub, 'g', -1, 64)
					fmt.Fprintf(w, "%s_bucket{%sle=\"%s\"} %d\n", h.name, lblWithComma, le, buckets[i])
				}
				fmt.Fprintf(w, "%s_bucket{%sle=\"+Inf\"} %d\n", h.name, lblWithComma, buckets[len(buckets)-1])
				fmt.Fprintf(w, "%s_sum{%s} %s\n", h.name, lbl, strconv.FormatFloat(sum, 'g', -1, 64))
				fmt.Fprintf(w, "%s_count{%s} %d\n", h.name, lbl, count)
				return true
			})
		}
		written[n] = true
	}
}

// Observe is HTTP middleware that records timing into a histogram.
// Use as:  mux.Handle("/", observability.Observe("byteport_http_request_duration_seconds")(handler))
func Observe(histName string) func(next http.Handler) http.Handler {
	registryMu.Lock()
	h, ok := histograms[histName]
	if !ok {
		h = NewHistogram(histName, "HTTP request duration in seconds",
			[]float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10})
	}
	registryMu.Unlock()
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := &statusRecorder{ResponseWriter: w, status: 200}
			next.ServeHTTP(rw, r)
			h.Observe(time.Since(start).Seconds(), "method="+r.Method, "route="+r.URL.Path, fmt.Sprintf("status=%d", rw.status))
		})
	}
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// ──────────────────────────────────────────────────────────────────────────────
// internal helpers
// ──────────────────────────────────────────────────────────────────────────────

// labelKey serializes k/v pairs into a stable map key. The empty string
// produces "labels=nil" so labeled and unlabeled metrics can co-exist.
func labelKey(labels ...string) string {
	if len(labels) == 0 {
		return "labels=nil"
	}
	return "labels=" + joinLabelKVs(labels...)
}

func joinLabelKVs(labels ...string) string {
	if len(labels)%2 != 0 {
		panic("observability: label key/value pairs must be even")
	}
	out := ""
	for i := 0; i < len(labels); i += 2 {
		if i > 0 {
			out += ","
		}
		out += labels[i] + "=\"" + labels[i+1] + "\""
	}
	return out
}

// renderLabels converts a stored map key back into Prometheus label form.
// Stored keys are "labels=nil" or "labels=key=\"val\",key=\"val\"". We strip
// the leading "labels=" to produce the final label string (or empty).
func renderLabels(key string) string {
	if key == "labels=nil" {
		return ""
	}
	return key[len("labels="):]
}
