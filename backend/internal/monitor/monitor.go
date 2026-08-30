package monitor

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ContainerStatus represents the status of a sandboxed container from the nanovms API.
type ContainerStatus struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Status   string  `json:"status"`
	CPU      float64 `json:"cpu"`
	MemoryMB float64 `json:"memory_mb"`
	Uptime   int64   `json:"uptime"`
	TierID   int     `json:"tier_id"`
}

// nvmsContainerResponse is the raw JSON shape returned by the nanovms status endpoint.
type nvmsContainerResponse struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Status   string  `json:"status"`
	CPU      float64 `json:"cpu"`
	MemoryMB float64 `json:"memory_mb"`
	Uptime   int64   `json:"uptime"`
	TierID   int     `json:"tier_id"`
}

// nvmsListResponse wraps a list of containers from the nanovms API.
type nvmsListResponse struct {
	Containers []nvmsContainerResponse `json:"containers"`
}

// Monitor provides container monitoring by proxying requests to the nanovms API.
type Monitor struct {
	nvmsURL string
	client  *http.Client
}

// NewMonitor creates a Monitor that talks to the nanovms API at nvmsURL.
func NewMonitor(nvmsURL string) *Monitor {
	return &Monitor{
		nvmsURL: nvmsURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetStatus fetches the status of a single sandbox from the nanovms API.
func (m *Monitor) GetStatus(ctx context.Context, sandboxID string) (*ContainerStatus, error) {
	url := fmt.Sprintf("%s/sandboxes/%s/status", m.nvmsURL, sandboxID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := m.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching sandbox status: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("nanovms API returned status %d: %s", resp.StatusCode, string(body))
	}

	var raw nvmsContainerResponse
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &ContainerStatus{
		ID:       raw.ID,
		Name:     raw.Name,
		Status:   raw.Status,
		CPU:      raw.CPU,
		MemoryMB: raw.MemoryMB,
		Uptime:   raw.Uptime,
		TierID:   raw.TierID,
	}, nil
}

// ListAll fetches the status of all containers from the nanovms API.
func (m *Monitor) ListAll(ctx context.Context) ([]ContainerStatus, error) {
	url := fmt.Sprintf("%s/sandboxes", m.nvmsURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := m.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching sandbox list: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("nanovms API returned status %d: %s", resp.StatusCode, string(body))
	}

	var raw nvmsListResponse
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	result := make([]ContainerStatus, len(raw.Containers))
	for i, c := range raw.Containers {
		result[i] = ContainerStatus{
			ID:       c.ID,
			Name:     c.Name,
			Status:   c.Status,
			CPU:      c.CPU,
			MemoryMB: c.MemoryMB,
			Uptime:   c.Uptime,
			TierID:   c.TierID,
		}
	}

	return result, nil
}
