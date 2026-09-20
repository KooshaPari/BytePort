package deployment

import (
	"strings"
	"testing"
	"time"
)

func TestNewDeployment(t *testing.T) {
	tests := []struct {
		name        string
		depName     string
		owner       string
		projectUUID *string
		wantErr     bool
	}{
		{
			name:        "valid deployment",
			depName:     "test-deployment",
			owner:       "user-123",
			projectUUID: nil,
			wantErr:     false,
		},
		{
			name:        "empty name",
			depName:     "",
			owner:       "user-123",
			projectUUID: nil,
			wantErr:     true,
		},
		{
			name:        "empty owner",
			depName:     "test-deployment",
			owner:       "",
			projectUUID: nil,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dep, err := NewDeployment(tt.depName, tt.owner, tt.projectUUID)
			assertErr(t, err, tt.wantErr)
			if err != nil {
				return
			}

			if dep.Name() != tt.depName {
				t.Errorf("Name() = %v, want %v", dep.Name(), tt.depName)
			}
			if dep.Owner() != tt.owner {
				t.Errorf("Owner() = %v, want %v", dep.Owner(), tt.owner)
			}
			if dep.Status() != StatusPending {
				t.Errorf("Status() = %v, want %v", dep.Status(), StatusPending)
			}
			if dep.UUID() == "" {
				t.Error("UUID() should not be empty")
			}
		})
	}
}

func TestDeployment_StatusTransitions(t *testing.T) {
	tests := []struct {
		name       string
		fromStatus Status
		toStatus   Status
		wantErr    bool
	}{
		{"pending to detecting", StatusPending, StatusDetecting, false},
		{"detecting to provisioning", StatusDetecting, StatusProvisioning, false},
		{"provisioning to deploying", StatusProvisioning, StatusDeploying, false},
		{"deploying to deployed", StatusDeploying, StatusDeployed, false},
		{"deployed to terminated", StatusDeployed, StatusTerminated, false},
		{"pending to deployed (invalid)", StatusPending, StatusDeployed, true},
		{"terminated to deploying (invalid)", StatusTerminated, StatusDeploying, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dep := newTestDeployment(t)
			transitionTo(dep, tt.fromStatus)

			err := dep.SetStatus(tt.toStatus)
			assertErr(t, err, tt.wantErr)
			if !tt.wantErr && dep.Status() != tt.toStatus {
				t.Errorf("Status() = %v, want %v", dep.Status(), tt.toStatus)
			}
		})
	}
}

func TestDeployment_SetStatus_Timestamps(t *testing.T) {
	dep := newTestDeployment(t)

	// Test deployed timestamp
	transitionTo(dep, StatusDeploying)

	beforeDeploy := time.Now().UTC()
	_ = dep.SetStatus(StatusDeployed)
	afterDeploy := time.Now().UTC()

	if dep.DeployedAt() == nil {
		t.Error("DeployedAt() should not be nil after StatusDeployed")
	} else if dep.DeployedAt().Before(beforeDeploy) || dep.DeployedAt().After(afterDeploy) {
		t.Error("DeployedAt() timestamp is out of expected range")
	}

	// Test terminated timestamp
	beforeTerminate := time.Now().UTC()
	_ = dep.SetStatus(StatusTerminated)
	afterTerminate := time.Now().UTC()

	if dep.TerminatedAt() == nil {
		t.Error("TerminatedAt() should not be nil after StatusTerminated")
	} else if dep.TerminatedAt().Before(beforeTerminate) || dep.TerminatedAt().After(afterTerminate) {
		t.Error("TerminatedAt() timestamp is out of expected range")
	}
}

func TestDeployment_AddService(t *testing.T) {
	tests := []struct {
		name    string
		service DeploymentService
		wantErr bool
	}{
		{
			name: "valid service",
			service: DeploymentService{
				Name:     "frontend",
				Type:     "frontend",
				Provider: "vercel",
				Status:   "pending",
			},
			wantErr: false,
		},
		{
			name:    "empty name",
			service: DeploymentService{Name: "", Type: "frontend", Provider: "vercel"},
			wantErr: true,
		},
		{
			name:    "empty type",
			service: DeploymentService{Name: "frontend", Type: "", Provider: "vercel"},
			wantErr: true,
		},
		{
			name:    "empty provider",
			service: DeploymentService{Name: "frontend", Type: "frontend", Provider: ""},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dep := newTestDeployment(t)
			err := dep.AddService(tt.service)
			assertErr(t, err, tt.wantErr)
			if tt.wantErr {
				return
			}
			if len(dep.Services()) != 1 {
				t.Errorf("Services() length = %v, want 1", len(dep.Services()))
			}
		})
	}
}

func TestDeployment_AddService_Duplicate(t *testing.T) {
	dep := newTestDeployment(t)

	service := DeploymentService{Name: "frontend", Type: "frontend", Provider: "vercel"}
	if err := dep.AddService(service); err != nil {
		t.Fatalf("AddService() first call failed: %v", err)
	}

	if err := dep.AddService(service); err == nil {
		t.Error("AddService() should fail for duplicate service name")
	}
}

func TestDeployment_RemoveService(t *testing.T) {
	dep := newTestDeployment(t)

	service := DeploymentService{Name: "frontend", Type: "frontend", Provider: "vercel"}
	if err := dep.AddService(service); err != nil {
		t.Fatalf("AddService() setup failed: %v", err)
	}

	if err := dep.RemoveService("frontend"); err != nil {
		t.Errorf("RemoveService() unexpected error: %v", err)
	}
	if len(dep.Services()) != 0 {
		t.Errorf("Services() length = %v, want 0", len(dep.Services()))
	}

	if err := dep.RemoveService("backend"); err == nil {
		t.Error("RemoveService() should fail for non-existent service")
	}
}

func TestDeployment_SetEnvVar(t *testing.T) {
	dep := newTestDeployment(t)

	dep.SetEnvVar("API_KEY", "secret-key")
	dep.SetEnvVar("DB_URL", "postgres://...")

	envVars := dep.EnvVars()
	if len(envVars) != 2 {
		t.Errorf("EnvVars() length = %v, want 2", len(envVars))
	}
	if envVars["API_KEY"] != "secret-key" {
		t.Errorf("EnvVars()[API_KEY] = %v, want 'secret-key'", envVars["API_KEY"])
	}
}

func TestDeployment_CalculateTotalCost(t *testing.T) {
	dep := newTestDeployment(t)

	if dep.CalculateTotalCost() != 0.0 {
		t.Errorf("CalculateTotalCost() = %v, want 0.0", dep.CalculateTotalCost())
	}

	costInfo := &CostInfo{
		Monthly: 0.0,
		Breakdown: map[string]float64{
			"vercel": 0.0,
			"render": 7.0,
			"neon":   0.0,
		},
	}
	dep.SetCostInfo(costInfo)

	if got, want := dep.CalculateTotalCost(), 7.0; got != want {
		t.Errorf("CalculateTotalCost() = %v, want %v", got, want)
	}
}

func TestDeployment_IsActive(t *testing.T) {
	dep := newTestDeployment(t)

	if dep.IsActive() {
		t.Error("IsActive() = true for pending deployment, want false")
	}

	transitionTo(dep, StatusDeploying)
	if !dep.IsActive() {
		t.Error("IsActive() = false for deploying deployment, want true")
	}

	_ = dep.SetStatus(StatusDeployed)
	if !dep.IsActive() {
		t.Error("IsActive() = false for deployed deployment, want true")
	}

	_ = dep.SetStatus(StatusTerminated)
	if dep.IsActive() {
		t.Error("IsActive() = true for terminated deployment, want false")
	}
}

func TestDeployment_IsFailed(t *testing.T) {
	dep := newTestDeployment(t)

	if dep.IsFailed() {
		t.Error("IsFailed() = true for pending deployment, want false")
	}
	_ = dep.SetStatus(StatusFailed)
	if !dep.IsFailed() {
		t.Error("IsFailed() = false for failed deployment, want true")
	}
}

func TestDeployment_IsTerminated(t *testing.T) {
	dep := newTestDeployment(t)

	if dep.IsTerminated() {
		t.Error("IsTerminated() = true for pending deployment, want false")
	}
	_ = dep.SetStatus(StatusTerminated)
	if !dep.IsTerminated() {
		t.Error("IsTerminated() = false for terminated deployment, want true")
	}
}

func TestDeployment_Validate(t *testing.T) {
	tests := []struct {
		name    string
		dep     *Deployment
		wantErr bool
		errMsg  string
	}{
		{name: "valid deployment", dep: newTestDeployment(t), wantErr: false},
		{
			name:    "reconstructed valid deployment",
			dep:     reconstructForValidate("uuid-123", "test", "owner", StatusPending),
			wantErr: false,
		},
		{
			name:    "empty uuid",
			dep:     reconstructForValidate("", "test", "owner", StatusPending),
			wantErr: true,
			errMsg:  "UUID cannot be empty",
		},
		{
			name:    "empty name",
			dep:     reconstructForValidate("uuid-123", "", "owner", StatusPending),
			wantErr: true,
			errMsg:  "name cannot be empty",
		},
		{
			name:    "empty owner",
			dep:     reconstructForValidate("uuid-123", "test", "", StatusPending),
			wantErr: true,
			errMsg:  "owner cannot be empty",
		},
		{
			name:    "invalid status",
			dep:     reconstructForValidate("uuid-123", "test", "owner", Status("invalid-status")),
			wantErr: true,
			errMsg:  "invalid deployment status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.dep.Validate()
			assertErr(t, err, tt.wantErr)
			if tt.wantErr && tt.errMsg != "" && err != nil &&
				!strings.Contains(strings.ToLower(err.Error()), strings.ToLower(tt.errMsg)) {
				t.Errorf("Validate() error message should contain '%s', got '%s'", tt.errMsg, err.Error())
			}
		})
	}
}

// TestDeployment_SetEnvVar_NilMap tests env var setting with nil map
func TestDeployment_SetEnvVar_NilMap(t *testing.T) {
	dep := reconstructForValidate("uuid-123", "test", "owner", StatusPending)
	dep.envVars = nil

	dep.SetEnvVar("TEST_KEY", "test_value")

	if dep.EnvVars() == nil {
		t.Error("EnvVars() should not be nil after SetEnvVar")
	}
	if len(dep.EnvVars()) != 1 {
		t.Errorf("EnvVars() length = %d, want 1", len(dep.EnvVars()))
	}
	if dep.EnvVars()["TEST_KEY"] != "test_value" {
		t.Errorf("EnvVars()[TEST_KEY] = %s, want 'test_value'", dep.EnvVars()["TEST_KEY"])
	}
}

func TestReconstructDeployment(t *testing.T) {
	uuid := "uuid-123"
	name := "test-deployment"
	owner := "owner-456"
	status := StatusDeployed
	now := time.Now().UTC()
	deployedAt := time.Now().UTC()

	dep := ReconstructDeployment(uuid, name, owner, nil, status, now, now, &deployedAt, nil)

	if dep.UUID() != uuid {
		t.Errorf("UUID() = %v, want %v", dep.UUID(), uuid)
	}
	if dep.Name() != name {
		t.Errorf("Name() = %v, want %v", dep.Name(), name)
	}
	if dep.Owner() != owner {
		t.Errorf("Owner() = %v, want %v", dep.Owner(), owner)
	}
	if dep.Status() != status {
		t.Errorf("Status() = %v, want %v", dep.Status(), status)
	}
	if dep.DeployedAt() == nil || !dep.DeployedAt().Equal(deployedAt) {
		t.Error("DeployedAt() timestamp mismatch")
	}
}
