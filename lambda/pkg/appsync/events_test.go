package appsync

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppSyncEvent_GetOperationType(t *testing.T) {
	tests := []struct {
		name      string
		fieldName string
		want      OperationType
	}{
		{
			name:      "Create operation",
			fieldName: "createUnit",
			want:      OperationTypeCreate,
		},
		{
			name:      "Read operation",
			fieldName: "getUnit",
			want:      OperationTypeRead,
		},
		{
			name:      "Update operation",
			fieldName: "updateUnit",
			want:      OperationTypeUpdate,
		},
		{
			name:      "Delete operation",
			fieldName: "deleteUnit",
			want:      OperationTypeDelete,
		},
		{
			name:      "List operation",
			fieldName: "listUnits",
			want:      OperationTypeList,
		},
		{
			name:      "Unknown operation defaults to read",
			fieldName: "unknownOperation",
			want:      OperationTypeRead,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := &AppSyncEvent{FieldName: tt.fieldName}
			assert.Equal(t, tt.want, event.GetOperationType())
		})
	}
}

func TestAppSyncEvent_ParseArguments_Create(t *testing.T) {
	input := CreateUnitInput{
		AccountID: "account-123",
		UnitType:  "commercialVehicleType",
		Data: map[string]interface{}{
			"suggestedVin": "1HGBH41JXMN109186",
			"make":         "Honda",
			"model":        "Civic",
		},
	}
	argsJSON, err := json.Marshal(input)
	require.NoError(t, err)

	event := &AppSyncEvent{
		FieldName: "createUnit",
		Arguments: argsJSON,
	}

	result, err := event.ParseArguments()
	require.NoError(t, err)

	parsedInput, ok := result.(CreateUnitInput)
	require.True(t, ok)

	assert.Equal(t, input.AccountID, parsedInput.AccountID)
	assert.Equal(t, input.UnitType, parsedInput.UnitType)
	assert.Equal(t, input.Data["suggestedVin"], parsedInput.Data["suggestedVin"])
	assert.Equal(t, input.Data["make"], parsedInput.Data["make"])
	assert.Equal(t, input.Data["model"], parsedInput.Data["model"])
}

func TestAppSyncEvent_ParseArguments_Read(t *testing.T) {
	input := GetUnitInput{
		ID:        "123e4567-e89b-12d3-a456-426614174000",
		AccountID: "account-123",
		UnitType:  "commercialVehicleType",
	}

	argsJSON, err := json.Marshal(input)
	require.NoError(t, err)

	event := &AppSyncEvent{
		FieldName: "getUnit",
		Arguments: argsJSON,
	}

	result, err := event.ParseArguments()
	require.NoError(t, err)

	parsedInput, ok := result.(GetUnitInput)
	require.True(t, ok)

	assert.Equal(t, input.ID, parsedInput.ID)
	assert.Equal(t, input.AccountID, parsedInput.AccountID)
	assert.Equal(t, input.UnitType, parsedInput.UnitType)
}

func TestAppSyncEvent_ParseArguments_Update(t *testing.T) {
	input := UpdateUnitInput{
		ID:        "123e4567-e89b-12d3-a456-426614174000",
		AccountID: "account-123",
		UnitType:  "commercialVehicleType",
		Data: map[string]interface{}{
			"suggestedVin": "1HGBH41JXMN109186",
			"make":         "Honda",
			"model":        "Accord", // Updated model
		},
	}

	argsJSON, err := json.Marshal(input)
	require.NoError(t, err)

	event := &AppSyncEvent{
		FieldName: "updateUnit",
		Arguments: argsJSON,
	}

	result, err := event.ParseArguments()
	require.NoError(t, err)

	parsedInput, ok := result.(UpdateUnitInput)
	require.True(t, ok)

	assert.Equal(t, input.ID, parsedInput.ID)
	assert.Equal(t, input.AccountID, parsedInput.AccountID)
	assert.Equal(t, input.UnitType, parsedInput.UnitType)
	assert.Equal(t, input.Data["model"], parsedInput.Data["model"])
}

func TestAppSyncEvent_ParseArguments_Delete(t *testing.T) {
	input := DeleteUnitInput{
		ID:        "123e4567-e89b-12d3-a456-426614174000",
		AccountID: "account-123",
		UnitType:  "commercialVehicleType",
	}

	argsJSON, err := json.Marshal(input)
	require.NoError(t, err)

	event := &AppSyncEvent{
		FieldName: "deleteUnit",
		Arguments: argsJSON,
	}

	result, err := event.ParseArguments()
	require.NoError(t, err)

	parsedInput, ok := result.(DeleteUnitInput)
	require.True(t, ok)

	assert.Equal(t, input.ID, parsedInput.ID)
	assert.Equal(t, input.AccountID, parsedInput.AccountID)
	assert.Equal(t, input.UnitType, parsedInput.UnitType)
}

func TestAppSyncEvent_ParseArguments_List(t *testing.T) {
	limit := 25
	filter := "active"

	input := ListUnitsInput{
		AccountID: "account-123",
		Limit:     &limit,
		Filter:    &filter,
	}

	argsJSON, err := json.Marshal(input)
	require.NoError(t, err)

	event := &AppSyncEvent{
		FieldName: "listUnits",
		Arguments: argsJSON,
	}

	result, err := event.ParseArguments()
	require.NoError(t, err)

	parsedInput, ok := result.(ListUnitsInput)
	require.True(t, ok)

	assert.Equal(t, input.AccountID, parsedInput.AccountID)
	require.NotNil(t, parsedInput.Limit)
	assert.Equal(t, *input.Limit, *parsedInput.Limit)
	require.NotNil(t, parsedInput.Filter)
	assert.Equal(t, *input.Filter, *parsedInput.Filter)
}

func TestAppSyncEvent_ParseArguments_InvalidJSON(t *testing.T) {
	event := &AppSyncEvent{
		FieldName: "createUnit",
		Arguments: json.RawMessage(`{"invalid": json}`),
	}

	result, err := event.ParseArguments()
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestNewSuccessResponse(t *testing.T) {
	data := map[string]string{"test": "data"}
	message := "Operation successful"

	response := NewSuccessResponse(data, message)

	assert.True(t, response.Success)
	assert.Equal(t, data, response.Data)
	assert.Equal(t, message, response.Message)
	assert.Nil(t, response.Error)
}

func TestNewErrorResponse(t *testing.T) {
	code := "TEST_ERROR"
	message := "Test error message"
	details := "Detailed error information"

	response := NewErrorResponse(code, message, details)

	assert.False(t, response.Success)
	assert.Nil(t, response.Data)
	assert.Empty(t, response.Message)
	require.NotNil(t, response.Error)
	assert.Equal(t, code, response.Error.Code)
	assert.Equal(t, message, response.Error.Message)
	assert.Equal(t, details, response.Error.Details)
}

func TestIdentity(t *testing.T) {
	identity := Identity{
		Sub:       "test-sub-123",
		Username:  "testuser",
		AccountID: "test-account-123",
		SourceIP:  []string{"192.168.1.1"},
		Groups:    []string{"admin", "user"},
	}

	assert.Equal(t, "test-sub-123", identity.Sub)
	assert.Equal(t, "testuser", identity.Username)
	assert.Equal(t, "test-account-123", identity.AccountID)
	assert.Contains(t, identity.SourceIP, "192.168.1.1")
	assert.Contains(t, identity.Groups, "admin")
	assert.Contains(t, identity.Groups, "user")
}

func TestListUnitsResponse(t *testing.T) {
	units := []map[string]interface{}{
		{"id": "unit-1", "suggestedVin": "VIN1"},
		{"id": "unit-2", "suggestedVin": "VIN2"},
	}
	nextToken := "token123"

	response := ListUnitsResponse{
		Items:     units,
		NextToken: &nextToken,
		Count:     len(units),
	}

	assert.Len(t, response.Items, 2)
	assert.Equal(t, 2, response.Count)
	require.NotNil(t, response.NextToken)
	assert.Equal(t, nextToken, *response.NextToken)
}

func TestOperationTypeConstants(t *testing.T) {
	assert.Equal(t, OperationType("CREATE"), OperationTypeCreate)
	assert.Equal(t, OperationType("READ"), OperationTypeRead)
	assert.Equal(t, OperationType("UPDATE"), OperationTypeUpdate)
	assert.Equal(t, OperationType("DELETE"), OperationTypeDelete)
	assert.Equal(t, OperationType("LIST"), OperationTypeList)
}
