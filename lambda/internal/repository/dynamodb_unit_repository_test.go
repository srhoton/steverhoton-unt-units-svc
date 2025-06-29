package repository

import (
	"encoding/base64"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/steverhoton/unt-units-svc/internal/models"
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
				"sk": &types.AttributeValueMemberS{Value: "location-123"},
			},
			expected: true,
		},
		{
			name: "mixed attributes with id",
			lastKey: map[string]types.AttributeValue{
				"pk":        &types.AttributeValueMemberS{Value: "account-456"},
				"sk":        &types.AttributeValueMemberS{Value: "location-123"},
				"id":        &types.AttributeValueMemberS{Value: "test-id-123"},
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
		"sk":        &types.AttributeValueMemberS{Value: "550e8400-e29b-41d4-a716-446655440000"},
		"id":        &types.AttributeValueMemberS{Value: "unit-123"},
		"createdAt": &types.AttributeValueMemberN{Value: "1640995200"},
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
	assert.Contains(t, decodedKey, "id")
	assert.Contains(t, decodedKey, "createdAt")

	// Verify string values are correctly handled
	pkAttr, ok := decodedKey["pk"].(*types.AttributeValueMemberS)
	require.True(t, ok)
	assert.Equal(t, "account-123", pkAttr.Value)

	// Verify SK is correctly handled as string
	skAttr, ok := decodedKey["sk"].(*types.AttributeValueMemberS)
	require.True(t, ok)
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", skAttr.Value)

	idAttr, ok := decodedKey["id"].(*types.AttributeValueMemberS)
	require.True(t, ok)
	assert.Equal(t, "unit-123", idAttr.Value)

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

func TestUnit_SetTimestamps(t *testing.T) {
	// Test the Unit model's timestamp functionality
	unit := &models.Unit{
		ID:           "test-unit-id",
		AccountID:    "test-account-123",
		SuggestedVin: "1HGBH41JXMN109186",
	}

	// Initially, timestamps should be zero
	assert.Equal(t, int64(0), unit.CreatedAt)
	assert.Equal(t, int64(0), unit.UpdatedAt)

	// Set timestamps for the first time
	unit.SetTimestamps()

	// Both timestamps should be set
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
