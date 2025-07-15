package repository

import (
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/steverhoton/unt-units-svc/internal/models"
	"github.com/steverhoton/unt-units-svc/pkg/appsync"
)

func TestDynamoDBUnitRepository_PaginationTokenEncoding(t *testing.T) {
	repo := &DynamoDBUnitRepository{}

	tests := []struct {
		name     string
		lastKey  map[string]types.AttributeValue
		expected bool // true if should successfully encode/decode
	}{
		{
			name:     "nil last key",
			lastKey:  nil,
			expected: true,
		},
		{
			name: "string attributes",
			lastKey: map[string]types.AttributeValue{
				"pk": &types.AttributeValueMemberS{Value: "account-456"},
				"sk": &types.AttributeValueMemberS{Value: "test-id-123#commercialVehicleType"},
			},
			expected: true,
		},
		{
			name: "mixed string and number attributes",
			lastKey: map[string]types.AttributeValue{
				"pk":        &types.AttributeValueMemberS{Value: "account-456"},
				"sk":        &types.AttributeValueMemberS{Value: "test-id-123#commercialVehicleType"},
				"createdAt": &types.AttributeValueMemberN{Value: "1609459200"},
			},
			expected: true,
		},
		{
			name: "empty string values",
			lastKey: map[string]types.AttributeValue{
				"pk": &types.AttributeValueMemberS{Value: ""},
				"sk": &types.AttributeValueMemberS{Value: ""},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test encoding
			token, err := repo.encodePaginationToken(tt.lastKey)
			if tt.expected {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				return
			}

			// For nil input, token should be empty
			if tt.lastKey == nil {
				assert.Empty(t, token)
				return
			}

			// Token should not be empty for valid input
			assert.NotEmpty(t, token)

			// Test decoding
			decodedKey, err := repo.decodePaginationToken(token)
			require.NoError(t, err)

			// Compare original and decoded keys
			require.Equal(t, len(tt.lastKey), len(decodedKey))
			for key := range tt.lastKey {
				assert.Contains(t, decodedKey, key)
			}
		})
	}
}

func TestDynamoDBUnitRepository_PaginationTokenDecoding(t *testing.T) {
	repo := &DynamoDBUnitRepository{}

	tests := []struct {
		name        string
		token       string
		expectError bool
	}{
		{
			name:        "empty token",
			token:       "",
			expectError: false,
		},
		{
			name:        "invalid base64",
			token:       "invalid-base64!@#$",
			expectError: true,
		},
		{
			name:        "valid base64 but invalid JSON",
			token:       "aW52YWxpZCBqc29u", // "invalid json" in base64
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decodedKey, err := repo.decodePaginationToken(tt.token)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, decodedKey)
			} else {
				assert.NoError(t, err)
				if tt.token == "" {
					assert.Nil(t, decodedKey)
				}
			}
		})
	}
}

func TestDynamoDBUnitRepository_PaginationTokenRoundTrip(t *testing.T) {
	repo := &DynamoDBUnitRepository{}

	// Create a comprehensive last key that represents what DynamoDB might return
	originalKey := map[string]types.AttributeValue{
		"pk":        &types.AttributeValueMemberS{Value: "account-123"},
		"sk":        &types.AttributeValueMemberS{Value: "550e8400-e29b-41d4-a716-446655440000#commercialVehicleType"},
		"createdAt": &types.AttributeValueMemberN{Value: "1640995200"},
		"updatedAt": &types.AttributeValueMemberN{Value: "1640995200"},
	}

	// Encode to token
	token, err := repo.encodePaginationToken(originalKey)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// Decode back to key
	decodedKey, err := repo.decodePaginationToken(token)
	require.NoError(t, err)
	require.NotNil(t, decodedKey)

	// Verify all keys are present
	require.Equal(t, len(originalKey), len(decodedKey))

	// Verify specific key-value pairs
	assert.Contains(t, decodedKey, "pk")
	assert.Contains(t, decodedKey, "sk")
	assert.Contains(t, decodedKey, "createdAt")
	assert.Contains(t, decodedKey, "updatedAt")

	// Verify string values are correctly handled
	pkAttr, ok := decodedKey["pk"].(*types.AttributeValueMemberS)
	require.True(t, ok)
	assert.Equal(t, "account-123", pkAttr.Value)

	skAttr, ok := decodedKey["sk"].(*types.AttributeValueMemberS)
	require.True(t, ok)
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000#commercialVehicleType", skAttr.Value)

	// Verify that the token is base64 encoded (robust check)
	_, err = base64.StdEncoding.DecodeString(token)
	require.NoError(t, err, "Token is not valid base64")
	assert.NotEmpty(t, token, "Decoded token should not be empty")
}

func TestUnit_SoftDeleteFunctionality(t *testing.T) {
	// Test the Unit model's soft delete functionality
	unit := &models.Unit{
		ID:           "test-unit-id",
		AccountID:    "test-account-123",
		UnitType:     "commercialVehicleType",
		SuggestedVin: "1HGBH41JXMN109186",
		Make:         "Honda",
		Model:        "Civic",
	}

	// Initially, unit should not be deleted
	assert.False(t, unit.IsDeleted())
	assert.Equal(t, int64(0), unit.DeletedAt)

	// Mark unit as deleted
	unit.MarkDeleted()

	// After marking as deleted, unit should be considered deleted
	assert.True(t, unit.IsDeleted())
	assert.Greater(t, unit.DeletedAt, int64(0))

	// Verify the deletedAt timestamp is in epoch time (reasonable range)
	now := time.Now().Unix()
	assert.LessOrEqual(t, unit.DeletedAt, now)
	assert.Greater(t, unit.DeletedAt, now-10) // Within last 10 seconds
}

func TestUnit_TimestampFunctionality(t *testing.T) {
	// Test the Unit model's timestamp functionality
	unit := &models.Unit{
		ID:           "test-unit-id",
		AccountID:    "test-account-123",
		SuggestedVin: "1HGBH41JXMN109186",
	}

	// Initially, timestamps should be zero
	assert.Equal(t, int64(0), unit.CreatedAt)
	assert.Equal(t, int64(0), unit.UpdatedAt)

	// Set timestamps
	unit.SetTimestamps()

	// After setting timestamps, both should be populated
	now := time.Now().Unix()
	assert.Greater(t, unit.CreatedAt, int64(0))
	assert.Greater(t, unit.UpdatedAt, int64(0))
	assert.LessOrEqual(t, unit.CreatedAt, now)
	assert.LessOrEqual(t, unit.UpdatedAt, now)

	// Store the original created time
	originalCreatedAt := unit.CreatedAt

	// Sleep a bit and set timestamps again
	time.Sleep(time.Second * 1)
	unit.SetTimestamps()

	// CreatedAt should remain the same, UpdatedAt should be newer
	assert.Equal(t, originalCreatedAt, unit.CreatedAt)
	assert.Greater(t, unit.UpdatedAt, originalCreatedAt)
}

func TestListUnitsResponse_EmptyItemsArray(t *testing.T) {
	// This test verifies that the ListUnitsResponse returns an empty array
	// instead of null when no units are found, which is critical for GraphQL
	// schema compliance where items is defined as non-nullable [Unit!]!

	// Test case 1: Response with no units should have empty array, not null
	response := &appsync.ListUnitsResponse{
		Items: []models.Unit{}, // Empty array
		Count: 0,
	}

	// Marshal to JSON to verify it serializes as [] not null
	jsonBytes, err := json.Marshal(response)
	require.NoError(t, err)

	// Parse as generic map to check the exact JSON structure
	var jsonMap map[string]interface{}
	err = json.Unmarshal(jsonBytes, &jsonMap)
	require.NoError(t, err)

	// Verify items field exists and is an array, not null
	items, exists := jsonMap["items"]
	assert.True(t, exists, "items field should exist")
	assert.NotNil(t, items, "items field should not be null")

	// Verify it's an array (slice in Go)
	itemsArray, ok := items.([]interface{})
	assert.True(t, ok, "items should be an array")
	assert.Equal(t, 0, len(itemsArray), "items array should be empty")

	// Test case 2: Response with nil slice should NOT happen (this is the bug)
	responseWithNilSlice := &appsync.ListUnitsResponse{
		Items: nil, // This is what causes the bug
		Count: 0,
	}

	// Marshal to JSON
	jsonBytes, err = json.Marshal(responseWithNilSlice)
	require.NoError(t, err)

	// Parse as generic map
	err = json.Unmarshal(jsonBytes, &jsonMap)
	require.NoError(t, err)

	// Verify items field is null (this is the problematic behavior)
	items, exists = jsonMap["items"]
	assert.True(t, exists, "items field should exist")
	assert.Nil(t, items, "items field is null when slice is nil - this is the bug!")
}

func TestDynamoDBUnitRepository_ListEmptySliceInitialization(t *testing.T) {
	// This test verifies that when we declare a slice variable,
	// it needs to be initialized as an empty slice, not nil

	// Test case 1: var declaration creates nil slice
	var nilSlice []models.Unit
	assert.Nil(t, nilSlice, "var declaration should create nil slice")

	// When marshaled to JSON, nil slice becomes null
	jsonBytes, err := json.Marshal(nilSlice)
	require.NoError(t, err)
	assert.Equal(t, "null", string(jsonBytes), "nil slice marshals to null")

	// Test case 2: make() creates empty slice
	emptySlice := make([]models.Unit, 0)
	assert.NotNil(t, emptySlice, "make() should create non-nil slice")
	assert.Equal(t, 0, len(emptySlice), "slice should be empty")

	// When marshaled to JSON, empty slice becomes []
	jsonBytes, err = json.Marshal(emptySlice)
	require.NoError(t, err)
	assert.Equal(t, "[]", string(jsonBytes), "empty slice marshals to []")

	// Test case 3: slice literal creates empty slice
	literalSlice := []models.Unit{}
	assert.NotNil(t, literalSlice, "slice literal should create non-nil slice")
	assert.Equal(t, 0, len(literalSlice), "slice should be empty")

	// When marshaled to JSON, empty slice becomes []
	jsonBytes, err = json.Marshal(literalSlice)
	require.NoError(t, err)
	assert.Equal(t, "[]", string(jsonBytes), "empty slice literal marshals to []")
}

func TestUnit_GetKey(t *testing.T) {
	// Test the Unit model's GetKey() method with new PK/SK structure
	unit := &models.Unit{
		ID:        "550e8400-e29b-41d4-a716-446655440000",
		AccountID: "test-account-123",
		UnitType:  "commercialVehicleType",
	}

	key := unit.GetKey()
	require.NotNil(t, key)

	// Verify PK is AccountID
	pkAttr, ok := key["pk"].(*types.AttributeValueMemberS)
	require.True(t, ok, "pk should be a string attribute")
	assert.Equal(t, "test-account-123", pkAttr.Value)

	// Verify SK is formatted as {unitId}#{unitType}
	skAttr, ok := key["sk"].(*types.AttributeValueMemberS)
	require.True(t, ok, "sk should be a string attribute")
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000#commercialVehicleType", skAttr.Value)
}

func TestUnit_GetSortKey(t *testing.T) {
	// Test the Unit model's GetSortKey() method
	unit := &models.Unit{
		ID:       "test-unit-id-123",
		UnitType: "commercialVehicleType",
	}

	sk := unit.GetSortKey()
	assert.Equal(t, "test-unit-id-123#commercialVehicleType", sk)

	// Test with different unit type
	unit.UnitType = "someOtherType"
	sk = unit.GetSortKey()
	assert.Equal(t, "test-unit-id-123#someOtherType", sk)
}
