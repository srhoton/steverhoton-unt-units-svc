package handlers

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/steverhoton/unt-units-svc/internal/models"
	"github.com/steverhoton/unt-units-svc/internal/repository"
	"github.com/steverhoton/unt-units-svc/pkg/appsync"
)

func TestUnitHandlers_HandleCreate_Success_Dynamic(t *testing.T) {
	mockRepo := &repository.MockUnitRepository{}
	handlers := NewUnitHandlers(mockRepo)

	input := appsync.CreateUnitInput{
		AccountID: "test-account-123",
		UnitType:  "commercialVehicleType",
		Data: map[string]interface{}{
			// Only include data that is actually needed for the test
			// The accountId and id will be injected automatically by ValidateAndSetData
		},
	}
	argsJSON, err := json.Marshal(input)
	require.NoError(t, err)

	event := &appsync.AppSyncEvent{
		TypeName:  "Mutation",
		FieldName: "createUnit",
		Arguments: argsJSON,
	}

	// Mock expectations - expect any DynamicUnit to be created
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.DynamicUnit")).Return(nil)

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

func TestUnitHandlers_HandleCreate_ValidationError_Dynamic(t *testing.T) {
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
				UnitType:  "commercialVehicleType",
				Data: map[string]interface{}{
					// Empty data for validation tests
				},
			},
			expectedError: "AccountID is required",
		},
		{
			name: "missing unitType",
			input: appsync.CreateUnitInput{
				AccountID: "test-account-123",
				UnitType:  "", // Missing
				Data: map[string]interface{}{
					// Empty data for validation tests
				},
			},
			expectedError: "UnitType is required",
		},
		{
			name: "invalid unitType",
			input: appsync.CreateUnitInput{
				AccountID: "test-account-123",
				UnitType:  "invalidType",
				Data: map[string]interface{}{
					// Empty data for validation tests
				},
			},
			expectedError: "Invalid unit type",
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

func TestUnitHandlers_HandleRead_Success_Dynamic(t *testing.T) {
	mockRepo := &repository.MockUnitRepository{}
	handlers := NewUnitHandlers(mockRepo)

	// Create a dynamic unit for testing
	unit, err := models.NewDynamicUnit("commercialVehicleType")
	require.NoError(t, err)
	unitID := "550e8400-e29b-41d4-a716-446655440000" // Valid UUID
	unit.ID = unitID
	unit.AccountID = "test-account-123"
	unit.UnitType = "commercialVehicleType"
	unit.Data = map[string]interface{}{
		"id":        unitID,
		"accountId": "test-account-123",
	}

	input := appsync.GetUnitInput{
		ID:        unitID,
		AccountID: "test-account-123",
		UnitType:  "commercialVehicleType",
	}
	argsJSON, err := json.Marshal(input)
	require.NoError(t, err)

	event := &appsync.AppSyncEvent{
		TypeName:  "Query",
		FieldName: "getUnit",
		Arguments: argsJSON,
	}

	// Mock expectations
	mockRepo.On("GetByID", mock.Anything, "test-account-123", "commercialVehicleType", unitID).Return(unit, nil)

	// Execute
	response, err := handlers.HandleRead(context.Background(), event)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.True(t, response.Success)
	assert.Equal(t, "Unit retrieved successfully", response.Message)

	// Verify mock expectations
	mockRepo.AssertExpectations(t)
}

func TestUnitHandlers_HandleRead_ValidationError_Dynamic(t *testing.T) {
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
				UnitType:  "commercialVehicleType",
			},
			expectedError: "ID is required",
		},
		{
			name: "missing accountId",
			input: appsync.GetUnitInput{
				ID:        "test-unit-id",
				AccountID: "", // Missing
				UnitType:  "commercialVehicleType",
			},
			expectedError: "AccountID is required",
		},
		{
			name: "missing unitType",
			input: appsync.GetUnitInput{
				ID:        "test-unit-id",
				AccountID: "test-account-123",
				UnitType:  "", // Missing
			},
			expectedError: "UnitType is required",
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

func TestUnitHandlers_HandleUpdate_Success_Dynamic(t *testing.T) {
	mockRepo := &repository.MockUnitRepository{}
	handlers := NewUnitHandlers(mockRepo)

	unitID := "550e8400-e29b-41d4-a716-446655440001" // Valid UUID
	input := appsync.UpdateUnitInput{
		ID:        unitID,
		AccountID: "test-account-123",
		UnitType:  "commercialVehicleType",
		Data: map[string]interface{}{
			// Minimal data for update test
		},
	}
	argsJSON, err := json.Marshal(input)
	require.NoError(t, err)

	event := &appsync.AppSyncEvent{
		TypeName:  "Mutation",
		FieldName: "updateUnit",
		Arguments: argsJSON,
	}

	// Existing unit to be returned by GetByID
	existingUnit, err := models.NewDynamicUnit("commercialVehicleType")
	require.NoError(t, err)
	existingUnit.ID = unitID
	existingUnit.AccountID = "test-account-123"
	existingUnit.UnitType = "commercialVehicleType"

	// Mock expectations
	mockRepo.On("GetByID", mock.Anything, "test-account-123", "commercialVehicleType", unitID).Return(existingUnit, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.DynamicUnit")).Return(nil)

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

func TestUnitHandlers_HandleDelete_Success_Dynamic(t *testing.T) {
	mockRepo := &repository.MockUnitRepository{}
	handlers := NewUnitHandlers(mockRepo)

	unitID := "550e8400-e29b-41d4-a716-446655440002" // Valid UUID
	input := appsync.DeleteUnitInput{
		ID:        unitID,
		AccountID: "test-account-123",
		UnitType:  "commercialVehicleType",
	}
	argsJSON, err := json.Marshal(input)
	require.NoError(t, err)

	event := &appsync.AppSyncEvent{
		TypeName:  "Mutation",
		FieldName: "deleteUnit",
		Arguments: argsJSON,
	}

	// Mock expectations - only Delete is called, not GetByID
	mockRepo.On("Delete", mock.Anything, "test-account-123", "commercialVehicleType", unitID).Return(nil)

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
	assert.Equal(t, unitID, data["id"])
	assert.Equal(t, "test-account-123", data["accountId"])
	assert.Equal(t, true, data["deleted"])

	// Verify mock expectations
	mockRepo.AssertExpectations(t)
}

func TestUnitHandlers_HandleList_Success_Dynamic(t *testing.T) {
	mockRepo := &repository.MockUnitRepository{}
	handlers := NewUnitHandlers(mockRepo)

	// Create dynamic units for testing
	unit1, err := models.NewDynamicUnit("commercialVehicleType")
	require.NoError(t, err)
	unit1ID := "550e8400-e29b-41d4-a716-446655440003"
	unit1.ID = unit1ID
	unit1.AccountID = "test-account-123"
	unit1.Data = map[string]interface{}{
		"id":        unit1ID,
		"accountId": "test-account-123",
	}

	unit2, err := models.NewDynamicUnit("commercialVehicleType")
	require.NoError(t, err)
	unit2ID := "550e8400-e29b-41d4-a716-446655440004"
	unit2.ID = unit2ID
	unit2.AccountID = "test-account-123"
	unit2.Data = map[string]interface{}{
		"id":        unit2ID,
		"accountId": "test-account-123",
	}

	units := []*models.DynamicUnit{unit1, unit2}

	listResponse := &appsync.ListUnitsResponse{
		Items:     make([]map[string]interface{}, len(units)),
		Count:     len(units),
		NextToken: nil,
	}

	// Convert units to maps for response
	for i, unit := range units {
		unitMap, err := unit.ToMap()
		require.NoError(t, err)
		listResponse.Items[i] = unitMap
	}

	unitType := "commercialVehicleType"
	limit := 25
	input := appsync.ListUnitsInput{
		AccountID: "test-account-123",
		UnitType:  &unitType,
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

func TestUnitHandlers_DumpEvent_Dynamic(t *testing.T) {
	mockRepo := &repository.MockUnitRepository{}
	handlers := NewUnitHandlers(mockRepo)

	input := appsync.CreateUnitInput{
		AccountID: "test-account-123",
		UnitType:  "commercialVehicleType",
		Data: map[string]interface{}{
			// Minimal data for DumpEvent test
		},
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

func TestUnitHandlers_HandleCreate_SchemaValidationError_Dynamic(t *testing.T) {
	mockRepo := &repository.MockUnitRepository{}
	handlers := NewUnitHandlers(mockRepo)

	input := appsync.CreateUnitInput{
		AccountID: "test-account-123",
		UnitType:  "commercialVehicleType",
		Data: map[string]interface{}{
			"invalidField": "this field doesn't exist in schema and additionalProperties is false",
		},
	}
	argsJSON, err := json.Marshal(input)
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
	assert.Contains(t, response.Error.Message, "validation failed")

	// No repository calls should be made for validation errors
	mockRepo.AssertNotCalled(t, "Create")
}