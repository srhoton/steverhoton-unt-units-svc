package repository

import (
	"context"

	"github.com/steverhoton/unt-units-svc/internal/models"
	"github.com/steverhoton/unt-units-svc/pkg/appsync"
)

// UnitRepository defines the interface for unit data operations
type UnitRepository interface {
	// Create creates a new unit in the repository
	Create(ctx context.Context, unit *models.Unit) error

	// GetByKey retrieves a unit by its composite primary key (accountID + unitID + unitType)
	GetByKey(ctx context.Context, accountID, unitID, unitType string) (*models.Unit, error)

	// Update updates an existing unit in the repository
	Update(ctx context.Context, unit *models.Unit) error

	// Delete soft deletes a unit (marks deletedAt timestamp)
	Delete(ctx context.Context, accountID, unitID, unitType string) error

	// List retrieves a paginated list of units for an account
	List(ctx context.Context, input *appsync.ListUnitsInput) (*appsync.ListUnitsResponse, error)

	// Exists checks if a unit exists by its composite primary key
	Exists(ctx context.Context, accountID, unitID, unitType string) (bool, error)

	// GetByUnitID retrieves all units with the given unit ID across all accounts and types
	GetByUnitID(ctx context.Context, unitID string) ([]models.Unit, error)
}
