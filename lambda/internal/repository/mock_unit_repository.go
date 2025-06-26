package repository

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/steverhoton/unt-units-svc/internal/models"
	"github.com/steverhoton/unt-units-svc/pkg/appsync"
)

// MockUnitRepository is a mock implementation of UnitRepository for testing
type MockUnitRepository struct {
	mock.Mock
}

// Create mocks the Create method
func (m *MockUnitRepository) Create(ctx context.Context, unit *models.Unit) error {
	args := m.Called(ctx, unit)
	return args.Error(0)
}

// GetByKey mocks the GetByKey method
func (m *MockUnitRepository) GetByKey(ctx context.Context, id, accountID string) (*models.Unit, error) {
	args := m.Called(ctx, id, accountID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Unit), args.Error(1)
}

// Update mocks the Update method
func (m *MockUnitRepository) Update(ctx context.Context, unit *models.Unit) error {
	args := m.Called(ctx, unit)
	return args.Error(0)
}

// Delete mocks the Delete method
func (m *MockUnitRepository) Delete(ctx context.Context, id, accountID string) error {
	args := m.Called(ctx, id, accountID)
	return args.Error(0)
}

// List mocks the List method
func (m *MockUnitRepository) List(ctx context.Context, input *appsync.ListUnitsInput) (*appsync.ListUnitsResponse, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*appsync.ListUnitsResponse), args.Error(1)
}

// Exists mocks the Exists method
func (m *MockUnitRepository) Exists(ctx context.Context, id, accountID string) (bool, error) {
	args := m.Called(ctx, id, accountID)
	return args.Bool(0), args.Error(1)
}
