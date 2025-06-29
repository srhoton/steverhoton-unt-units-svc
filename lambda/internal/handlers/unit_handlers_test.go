package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/steverhoton/unt-units-svc/internal/models"
	"github.com/steverhoton/unt-units-svc/internal/repository"
	"github.com/steverhoton/unt-units-svc/pkg/appsync"
)

func TestUnitHandlers_HandleCreate_Success(t *testing.T) {
	mockRepo := &repository.MockUnitRepository{}
	handlers := NewUnitHandlers(mockRepo)

	input := appsync.CreateUnitInput{
		AccountID: "test-account-123",
		Unit: models.Unit{
			LocationID:   "550e8400-e29b-41d4-a716-446655440000", // Valid UUID
			SuggestedVin: "1HGBH41JXMN109186",
			Make:         "Honda",
			Model:        "Civic",
		},
	}
	argsJSON, err := json.Marshal(input)
	require.NoError(t, err)

	event := &appsync.AppSyncEvent{
		TypeName:  "Mutation",
		FieldName: "createUnit",
		Arguments: argsJSON,
	}

	// Expected unit should have AccountID set from input
	expectedUnit := input.Unit
	expectedUnit.AccountID = input.AccountID

	// Mock expectations
	mockRepo.On("Create", mock.Anything, &expectedUnit).Return(nil)

	// Execute
	response, err := handlers.HandleCreate(context.Background(), event)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.True(t, response.Success)
	assert.Equal(t, "Unit created successfully", response.Message)
	assert.Nil(t, response.Error)

	// Verify mock expectations
	mockRepo.AssertExpectations(t)
}

func TestUnitHandlers_HandleCreate_ValidationError(t *testing.T) {
	mockRepo := &repository.MockUnitRepository{}
	handlers := NewUnitHandlers(mockRepo)

	tests := []struct {
		name          string
		input         appsync.CreateUnitInput
		expectedError string
	}{
		{
			name: "missing accountId",
			input: appsync.CreateUnitInput{
				AccountID: "", // Missing
				Unit: models.Unit{
					LocationID:   "550e8400-e29b-41d4-a716-446655440000",
					SuggestedVin: "1HGBH41JXMN109186",
					Make:         "Honda",
					Model:        "Civic",
				},
			},
			expectedError: "AccountID is required",
		},
		{
			name: "missing suggestedVin",
			input: appsync.CreateUnitInput{
				AccountID: "test-account-123",
				Unit: models.Unit{
					LocationID:   "550e8400-e29b-41d4-a716-446655440000",
					SuggestedVin: "", // Missing
					Make:         "Honda",
					Model:        "Civic",
				},
			},
			expectedError: "SuggestedVin is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			argsJSON, err := json.Marshal(tt.input)
			require.NoError(t, err)

			event := &appsync.AppSyncEvent{
				TypeName:  "Mutation",
				FieldName: "createUnit",
				Arguments: argsJSON,
			}

			// Execute
			response, err := handlers.HandleCreate(context.Background(), event)

			// Assertions
			require.NoError(t, err)
			require.NotNil(t, response)
			assert.False(t, response.Success)
			assert.NotNil(t, response.Error)
			assert.Equal(t, "VALIDATION_ERROR", response.Error.Code)
			assert.Contains(t, response.Error.Message, tt.expectedError)

			// No repository calls should be made
			mockRepo.AssertNotCalled(t, "Create")
		})
	}
}

func TestUnitHandlers_HandleCreate_RepositoryError(t *testing.T) {
	mockRepo := &repository.MockUnitRepository{}
	handlers := NewUnitHandlers(mockRepo)

	input := appsync.CreateUnitInput{
		AccountID: "test-account-123",
		Unit: models.Unit{
			LocationID:   "550e8400-e29b-41d4-a716-446655440000",
			SuggestedVin: "1HGBH41JXMN109186",
			Make:         "Honda",
			Model:        "Civic",
		},
	}
	argsJSON, err := json.Marshal(input)
	require.NoError(t, err)

	event := &appsync.AppSyncEvent{
		TypeName:  "Mutation",
		FieldName: "createUnit",
		Arguments: argsJSON,
	}

	// Expected unit should have AccountID set
	expectedUnit := input.Unit
	expectedUnit.AccountID = input.AccountID

	// Mock expectations - repository returns error
	mockRepo.On("Create", mock.Anything, &expectedUnit).Return(errors.New("database connection failed"))

	// Execute
	response, err := handlers.HandleCreate(context.Background(), event)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.False(t, response.Success)
	assert.NotNil(t, response.Error)
	assert.Equal(t, "CREATE_FAILED", response.Error.Code)
	assert.Equal(t, "Failed to create unit", response.Error.Message)

	// Verify mock expectations
	mockRepo.AssertExpectations(t)
}

func TestUnitHandlers_HandleRead_Success(t *testing.T) {
	mockRepo := &repository.MockUnitRepository{}
	handlers := NewUnitHandlers(mockRepo)

	unit := &models.Unit{
		ID:           "test-unit-id",
		AccountID:    "test-account-123",
		SuggestedVin: "1HGBH41JXMN109186",
		Make:         "Honda",
		Model:        "Civic",
	}

	input := appsync.GetUnitInput{
		ID:        "test-unit-id",
		AccountID: "test-account-123",
	}
	argsJSON, err := json.Marshal(input)
	require.NoError(t, err)

	event := &appsync.AppSyncEvent{
		TypeName:  "Query",
		FieldName: "getUnit",
		Arguments: argsJSON,
	}

	// Mock expectations
	mockRepo.On("GetByID", mock.Anything, "test-account-123", "test-unit-id").Return(unit, nil)

	// Execute
	response, err := handlers.HandleRead(context.Background(), event)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.True(t, response.Success)
	assert.Equal(t, "Unit retrieved successfully", response.Message)
	assert.Equal(t, unit, response.Data)

	// Verify mock expectations
	mockRepo.AssertExpectations(t)
}

func TestUnitHandlers_HandleRead_NotFound(t *testing.T) {
	mockRepo := &repository.MockUnitRepository{}
	handlers := NewUnitHandlers(mockRepo)

	input := appsync.GetUnitInput{
		ID:        "test-unit-id",
		AccountID: "test-account-123",
	}
	argsJSON, err := json.Marshal(input)
	require.NoError(t, err)

	event := &appsync.AppSyncEvent{
		TypeName:  "Query",
		FieldName: "getUnit",
		Arguments: argsJSON,
	}

	// Mock expectations - unit not found
	mockRepo.On("GetByID", mock.Anything, "test-account-123", "test-unit-id").Return(nil, nil)

	// Execute
	response, err := handlers.HandleRead(context.Background(), event)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.False(t, response.Success)
	assert.NotNil(t, response.Error)
	assert.Equal(t, "NOT_FOUND", response.Error.Code)
	assert.Equal(t, "Unit not found", response.Error.Message)

	// Verify mock expectations
	mockRepo.AssertExpectations(t)
}

func TestUnitHandlers_HandleRead_ValidationError(t *testing.T) {
	mockRepo := &repository.MockUnitRepository{}
	handlers := NewUnitHandlers(mockRepo)

	tests := []struct {
		name          string
		input         appsync.GetUnitInput
		expectedError string
	}{
		{
			name: "missing id",
			input: appsync.GetUnitInput{
				ID:        "", // Missing
				AccountID: "test-account-123",
			},
			expectedError: "ID is required",
		},
		{
			name: "missing accountId",
			input: appsync.GetUnitInput{
				ID:        "test-unit-id",
				AccountID: "", // Missing
			},
			expectedError: "AccountID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			argsJSON, err := json.Marshal(tt.input)
			require.NoError(t, err)

			event := &appsync.AppSyncEvent{
				TypeName:  "Query",
				FieldName: "getUnit",
				Arguments: argsJSON,
			}

			// Execute
			response, err := handlers.HandleRead(context.Background(), event)

			// Assertions
			require.NoError(t, err)
			require.NotNil(t, response)
			assert.False(t, response.Success)
			assert.NotNil(t, response.Error)
			assert.Equal(t, "VALIDATION_ERROR", response.Error.Code)
			assert.Contains(t, response.Error.Message, tt.expectedError)

			// No repository calls should be made
			mockRepo.AssertNotCalled(t, "GetByID")
		})
	}
}

func TestUnitHandlers_HandleUpdate_Success(t *testing.T) {
	mockRepo := &repository.MockUnitRepository{}
	handlers := NewUnitHandlers(mockRepo)

	unit := models.Unit{
		SuggestedVin: "1HGBH41JXMN109186",
		Make:         "Honda",
		Model:        "Accord", // Updated model
	}

	input := appsync.UpdateUnitInput{
		ID:        "test-unit-id",
		AccountID: "test-account-123",
		Unit:      unit,
	}
	argsJSON, err := json.Marshal(input)
	require.NoError(t, err)

	event := &appsync.AppSyncEvent{
		TypeName:  "Mutation",
		FieldName: "updateUnit",
		Arguments: argsJSON,
	}

	// Existing unit to be returned by GetByKey
	existingUnit := &models.Unit{
		ID:           "test-unit-id",
		AccountID:    "test-account-123",
		SuggestedVin: "OLD_VIN_123",
		Make:         "Honda",
		Model:        "Civic",
	}

	// Mock expectations
	expectedUnit := input.Unit
	expectedUnit.ID = input.ID
	expectedUnit.AccountID = input.AccountID
	mockRepo.On("GetByID", mock.Anything, "test-account-123", "test-unit-id").Return(existingUnit, nil)
	mockRepo.On("Update", mock.Anything, &expectedUnit).Return(nil)

	// Execute
	response, err := handlers.HandleUpdate(context.Background(), event)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.True(t, response.Success)
	assert.Equal(t, "Unit updated successfully", response.Message)

	// Verify mock expectations
	mockRepo.AssertExpectations(t)
}

func TestUnitHandlers_HandleUpdate_ValidationError(t *testing.T) {
	mockRepo := &repository.MockUnitRepository{}
	handlers := NewUnitHandlers(mockRepo)

	tests := []struct {
		name          string
		input         appsync.UpdateUnitInput
		expectedError string
	}{
		{
			name: "missing id",
			input: appsync.UpdateUnitInput{
				ID:        "", // Missing
				AccountID: "test-account-123",
				Unit: models.Unit{
					SuggestedVin: "1HGBH41JXMN109186",
					Make:         "Honda",
					Model:        "Civic",
				},
			},
			expectedError: "ID is required",
		},
		{
			name: "missing accountId",
			input: appsync.UpdateUnitInput{
				ID:        "test-unit-id",
				AccountID: "", // Missing
				Unit: models.Unit{
					SuggestedVin: "1HGBH41JXMN109186",
					Make:         "Honda",
					Model:        "Civic",
				},
			},
			expectedError: "AccountID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			argsJSON, err := json.Marshal(tt.input)
			require.NoError(t, err)

			event := &appsync.AppSyncEvent{
				TypeName:  "Mutation",
				FieldName: "updateUnit",
				Arguments: argsJSON,
			}

			// Execute
			response, err := handlers.HandleUpdate(context.Background(), event)

			// Assertions
			require.NoError(t, err)
			require.NotNil(t, response)
			assert.False(t, response.Success)
			assert.NotNil(t, response.Error)
			assert.Equal(t, "VALIDATION_ERROR", response.Error.Code)
			assert.Contains(t, response.Error.Message, tt.expectedError)

			// No repository calls should be made for validation errors
			mockRepo.AssertNotCalled(t, "GetByID")
			mockRepo.AssertNotCalled(t, "Update")
		})
	}
}

func TestUnitHandlers_HandleUpdate_NotFound(t *testing.T) {
	mockRepo := &repository.MockUnitRepository{}
	handlers := NewUnitHandlers(mockRepo)

	unit := models.Unit{
		SuggestedVin: "1HGBH41JXMN109186",
		Make:         "Honda",
		Model:        "Accord",
	}

	input := appsync.UpdateUnitInput{
		ID:        "non-existent-id",
		AccountID: "test-account-123",
		Unit:      unit,
	}
	argsJSON, err := json.Marshal(input)
	require.NoError(t, err)

	event := &appsync.AppSyncEvent{
		TypeName:  "Mutation",
		FieldName: "updateUnit",
		Arguments: argsJSON,
	}

	// Mock expectations - unit not found
	mockRepo.On("GetByID", mock.Anything, "test-account-123", "non-existent-id").Return(nil, nil)

	// Execute
	response, err := handlers.HandleUpdate(context.Background(), event)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.False(t, response.Success)
	assert.NotNil(t, response.Error)
	assert.Equal(t, "NOT_FOUND", response.Error.Code)
	assert.Equal(t, "Unit not found", response.Error.Message)

	// GetByID should be called, but Update should not
	mockRepo.AssertNotCalled(t, "Update")
	mockRepo.AssertExpectations(t)
}

func TestUnitHandlers_HandleUpdate_ExistenceCheckError(t *testing.T) {
	mockRepo := &repository.MockUnitRepository{}
	handlers := NewUnitHandlers(mockRepo)

	unit := models.Unit{
		SuggestedVin: "1HGBH41JXMN109186",
		Make:         "Honda",
		Model:        "Accord",
	}

	input := appsync.UpdateUnitInput{
		ID:        "test-unit-id",
		AccountID: "test-account-123",
		Unit:      unit,
	}
	argsJSON, err := json.Marshal(input)
	require.NoError(t, err)

	event := &appsync.AppSyncEvent{
		TypeName:  "Mutation",
		FieldName: "updateUnit",
		Arguments: argsJSON,
	}

	// Mock expectations - GetByID returns error
	mockRepo.On("GetByID", mock.Anything, "test-account-123", "test-unit-id").Return(nil, errors.New("database connection failed"))

	// Execute
	response, err := handlers.HandleUpdate(context.Background(), event)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.False(t, response.Success)
	assert.NotNil(t, response.Error)
	assert.Equal(t, "UPDATE_FAILED", response.Error.Code)
	assert.Equal(t, "Failed to verify unit existence", response.Error.Message)

	// GetByID should be called, but Update should not
	mockRepo.AssertNotCalled(t, "Update")
	mockRepo.AssertExpectations(t)
}

func TestUnitHandlers_HandleUpdate_RepositoryError(t *testing.T) {
	mockRepo := &repository.MockUnitRepository{}
	handlers := NewUnitHandlers(mockRepo)

	unit := models.Unit{
		SuggestedVin: "1HGBH41JXMN109186",
		Make:         "Honda",
		Model:        "Accord",
	}

	input := appsync.UpdateUnitInput{
		ID:        "test-unit-id",
		AccountID: "test-account-123",
		Unit:      unit,
	}
	argsJSON, err := json.Marshal(input)
	require.NoError(t, err)

	event := &appsync.AppSyncEvent{
		TypeName:  "Mutation",
		FieldName: "updateUnit",
		Arguments: argsJSON,
	}

	// Existing unit to be returned by GetByKey
	existingUnit := &models.Unit{
		ID:           "test-unit-id",
		AccountID:    "test-account-123",
		SuggestedVin: "OLD_VIN_123",
		Make:         "Honda",
		Model:        "Civic",
	}

	// Mock expectations
	expectedUnit := input.Unit
	expectedUnit.ID = input.ID
	expectedUnit.AccountID = input.AccountID
	mockRepo.On("GetByID", mock.Anything, "test-account-123", "test-unit-id").Return(existingUnit, nil)
	mockRepo.On("Update", mock.Anything, &expectedUnit).Return(errors.New("database update failed"))

	// Execute
	response, err := handlers.HandleUpdate(context.Background(), event)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.False(t, response.Success)
	assert.NotNil(t, response.Error)
	assert.Equal(t, "UPDATE_FAILED", response.Error.Code)
	assert.Equal(t, "Failed to update unit", response.Error.Message)

	// Verify mock expectations
	mockRepo.AssertExpectations(t)
}

func TestUnitHandlers_HandleDelete_Success(t *testing.T) {
	mockRepo := &repository.MockUnitRepository{}
	handlers := NewUnitHandlers(mockRepo)

	unit := &models.Unit{
		ID:           "test-unit-id",
		AccountID:    "test-account-123",
		LocationID:   "test-location-id",
		SuggestedVin: "1HGBH41JXMN109186",
		Make:         "Honda",
		Model:        "Civic",
	}

	input := appsync.DeleteUnitInput{
		ID:        "test-unit-id",
		AccountID: "test-account-123",
	}
	argsJSON, err := json.Marshal(input)
	require.NoError(t, err)

	event := &appsync.AppSyncEvent{
		TypeName:  "Mutation",
		FieldName: "deleteUnit",
		Arguments: argsJSON,
	}

	// Mock expectations - first GetByID to find the unit, then Delete with locationID
	mockRepo.On("GetByID", mock.Anything, "test-account-123", "test-unit-id").Return(unit, nil)
	mockRepo.On("Delete", mock.Anything, "test-account-123", "test-location-id").Return(nil)

	// Execute
	response, err := handlers.HandleDelete(context.Background(), event)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.True(t, response.Success)
	assert.Equal(t, "Unit deleted successfully", response.Message)

	// Check response data
	data, ok := response.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "test-unit-id", data["id"])
	assert.Equal(t, "test-account-123", data["accountId"])
	assert.Equal(t, true, data["deleted"])

	// Verify mock expectations
	mockRepo.AssertExpectations(t)
}

func TestUnitHandlers_HandleDelete_ValidationError(t *testing.T) {
	mockRepo := &repository.MockUnitRepository{}
	handlers := NewUnitHandlers(mockRepo)

	tests := []struct {
		name          string
		input         appsync.DeleteUnitInput
		expectedError string
	}{
		{
			name: "missing id",
			input: appsync.DeleteUnitInput{
				ID:        "", // Missing
				AccountID: "test-account-123",
			},
			expectedError: "ID is required",
		},
		{
			name: "missing accountId",
			input: appsync.DeleteUnitInput{
				ID:        "test-unit-id",
				AccountID: "", // Missing
			},
			expectedError: "AccountID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			argsJSON, err := json.Marshal(tt.input)
			require.NoError(t, err)

			event := &appsync.AppSyncEvent{
				TypeName:  "Mutation",
				FieldName: "deleteUnit",
				Arguments: argsJSON,
			}

			// Execute
			response, err := handlers.HandleDelete(context.Background(), event)

			// Assertions
			require.NoError(t, err)
			require.NotNil(t, response)
			assert.False(t, response.Success)
			assert.NotNil(t, response.Error)
			assert.Equal(t, "VALIDATION_ERROR", response.Error.Code)
			assert.Contains(t, response.Error.Message, tt.expectedError)

			// No repository calls should be made
			mockRepo.AssertNotCalled(t, "Delete")
		})
	}
}

func TestUnitHandlers_HandleDelete_RepositoryError(t *testing.T) {
	mockRepo := &repository.MockUnitRepository{}
	handlers := NewUnitHandlers(mockRepo)

	input := appsync.DeleteUnitInput{
		ID:        "test-unit-id",
		AccountID: "test-account-123",
	}
	argsJSON, err := json.Marshal(input)
	require.NoError(t, err)

	event := &appsync.AppSyncEvent{
		TypeName:  "Mutation",
		FieldName: "deleteUnit",
		Arguments: argsJSON,
	}

	unit := &models.Unit{
		ID:           "test-unit-id",
		AccountID:    "test-account-123",
		LocationID:   "test-location-id",
		SuggestedVin: "1HGBH41JXMN109186",
		Make:         "Honda",
		Model:        "Civic",
	}

	// Mock expectations - GetByID succeeds but Delete returns error
	mockRepo.On("GetByID", mock.Anything, "test-account-123", "test-unit-id").Return(unit, nil)
	mockRepo.On("Delete", mock.Anything, "test-account-123", "test-location-id").Return(errors.New("database connection failed"))

	// Execute
	response, err := handlers.HandleDelete(context.Background(), event)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.False(t, response.Success)
	assert.NotNil(t, response.Error)
	assert.Equal(t, "DELETE_FAILED", response.Error.Code)
	assert.Equal(t, "Failed to delete unit", response.Error.Message)

	// Verify mock expectations
	mockRepo.AssertExpectations(t)
}

func TestUnitHandlers_HandleDelete_UnitNotFound(t *testing.T) {
	mockRepo := &repository.MockUnitRepository{}
	handlers := NewUnitHandlers(mockRepo)

	input := appsync.DeleteUnitInput{
		ID:        "non-existent-id",
		AccountID: "test-account-123",
	}
	argsJSON, err := json.Marshal(input)
	require.NoError(t, err)

	event := &appsync.AppSyncEvent{
		TypeName:  "Mutation",
		FieldName: "deleteUnit",
		Arguments: argsJSON,
	}

	// Mock expectations - GetByID returns nil (unit not found)
	mockRepo.On("GetByID", mock.Anything, "test-account-123", "non-existent-id").Return(nil, nil)

	// Execute
	response, err := handlers.HandleDelete(context.Background(), event)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.False(t, response.Success)
	assert.NotNil(t, response.Error)
	assert.Equal(t, "NOT_FOUND", response.Error.Code)
	assert.Equal(t, "Unit not found", response.Error.Message)

	// Delete should not be called if unit is not found
	mockRepo.AssertNotCalled(t, "Delete")
	mockRepo.AssertExpectations(t)
}

func TestUnitHandlers_HandleDelete_GetByIDError(t *testing.T) {
	mockRepo := &repository.MockUnitRepository{}
	handlers := NewUnitHandlers(mockRepo)

	input := appsync.DeleteUnitInput{
		ID:        "test-unit-id",
		AccountID: "test-account-123",
	}
	argsJSON, err := json.Marshal(input)
	require.NoError(t, err)

	event := &appsync.AppSyncEvent{
		TypeName:  "Mutation",
		FieldName: "deleteUnit",
		Arguments: argsJSON,
	}

	// Mock expectations - GetByID returns error
	mockRepo.On("GetByID", mock.Anything, "test-account-123", "test-unit-id").Return(nil, errors.New("database connection failed"))

	// Execute
	response, err := handlers.HandleDelete(context.Background(), event)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.False(t, response.Success)
	assert.NotNil(t, response.Error)
	assert.Equal(t, "DELETE_FAILED", response.Error.Code)
	assert.Equal(t, "Failed to find unit for deletion", response.Error.Message)

	// Delete should not be called if GetByID fails
	mockRepo.AssertNotCalled(t, "Delete")
	mockRepo.AssertExpectations(t)
}

func TestUnitHandlers_HandleList_Success(t *testing.T) {
	mockRepo := &repository.MockUnitRepository{}
	handlers := NewUnitHandlers(mockRepo)

	units := []models.Unit{
		{
			ID:           "unit-1",
			AccountID:    "test-account-123",
			SuggestedVin: "VIN1",
		},
		{
			ID:           "unit-2",
			AccountID:    "test-account-123",
			SuggestedVin: "VIN2",
		},
	}

	listResponse := &appsync.ListUnitsResponse{
		Items: units,
		Count: len(units),
	}

	limit := 25
	input := appsync.ListUnitsInput{
		AccountID: "test-account-123",
		Limit:     &limit,
	}
	argsJSON, err := json.Marshal(input)
	require.NoError(t, err)

	event := &appsync.AppSyncEvent{
		TypeName:  "Query",
		FieldName: "listUnits",
		Arguments: argsJSON,
	}

	// Mock expectations
	mockRepo.On("List", mock.Anything, &input).Return(listResponse, nil)

	// Execute
	response, err := handlers.HandleList(context.Background(), event)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.True(t, response.Success)
	assert.Contains(t, response.Message, "Retrieved 2 units")
	assert.Equal(t, listResponse, response.Data)

	// Verify mock expectations
	mockRepo.AssertExpectations(t)
}

func TestUnitHandlers_HandleList_ValidationError(t *testing.T) {
	mockRepo := &repository.MockUnitRepository{}
	handlers := NewUnitHandlers(mockRepo)

	// Input with missing required AccountID
	input := appsync.ListUnitsInput{
		AccountID: "", // Missing required field
	}
	argsJSON, err := json.Marshal(input)
	require.NoError(t, err)

	event := &appsync.AppSyncEvent{
		TypeName:  "Query",
		FieldName: "listUnits",
		Arguments: argsJSON,
	}

	// Execute
	response, err := handlers.HandleList(context.Background(), event)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.False(t, response.Success)
	assert.NotNil(t, response.Error)
	assert.Equal(t, "VALIDATION_ERROR", response.Error.Code)
	assert.Contains(t, response.Error.Message, "AccountID is required")

	// No repository calls should be made
	mockRepo.AssertNotCalled(t, "List")
}

func TestUnitHandlers_DumpEvent(t *testing.T) {
	mockRepo := &repository.MockUnitRepository{}
	handlers := NewUnitHandlers(mockRepo)

	unit := models.Unit{
		SuggestedVin: "1HGBH41JXMN109186",
	}

	input := appsync.CreateUnitInput{
		AccountID: "test-account-123",
		Unit:      unit,
	}
	argsJSON, err := json.Marshal(input)
	require.NoError(t, err)

	event := &appsync.AppSyncEvent{
		TypeName:  "Mutation",
		FieldName: "createUnit",
		Arguments: argsJSON,
		Identity: appsync.Identity{
			Sub:       "test-sub",
			Username:  "testuser",
			AccountID: "test-account-123",
			SourceIP:  []string{"192.168.1.1"},
		},
		Request: appsync.RequestHeaders{
			Headers: map[string]string{
				"content-type": "application/json",
			},
		},
		Info: appsync.Info{
			FieldName:      "createUnit",
			ParentTypeName: "Mutation",
			Variables:      map[string]interface{}{"test": "value"},
		},
	}

	// This test mainly ensures DumpEvent doesn't panic
	// In a real scenario, you might want to capture log output
	require.NotPanics(t, func() {
		handlers.DumpEvent(context.Background(), event)
	})
}

func TestNewUnitHandlers(t *testing.T) {
	mockRepo := &repository.MockUnitRepository{}
	handlers := NewUnitHandlers(mockRepo)

	assert.NotNil(t, handlers)
	assert.Equal(t, mockRepo, handlers.repo)
}
