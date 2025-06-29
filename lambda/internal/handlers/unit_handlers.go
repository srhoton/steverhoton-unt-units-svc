package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/steverhoton/unt-units-svc/internal/repository"
	"github.com/steverhoton/unt-units-svc/pkg/appsync"
)

// UnitHandlers contains handlers for unit CRUD operations
type UnitHandlers struct {
	repo repository.UnitRepository
}

// NewUnitHandlers creates a new instance of UnitHandlers
func NewUnitHandlers(repo repository.UnitRepository) *UnitHandlers {
	return &UnitHandlers{
		repo: repo,
	}
}

// HandleCreate handles unit creation requests
func (h *UnitHandlers) HandleCreate(ctx context.Context, event *appsync.AppSyncEvent) (*appsync.Response, error) {
	log.Printf("HandleCreate called with event: %+v", event)

	// Parse arguments
	args, err := event.ParseArguments()
	if err != nil {
		log.Printf("Error parsing arguments: %v", err)
		return appsync.NewErrorResponse("INVALID_INPUT", "Invalid input parameters", err.Error()), nil
	}

	input, ok := args.(appsync.CreateUnitInput)
	if !ok {
		log.Printf("Invalid input type for create operation")
		return appsync.NewErrorResponse("INVALID_INPUT", "Invalid input type for create operation", ""), nil
	}

	// Validate required fields
	if input.AccountID == "" {
		log.Printf("Missing required field: accountId")
		return appsync.NewErrorResponse("VALIDATION_ERROR", "AccountID is required", ""), nil
	}
	if input.SuggestedVin == "" {
		log.Printf("Missing required field: suggestedVin")
		return appsync.NewErrorResponse("VALIDATION_ERROR", "SuggestedVin is required", ""), nil
	}

	// Since Unit is embedded in CreateUnitInput, extract the Unit fields
	unit := input.Unit
	unit.AccountID = input.AccountID

	// Attempt to create the unit
	err = h.repo.Create(ctx, &unit)
	if err != nil {
		log.Printf("Error creating unit: %v", err)
		return appsync.NewErrorResponse("CREATE_FAILED", "Failed to create unit", err.Error()), nil
	}

	log.Printf("Unit created successfully with ID: %s for account: %s", unit.ID, unit.AccountID)
	return appsync.NewSuccessResponse(unit, "Unit created successfully"), nil
}

// HandleRead handles unit retrieval requests
func (h *UnitHandlers) HandleRead(ctx context.Context, event *appsync.AppSyncEvent) (*appsync.Response, error) {
	log.Printf("HandleRead called with event: %+v", event)

	// Parse arguments
	args, err := event.ParseArguments()
	if err != nil {
		log.Printf("Error parsing arguments: %v", err)
		return appsync.NewErrorResponse("INVALID_INPUT", "Invalid input parameters", err.Error()), nil
	}

	input, ok := args.(appsync.GetUnitInput)
	if !ok {
		log.Printf("Invalid input type for read operation")
		return appsync.NewErrorResponse("INVALID_INPUT", "Invalid input type for read operation", ""), nil
	}

	// Validate required fields
	if input.ID == "" {
		log.Printf("Missing required field: id")
		return appsync.NewErrorResponse("VALIDATION_ERROR", "ID is required", ""), nil
	}
	if input.AccountID == "" {
		log.Printf("Missing required field: accountId")
		return appsync.NewErrorResponse("VALIDATION_ERROR", "AccountID is required", ""), nil
	}

	// Retrieve the unit
	unit, err := h.repo.GetByID(ctx, input.AccountID, input.ID)
	if err != nil {
		log.Printf("Error retrieving unit: %v", err)
		return appsync.NewErrorResponse("READ_FAILED", "Failed to retrieve unit", err.Error()), nil
	}

	if unit == nil {
		log.Printf("Unit not found with ID: %s for account: %s", input.ID, input.AccountID)
		return appsync.NewErrorResponse("NOT_FOUND", "Unit not found", ""), nil
	}

	log.Printf("Unit retrieved successfully with ID: %s for account: %s", unit.ID, unit.AccountID)
	return appsync.NewSuccessResponse(unit, "Unit retrieved successfully"), nil
}

// HandleUpdate handles unit update requests
func (h *UnitHandlers) HandleUpdate(ctx context.Context, event *appsync.AppSyncEvent) (*appsync.Response, error) {
	log.Printf("HandleUpdate called with event: %+v", event)

	// Parse arguments
	args, err := event.ParseArguments()
	if err != nil {
		log.Printf("Error parsing arguments: %v", err)
		return appsync.NewErrorResponse("INVALID_INPUT", "Invalid input parameters", err.Error()), nil
	}

	input, ok := args.(appsync.UpdateUnitInput)
	if !ok {
		log.Printf("Invalid input type for update operation")
		return appsync.NewErrorResponse("INVALID_INPUT", "Invalid input type for update operation", ""), nil
	}

	// Validate required fields for update (only ID and AccountID are required)
	if input.ID == "" {
		log.Printf("Missing required field: id")
		return appsync.NewErrorResponse("VALIDATION_ERROR", "ID is required", ""), nil
	}
	if input.AccountID == "" {
		log.Printf("Missing required field: accountId")
		return appsync.NewErrorResponse("VALIDATION_ERROR", "AccountID is required", ""), nil
	}
	// Note: suggestedVin is not required for updates as they can be partial

	// Check if the unit exists before attempting to update
	existingUnit, err := h.repo.GetByID(ctx, input.AccountID, input.ID)
	if err != nil {
		log.Printf("Error checking if unit exists: %v", err)
		return appsync.NewErrorResponse("UPDATE_FAILED", "Failed to verify unit existence", err.Error()), nil
	}
	if existingUnit == nil {
		log.Printf("Unit not found with ID: %s for account: %s", input.ID, input.AccountID)
		return appsync.NewErrorResponse("NOT_FOUND", "Unit not found", ""), nil
	}

	// Merge the update data with the existing unit (partial update)
	updatedUnit := *existingUnit // Copy existing unit

	// Apply only the fields that were provided in the input
	if input.SuggestedVin != "" {
		updatedUnit.SuggestedVin = input.SuggestedVin
	}
	if input.ErrorCode != "" {
		updatedUnit.ErrorCode = input.ErrorCode
	}
	if input.PossibleValues != "" {
		updatedUnit.PossibleValues = input.PossibleValues
	}
	if input.ErrorText != "" {
		updatedUnit.ErrorText = input.ErrorText
	}
	if input.VehicleDescriptor != "" {
		updatedUnit.VehicleDescriptor = input.VehicleDescriptor
	}
	if input.Note != "" {
		updatedUnit.Note = input.Note
	}
	if input.Make != "" {
		updatedUnit.Make = input.Make
	}
	if input.ManufacturerName != "" {
		updatedUnit.ManufacturerName = input.ManufacturerName
	}
	if input.Model != "" {
		updatedUnit.Model = input.Model
	}
	if input.ModelYear != "" {
		updatedUnit.ModelYear = input.ModelYear
	}
	if input.Series != "" {
		updatedUnit.Series = input.Series
	}
	if input.VehicleType != "" {
		updatedUnit.VehicleType = input.VehicleType
	}
	// Add more fields as needed for the update...

	// Ensure the unit key matches the input
	updatedUnit.ID = input.ID
	updatedUnit.AccountID = input.AccountID

	// Attempt to update the unit
	err = h.repo.Update(ctx, &updatedUnit)
	if err != nil {
		log.Printf("Error updating unit: %v", err)
		return appsync.NewErrorResponse("UPDATE_FAILED", "Failed to update unit", err.Error()), nil
	}

	log.Printf("Unit updated successfully with ID: %s for account: %s", updatedUnit.ID, updatedUnit.AccountID)
	return appsync.NewSuccessResponse(updatedUnit, "Unit updated successfully"), nil
}

// HandleDelete handles unit deletion requests
func (h *UnitHandlers) HandleDelete(ctx context.Context, event *appsync.AppSyncEvent) (*appsync.Response, error) {
	log.Printf("HandleDelete called with event: %+v", event)

	// Parse arguments
	args, err := event.ParseArguments()
	if err != nil {
		log.Printf("Error parsing arguments: %v", err)
		return appsync.NewErrorResponse("INVALID_INPUT", "Invalid input parameters", err.Error()), nil
	}

	input, ok := args.(appsync.DeleteUnitInput)
	if !ok {
		log.Printf("Invalid input type for delete operation")
		return appsync.NewErrorResponse("INVALID_INPUT", "Invalid input type for delete operation", ""), nil
	}

	// Validate required fields
	if input.ID == "" {
		log.Printf("Missing required field: id")
		return appsync.NewErrorResponse("VALIDATION_ERROR", "ID is required", ""), nil
	}
	if input.AccountID == "" {
		log.Printf("Missing required field: accountId")
		return appsync.NewErrorResponse("VALIDATION_ERROR", "AccountID is required", ""), nil
	}

	// First, get the unit to find its locationID
	unit, err := h.repo.GetByID(ctx, input.AccountID, input.ID)
	if err != nil {
		log.Printf("Error finding unit to delete: %v", err)
		return appsync.NewErrorResponse("DELETE_FAILED", "Failed to find unit for deletion", err.Error()), nil
	}
	if unit == nil {
		log.Printf("Unit not found with ID: %s for account: %s", input.ID, input.AccountID)
		return appsync.NewErrorResponse("NOT_FOUND", "Unit not found", ""), nil
	}

	// Attempt to delete the unit
	err = h.repo.Delete(ctx, input.AccountID, unit.LocationID)
	if err != nil {
		log.Printf("Error deleting unit: %v", err)
		return appsync.NewErrorResponse("DELETE_FAILED", "Failed to delete unit", err.Error()), nil
	}

	response := map[string]interface{}{
		"id":        input.ID,
		"accountId": input.AccountID,
		"deleted":   true,
	}

	log.Printf("Unit deleted successfully with ID: %s for account: %s", input.ID, input.AccountID)
	return appsync.NewSuccessResponse(response, "Unit deleted successfully"), nil
}

// HandleList handles unit listing requests
func (h *UnitHandlers) HandleList(ctx context.Context, event *appsync.AppSyncEvent) (*appsync.Response, error) {
	log.Printf("HandleList called with event: %+v", event)

	// Parse arguments
	args, err := event.ParseArguments()
	if err != nil {
		log.Printf("Error parsing arguments: %v", err)
		return appsync.NewErrorResponse("INVALID_INPUT", "Invalid input parameters", err.Error()), nil
	}

	input, ok := args.(appsync.ListUnitsInput)
	if !ok {
		log.Printf("Invalid input type for list operation")
		return appsync.NewErrorResponse("INVALID_INPUT", "Invalid input type for list operation", ""), nil
	}

	// Validate required fields
	if input.AccountID == "" {
		log.Printf("Missing required field: accountId")
		return appsync.NewErrorResponse("VALIDATION_ERROR", "AccountID is required", ""), nil
	}

	// Retrieve the list of units
	result, err := h.repo.List(ctx, &input)
	if err != nil {
		log.Printf("Error listing units: %v", err)
		return appsync.NewErrorResponse("LIST_FAILED", "Failed to list units", err.Error()), nil
	}

	log.Printf("Units listed successfully: %d items", result.Count)
	return appsync.NewSuccessResponse(result, fmt.Sprintf("Retrieved %d units", result.Count)), nil
}

// DumpEvent logs the complete event for debugging purposes
func (h *UnitHandlers) DumpEvent(ctx context.Context, event *appsync.AppSyncEvent) {
	log.Printf("=== EVENT DUMP START ===")
	log.Printf("TypeName: %s", event.TypeName)
	log.Printf("FieldName: %s", event.FieldName)
	log.Printf("Operation Type: %s", event.GetOperationType())

	// Pretty print arguments
	if len(event.Arguments) > 0 {
		var prettyArgs interface{}
		if err := json.Unmarshal(event.Arguments, &prettyArgs); err == nil {
			if argsJSON, err := json.MarshalIndent(prettyArgs, "", "  "); err == nil {
				log.Printf("Arguments:\n%s", string(argsJSON))
			}
		}
	}

	// Log identity information
	log.Printf("Identity Sub: %s", event.Identity.Sub)
	log.Printf("Identity Username: %s", event.Identity.Username)
	log.Printf("Account ID: %s", event.Identity.AccountID)
	log.Printf("Source IP: %v", event.Identity.SourceIP)

	// Log request headers
	if len(event.Request.Headers) > 0 {
		log.Printf("Request Headers:")
		for key, value := range event.Request.Headers {
			log.Printf("  %s: %s", key, value)
		}
	}

	// Log GraphQL info
	log.Printf("GraphQL Field: %s", event.Info.FieldName)
	log.Printf("GraphQL Parent Type: %s", event.Info.ParentTypeName)
	if len(event.Info.Variables) > 0 {
		if varsJSON, err := json.MarshalIndent(event.Info.Variables, "", "  "); err == nil {
			log.Printf("GraphQL Variables:\n%s", string(varsJSON))
		}
	}

	log.Printf("=== EVENT DUMP END ===")
}
