package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/steverhoton/unt-units-svc/internal/models"
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
		return appsync.NewErrorResponse("INVALID_INPUT", "Invalid input parameters", ""), nil
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
	if input.UnitType == "" {
		log.Printf("Missing required field: unitType")
		return appsync.NewErrorResponse("VALIDATION_ERROR", "UnitType is required", ""), nil
	}

	// Validate that the unitType corresponds to an available schema
	availableTypes, err := models.GetAvailableUnitTypes()
	if err != nil {
		log.Printf("Error getting available unit types: %v", err)
		return appsync.NewErrorResponse("INTERNAL_ERROR", "Failed to validate unit type", ""), nil
	}
	
	validType := false
	for _, availableType := range availableTypes {
		if availableType == input.UnitType {
			validType = true
			break
		}
	}
	if !validType {
		log.Printf("Invalid unit type: %s", input.UnitType)
		return appsync.NewErrorResponse("VALIDATION_ERROR", "Invalid unit type", ""), nil
	}

	// Create dynamic unit
	unit, err := models.NewDynamicUnit(input.UnitType)
	if err != nil {
		log.Printf("Error creating dynamic unit: %v", err)
		return appsync.NewErrorResponse("VALIDATION_ERROR", "Failed to create unit", ""), nil
	}

	// Set account ID and generate ID
	unit.AccountID = input.AccountID
	unit.GenerateID()

	// Validate and set data
	err = unit.ValidateAndSetData(input.Data)
	if err != nil {
		log.Printf("Error validating unit data: %v", err)
		return appsync.NewErrorResponse("VALIDATION_ERROR", "Data validation failed", ""), nil
	}

	// Attempt to create the unit
	err = h.repo.Create(ctx, unit)
	if err != nil {
		log.Printf("Error creating unit: %v", err)
		return appsync.NewErrorResponse("CREATE_FAILED", "Failed to create unit", ""), nil
	}

	// Convert to map for response
	responseData, err := unit.ToMap()
	if err != nil {
		log.Printf("Error converting unit to response: %v", err)
		return appsync.NewErrorResponse("INTERNAL_ERROR", "Failed to format response", ""), nil
	}

	log.Printf("Unit created successfully with ID: %s for account: %s", unit.ID, unit.AccountID)
	return appsync.NewSuccessResponse(responseData, "Unit created successfully"), nil
}

// HandleRead handles unit retrieval requests
func (h *UnitHandlers) HandleRead(ctx context.Context, event *appsync.AppSyncEvent) (*appsync.Response, error) {
	log.Printf("HandleRead called with event: %+v", event)

	// Parse arguments
	args, err := event.ParseArguments()
	if err != nil {
		log.Printf("Error parsing arguments: %v", err)
		return appsync.NewErrorResponse("INVALID_INPUT", "Invalid input parameters", ""), nil
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
	if input.UnitType == "" {
		log.Printf("Missing required field: unitType")
		return appsync.NewErrorResponse("VALIDATION_ERROR", "UnitType is required", ""), nil
	}

	// Retrieve the unit
	unit, err := h.repo.GetByID(ctx, input.AccountID, input.UnitType, input.ID)
	if err != nil {
		log.Printf("Error retrieving unit: %v", err)
		return appsync.NewErrorResponse("READ_FAILED", "Failed to retrieve unit", ""), nil
	}

	if unit == nil {
		log.Printf("Unit not found with ID: %s, unitType: %s for account: %s", input.ID, input.UnitType, input.AccountID)
		return appsync.NewErrorResponse("NOT_FOUND", "Unit not found", ""), nil
	}

	// Convert to map for response
	responseData, err := unit.ToMap()
	if err != nil {
		log.Printf("Error converting unit to response: %v", err)
		return appsync.NewErrorResponse("INTERNAL_ERROR", "Failed to format response", ""), nil
	}

	log.Printf("Unit retrieved successfully with ID: %s for account: %s", unit.ID, unit.AccountID)
	return appsync.NewSuccessResponse(responseData, "Unit retrieved successfully"), nil
}

// HandleUpdate handles unit update requests
func (h *UnitHandlers) HandleUpdate(ctx context.Context, event *appsync.AppSyncEvent) (*appsync.Response, error) {
	log.Printf("HandleUpdate called with event: %+v", event)

	// Parse arguments
	args, err := event.ParseArguments()
	if err != nil {
		log.Printf("Error parsing arguments: %v", err)
		return appsync.NewErrorResponse("INVALID_INPUT", "Invalid input parameters", ""), nil
	}

	input, ok := args.(appsync.UpdateUnitInput)
	if !ok {
		log.Printf("Invalid input type for update operation")
		return appsync.NewErrorResponse("INVALID_INPUT", "Invalid input type for update operation", ""), nil
	}

	// Validate required fields for update
	if input.ID == "" {
		log.Printf("Missing required field: id")
		return appsync.NewErrorResponse("VALIDATION_ERROR", "ID is required", ""), nil
	}
	if input.AccountID == "" {
		log.Printf("Missing required field: accountId")
		return appsync.NewErrorResponse("VALIDATION_ERROR", "AccountID is required", ""), nil
	}

	if input.UnitType == "" {
		log.Printf("Missing required field: unitType")
		return appsync.NewErrorResponse("VALIDATION_ERROR", "UnitType is required", ""), nil
	}

	// Validate that the unitType corresponds to an available schema
	availableTypes, err := models.GetAvailableUnitTypes()
	if err != nil {
		log.Printf("Error getting available unit types: %v", err)
		return appsync.NewErrorResponse("INTERNAL_ERROR", "Failed to validate unit type", ""), nil
	}
	
	validType := false
	for _, availableType := range availableTypes {
		if availableType == input.UnitType {
			validType = true
			break
		}
	}
	if !validType {
		log.Printf("Invalid unit type: %s", input.UnitType)
		return appsync.NewErrorResponse("VALIDATION_ERROR", "Invalid unit type", ""), nil
	}

	// Check if the unit exists before attempting to update
	existingUnit, err := h.repo.GetByID(ctx, input.AccountID, input.UnitType, input.ID)
	if err != nil {
		log.Printf("Error checking if unit exists: %v", err)
		return appsync.NewErrorResponse("UPDATE_FAILED", "Failed to verify unit existence", ""), nil
	}

	if existingUnit == nil {
		log.Printf("Unit not found with ID: %s, unitType: %s for account: %s", input.ID, input.UnitType, input.AccountID)
		return appsync.NewErrorResponse("NOT_FOUND", "Unit not found", ""), nil
	}

	// Merge the update data with existing data
	mergedData := make(map[string]interface{})
	
	// Start with existing data
	for k, v := range existingUnit.Data {
		mergedData[k] = v
	}
	
	// Override with new data
	for k, v := range input.Data {
		mergedData[k] = v
	}

	// Validate and set the merged data
	err = existingUnit.ValidateAndSetData(mergedData)
	if err != nil {
		log.Printf("Error validating updated unit data: %v", err)
		return appsync.NewErrorResponse("VALIDATION_ERROR", "Data validation failed", ""), nil
	}

	// Attempt to update the unit
	err = h.repo.Update(ctx, existingUnit)
	if err != nil {
		log.Printf("Error updating unit: %v", err)
		return appsync.NewErrorResponse("UPDATE_FAILED", "Failed to update unit", ""), nil
	}

	// Convert to map for response
	responseData, err := existingUnit.ToMap()
	if err != nil {
		log.Printf("Error converting unit to response: %v", err)
		return appsync.NewErrorResponse("INTERNAL_ERROR", "Failed to format response", ""), nil
	}

	log.Printf("Unit updated successfully with ID: %s for account: %s", existingUnit.ID, existingUnit.AccountID)
	return appsync.NewSuccessResponse(responseData, "Unit updated successfully"), nil
}

// HandleDelete handles unit deletion requests
func (h *UnitHandlers) HandleDelete(ctx context.Context, event *appsync.AppSyncEvent) (*appsync.Response, error) {
	log.Printf("HandleDelete called with event: %+v", event)

	// Parse arguments
	args, err := event.ParseArguments()
	if err != nil {
		log.Printf("Error parsing arguments: %v", err)
		return appsync.NewErrorResponse("INVALID_INPUT", "Invalid input parameters", ""), nil
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
	if input.UnitType == "" {
		log.Printf("Missing required field: unitType")
		return appsync.NewErrorResponse("VALIDATION_ERROR", "UnitType is required", ""), nil
	}

	// Attempt to delete the unit
	err = h.repo.Delete(ctx, input.AccountID, input.UnitType, input.ID)
	if err != nil {
		log.Printf("Error deleting unit: %v", err)
		return appsync.NewErrorResponse("DELETE_FAILED", "Failed to delete unit", ""), nil
	}

	response := map[string]interface{}{
		"id":        input.ID,
		"accountId": input.AccountID,
		"unitType":  input.UnitType,
		"deleted":   true,
	}

	log.Printf("Unit deleted successfully with ID: %s, unitType: %s for account: %s", input.ID, input.UnitType, input.AccountID)
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
