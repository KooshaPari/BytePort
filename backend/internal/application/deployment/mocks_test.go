package deployment

import (
	"context"

	"github.com/byteport/api/internal/domain/deployment"
)

// =============================================================================
// MockRepository — shared test double for deployment.Repository.
// Each method delegates to its corresponding *Func field; if the field is
// nil, the method returns a benign default. Tests configure only the
// behaviour they care about.
// =============================================================================

// MockRepository is a mock implementation of deployment.Repository.
type MockRepository struct {
	CreateFunc       func(ctx context.Context, dep *deployment.Deployment) error
	UpdateFunc       func(ctx context.Context, dep *deployment.Deployment) error
	FindByUUIDFunc   func(ctx context.Context, uuid string) (*deployment.Deployment, error)
	FindByOwnerFunc  func(ctx context.Context, owner string) ([]*deployment.Deployment, error)
	FindByStatusFunc func(ctx context.Context, status deployment.Status) ([]*deployment.Deployment, error)
	ListFunc         func(ctx context.Context, offset, limit int) ([]*deployment.Deployment, error)
	CountFunc        func(ctx context.Context) (int64, error)
	CountByOwnerFunc func(ctx context.Context, owner string) (int64, error)
}

func (m *MockRepository) Create(ctx context.Context, dep *deployment.Deployment) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, dep)
	}
	return nil
}

func (m *MockRepository) Update(ctx context.Context, dep *deployment.Deployment) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, dep)
	}
	return nil
}

func (m *MockRepository) Delete(ctx context.Context, uuid string) error {
	return nil
}

func (m *MockRepository) FindByUUID(ctx context.Context, uuid string) (*deployment.Deployment, error) {
	if m.FindByUUIDFunc != nil {
		return m.FindByUUIDFunc(ctx, uuid)
	}
	return nil, deployment.NewDeploymentNotFoundError(uuid)
}

func (m *MockRepository) FindByOwner(ctx context.Context, owner string) ([]*deployment.Deployment, error) {
	if m.FindByOwnerFunc != nil {
		return m.FindByOwnerFunc(ctx, owner)
	}
	return []*deployment.Deployment{}, nil
}

func (m *MockRepository) FindByProject(ctx context.Context, projectUUID string) ([]*deployment.Deployment, error) {
	return []*deployment.Deployment{}, nil
}

func (m *MockRepository) FindByStatus(ctx context.Context, status deployment.Status) ([]*deployment.Deployment, error) {
	if m.FindByStatusFunc != nil {
		return m.FindByStatusFunc(ctx, status)
	}
	return []*deployment.Deployment{}, nil
}

func (m *MockRepository) List(ctx context.Context, offset, limit int) ([]*deployment.Deployment, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, offset, limit)
	}
	return []*deployment.Deployment{}, nil
}

func (m *MockRepository) Count(ctx context.Context) (int64, error) {
	if m.CountFunc != nil {
		return m.CountFunc(ctx)
	}
	return 0, nil
}

func (m *MockRepository) CountByOwner(ctx context.Context, owner string) (int64, error) {
	if m.CountByOwnerFunc != nil {
		return m.CountByOwnerFunc(ctx, owner)
	}
	return 0, nil
}

// =============================================================================
// MockService — shared test double for deployment.Service.
// =============================================================================

// MockService is a mock implementation of deployment.Service.
type MockService struct {
	ValidateDeploymentFunc      func(ctx context.Context, dep *deployment.Deployment) error
	CanUserAccessDeploymentFunc func(ctx context.Context, userUUID, deploymentUUID string) (bool, error)
	CalculateEstimatedCostFunc  func(ctx context.Context, dep *deployment.Deployment) (*deployment.CostInfo, error)
	SelectOptimalProviderFunc   func(ctx context.Context, serviceType string, constraints map[string]interface{}) (string, error)
}

func (m *MockService) ValidateDeployment(ctx context.Context, dep *deployment.Deployment) error {
	if m.ValidateDeploymentFunc != nil {
		return m.ValidateDeploymentFunc(ctx, dep)
	}
	return nil
}

func (m *MockService) CanUserAccessDeployment(ctx context.Context, userUUID, deploymentUUID string) (bool, error) {
	if m.CanUserAccessDeploymentFunc != nil {
		return m.CanUserAccessDeploymentFunc(ctx, userUUID, deploymentUUID)
	}
	return true, nil
}

func (m *MockService) CalculateEstimatedCost(ctx context.Context, dep *deployment.Deployment) (*deployment.CostInfo, error) {
	if m.CalculateEstimatedCostFunc != nil {
		return m.CalculateEstimatedCostFunc(ctx, dep)
	}
	return &deployment.CostInfo{Monthly: 0, Breakdown: map[string]float64{}}, nil
}

func (m *MockService) SelectOptimalProvider(ctx context.Context, serviceType string, constraints map[string]interface{}) (string, error) {
	if m.SelectOptimalProviderFunc != nil {
		return m.SelectOptimalProviderFunc(ctx, serviceType, constraints)
	}
	return "default", nil
}
