package container

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestNewContainer tests container initialization
func TestNewContainer(t *testing.T) {
	// Create an in-memory SQLite database for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	container := NewContainer(db)

	if container == nil {
		t.Fatal("Expected container to be created, got nil")
	}

	// Verify database is set
	if container.DB == nil {
		t.Error("Expected DB to be set")
	}

	// Verify repositories are initialized
	if container.DeploymentRepository == nil {
		t.Error("Expected DeploymentRepository to be initialized")
	}

	// Verify domain services are initialized
	if container.DeploymentDomainService == nil {
		t.Error("Expected DeploymentDomainService to be initialized")
	}

	// Verify use cases are initialized
	if container.CreateDeploymentUseCase == nil {
		t.Error("Expected CreateDeploymentUseCase to be initialized")
	}
	if container.GetDeploymentUseCase == nil {
		t.Error("Expected GetDeploymentUseCase to be initialized")
	}
	if container.ListDeploymentsUseCase == nil {
		t.Error("Expected ListDeploymentsUseCase to be initialized")
	}
	if container.TerminateDeploymentUseCase == nil {
		t.Error("Expected TerminateDeploymentUseCase to be initialized")
	}
	if container.UpdateStatusUseCase == nil {
		t.Error("Expected UpdateStatusUseCase to be initialized")
	}

	// Verify handlers are initialized
	if container.DeploymentHandler == nil {
		t.Error("Expected DeploymentHandler to be initialized")
	}

	// Engine daemon client is nil when env is not set
	if container.EngineDaemonClient != nil {
		t.Error("Expected EngineDaemonClient to be nil without BYTEPORT_ENGINE_SOCKET")
	}
	if container.EngineDeployHandler != nil {
		t.Error("Expected EngineDeployHandler to be nil without BYTEPORT_ENGINE_SOCKET")
	}
}

// TestInitRepositories tests repository initialization
func TestInitRepositories(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	container := &Container{DB: db}
	container.initRepositories()

	if container.DeploymentRepository == nil {
		t.Error("Expected DeploymentRepository to be initialized")
	}
}

// TestInitDomainServices tests domain service initialization
func TestInitDomainServices(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	container := &Container{DB: db}
	container.initRepositories()
	container.initDomainServices()

	if container.DeploymentDomainService == nil {
		t.Error("Expected DeploymentDomainService to be initialized")
	}
}

// TestInitUseCases tests use case initialization
func TestInitUseCases(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	container := &Container{DB: db}
	container.initRepositories()
	container.initDomainServices()
	container.initUseCases()

	if container.CreateDeploymentUseCase == nil {
		t.Error("Expected CreateDeploymentUseCase to be initialized")
	}
	if container.GetDeploymentUseCase == nil {
		t.Error("Expected GetDeploymentUseCase to be initialized")
	}
	if container.ListDeploymentsUseCase == nil {
		t.Error("Expected ListDeploymentsUseCase to be initialized")
	}
	if container.TerminateDeploymentUseCase == nil {
		t.Error("Expected TerminateDeploymentUseCase to be initialized")
	}
	if container.UpdateStatusUseCase == nil {
		t.Error("Expected UpdateStatusUseCase to be initialized")
	}
}

// TestInitEngineClient verifies engine client is initialized only when
// the env var is set.
func TestInitEngineClient(t *testing.T) {
	t.Run("not initialized when env unset", func(t *testing.T) {
		t.Setenv("BYTEPORT_ENGINE_SOCKET", "")
		c := &Container{}
		c.initEngineClient()
		if c.EngineDaemonClient != nil {
			t.Error("Expected EngineDaemonClient to be nil without BYTEPORT_ENGINE_SOCKET")
		}
	})

	t.Run("initialized when env set", func(t *testing.T) {
		t.Setenv("BYTEPORT_ENGINE_SOCKET", "/tmp/test-engine.sock")
		c := &Container{}
		c.initEngineClient()
		if c.EngineDaemonClient == nil {
			t.Fatal("Expected EngineDaemonClient to be initialized with BYTEPORT_ENGINE_SOCKET")
		}
		if c.EngineDaemonClient.SocketPath() != "/tmp/test-engine.sock" {
			t.Errorf("Expected socket path /tmp/test-engine.sock, got %s", c.EngineDaemonClient.SocketPath())
		}
	})
}

// TestInitHandlers tests handler initialization
func TestInitHandlers(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	container := &Container{DB: db}
	container.initRepositories()
	container.initDomainServices()
	container.initUseCases()
	container.initEngineClient()
	container.initHandlers()

	if container.DeploymentHandler == nil {
		t.Error("Expected DeploymentHandler to be initialized")
	}

	// Without BYTEPORT_ENGINE_SOCKET, the engine handler stays nil.
	if container.EngineDeployHandler != nil {
		t.Error("Expected EngineDeployHandler to be nil without BYTEPORT_ENGINE_SOCKET")
	}
}

// TestContainerIntegration tests full integration
func TestContainerIntegration(t *testing.T) {
	// This is a smoke test to ensure all dependencies wire correctly
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	container := NewContainer(db)

	// Verify the entire dependency chain is properly wired
	// Repository -> Domain Service -> Use Cases -> Handlers
	if container.DeploymentHandler == nil {
		t.Fatal("Handler not initialized")
	}

	// All dependencies should be non-nil at this point
	tests := []struct {
		name  string
		value interface{}
	}{
		{"DB", container.DB},
		{"DeploymentRepository", container.DeploymentRepository},
		{"DeploymentDomainService", container.DeploymentDomainService},
		{"CreateDeploymentUseCase", container.CreateDeploymentUseCase},
		{"GetDeploymentUseCase", container.GetDeploymentUseCase},
		{"ListDeploymentsUseCase", container.ListDeploymentsUseCase},
		{"TerminateDeploymentUseCase", container.TerminateDeploymentUseCase},
		{"UpdateStatusUseCase", container.UpdateStatusUseCase},
		{"DeploymentHandler", container.DeploymentHandler},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value == nil {
				t.Errorf("%s is nil", tt.name)
			}
		})
	}

	// Engine daemon fields should be nil without BYTEPORT_ENGINE_SOCKET.
	if container.EngineDaemonClient != nil {
		t.Error("Expected EngineDaemonClient to be nil without BYTEPORT_ENGINE_SOCKET")
	}
	if container.EngineDeployHandler != nil {
		t.Error("Expected EngineDeployHandler to be nil without BYTEPORT_ENGINE_SOCKET")
	}
}
