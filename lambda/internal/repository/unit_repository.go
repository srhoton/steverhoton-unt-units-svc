package repository

import (
	"context"

	"github.com/steverhoton/unt-units-svc/internal/models"
	"github.com/steverhoton/unt-units-svc/pkg/appsync"
)

// UnitRepository defines the interface for unit data operations
type UnitRepository interface {
	// Create creates a new unit in the repository
	Create(ctx context.Context, unit *models.DynamicUnit) error

	// GetByKey retrieves a unit by its composite primary key (accountId + unitType#id)
	GetByKey(ctx context.Context, accountID, sortKey string) (*models.DynamicUnit, error)

	// GetByID retrieves a unit by its unique ID within an account and unitType
	GetByID(ctx context.Context, accountID, unitType, unitID string) (*models.DynamicUnit, error)

	// Update updates an existing unit in the repository
	Update(ctx context.Context, unit *models.DynamicUnit) error

	// Delete soft deletes a unit (marks deletedAt timestamp)
	Delete(ctx context.Context, accountID, unitType, unitID string) error

	// List retrieves a paginated list of units for an account
	List(ctx context.Context, input *appsync.ListUnitsInput) (*appsync.ListUnitsResponse, error)

	// Exists checks if a unit exists by its composite primary key
	Exists(ctx context.Context, accountID, sortKey string) (bool, error)
}
