package container

import (
	"os"

	"github.com/byteport/api/internal/application/deployment"
	domaindep "github.com/byteport/api/internal/domain/deployment"
	"github.com/byteport/api/internal/infrastructure/clients"
	"github.com/byteport/api/internal/infrastructure/http/handlers"
	"github.com/byteport/api/internal/infrastructure/persistence/postgres"
	"gorm.io/gorm"
)

// Container holds all dependencies
type Container struct {
	// Database
	DB *gorm.DB

	// Repositories
	DeploymentRepository domaindep.Repository

	// Domain Services
	DeploymentDomainService domaindep.Service

	// Use Cases
	CreateDeploymentUseCase    *deployment.CreateDeploymentUseCase
	GetDeploymentUseCase       *deployment.GetDeploymentUseCase
	ListDeploymentsUseCase     *deployment.ListDeploymentsUseCase
	TerminateDeploymentUseCase *deployment.TerminateDeploymentUseCase
	UpdateStatusUseCase        *deployment.UpdateStatusUseCase

	// Engine daemon client (nil when BYTEPORT_ENGINE_SOCKET is not set)
	EngineDaemonClient *clients.EngineDaemonClient

	// HTTP Handlers
	DeploymentHandler    *handlers.DeploymentHandler
	EngineDeployHandler  *handlers.EngineDeploymentHandler
	WebhookHandler       *handlers.GitHubWebhookHandler
}

// NewContainer creates a new dependency injection container
func NewContainer(db *gorm.DB) *Container {
	c := &Container{
		DB: db,
	}

	// Initialize dependencies in order
	c.initRepositories()
	c.initDomainServices()
	c.initUseCases()
	c.initEngineClient()
	c.initHandlers()

	return c
}

// initRepositories initializes repository implementations
func (c *Container) initRepositories() {
	c.DeploymentRepository = postgres.NewDeploymentRepository(c.DB)
}

// initDomainServices initializes domain services
func (c *Container) initDomainServices() {
	c.DeploymentDomainService = domaindep.NewDomainService(c.DeploymentRepository)
}

// initUseCases initializes application use cases
func (c *Container) initUseCases() {
	c.CreateDeploymentUseCase = deployment.NewCreateDeploymentUseCase(
		c.DeploymentRepository,
		c.DeploymentDomainService,
	)

	c.GetDeploymentUseCase = deployment.NewGetDeploymentUseCase(
		c.DeploymentRepository,
		c.DeploymentDomainService,
	)

	c.ListDeploymentsUseCase = deployment.NewListDeploymentsUseCase(
		c.DeploymentRepository,
	)

	c.TerminateDeploymentUseCase = deployment.NewTerminateDeploymentUseCase(
		c.DeploymentRepository,
		c.DeploymentDomainService,
	)

	c.UpdateStatusUseCase = deployment.NewUpdateStatusUseCase(
		c.DeploymentRepository,
		c.DeploymentDomainService,
	)
}

// initEngineClient initializes the engine daemon client when the socket
// env var is set.  Should be called before initHandlers.
func (c *Container) initEngineClient() {
	if os.Getenv("BYTEPORT_ENGINE_SOCKET") != "" {
		c.EngineDaemonClient = clients.NewEngineDaemonClient()
	}
}

// initHandlers initializes HTTP handlers
func (c *Container) initHandlers() {
	c.DeploymentHandler = handlers.NewDeploymentHandler(
		c.CreateDeploymentUseCase,
		c.GetDeploymentUseCase,
		c.ListDeploymentsUseCase,
		c.TerminateDeploymentUseCase,
		c.UpdateStatusUseCase,
		c.EngineDaemonClient,
	)

	c.WebhookHandler = handlers.NewGitHubWebhookHandler(c.CreateDeploymentUseCase)

	if c.EngineDaemonClient != nil {
		c.EngineDeployHandler = handlers.NewEngineDeploymentHandler(c.EngineDaemonClient)
	}
}
