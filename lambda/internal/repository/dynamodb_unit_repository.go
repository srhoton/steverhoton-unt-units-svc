package repository

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/steverhoton/unt-units-svc/internal/models"
	"github.com/steverhoton/unt-units-svc/pkg/appsync"
)

// DynamoDBUnitRepository implements UnitRepository using DynamoDB
type DynamoDBUnitRepository struct {
	client    *dynamodb.Client
	tableName string
}

// NewDynamoDBUnitRepository creates a new DynamoDB unit repository
func NewDynamoDBUnitRepository(client *dynamodb.Client, tableName string) *DynamoDBUnitRepository {
	return &DynamoDBUnitRepository{
		client:    client,
		tableName: tableName,
	}
}

// encodePaginationToken encodes the LastEvaluatedKey into a base64 token
func (r *DynamoDBUnitRepository) encodePaginationToken(lastKey map[string]types.AttributeValue) (string, error) {
	if lastKey == nil {
		return "", nil
	}

	// Convert the LastEvaluatedKey to a simple map[string]interface{}
	tokenData := make(map[string]interface{})
	for key, value := range lastKey {
		switch v := value.(type) {
		case *types.AttributeValueMemberS:
			tokenData[key] = v.Value
		case *types.AttributeValueMemberN:
			tokenData[key] = v.Value
		default:
			// For other types, convert to string representation
			tokenData[key] = fmt.Sprintf("%v", value)
		}
	}

	// Marshal to JSON
	jsonBytes, err := json.Marshal(tokenData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal token data: %w", err)
	}

	// Encode to base64
	return base64.StdEncoding.EncodeToString(jsonBytes), nil
}

// decodePaginationToken decodes a base64 token back to LastEvaluatedKey format
func (r *DynamoDBUnitRepository) decodePaginationToken(token string) (map[string]types.AttributeValue, error) {
	if token == "" {
		return nil, nil
	}

	// Decode from base64
	jsonBytes, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return nil, fmt.Errorf("failed to decode token: %w", err)
	}

	// Unmarshal from JSON
	var tokenData map[string]interface{}
	err = json.Unmarshal(jsonBytes, &tokenData)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal token data: %w", err)
	}

	// Convert back to DynamoDB AttributeValue format
	lastKey := make(map[string]types.AttributeValue)
	for key, value := range tokenData {
		switch v := value.(type) {
		case string:
			// Try to determine if it's a number or string
			if key == "pk" || key == "sk" {
				// These are typically strings
				lastKey[key] = &types.AttributeValueMemberS{Value: v}
			} else {
				// For other fields, try to parse as number first
				if _, err := strconv.ParseFloat(v, 64); err == nil {
					lastKey[key] = &types.AttributeValueMemberN{Value: v}
				} else {
					lastKey[key] = &types.AttributeValueMemberS{Value: v}
				}
			}
		case float64:
			// JSON numbers become float64
			lastKey[key] = &types.AttributeValueMemberN{Value: fmt.Sprintf("%.0f", v)}
		default:
			// Fallback to string representation
			lastKey[key] = &types.AttributeValueMemberS{Value: fmt.Sprintf("%v", v)}
		}
	}

	return lastKey, nil
}

// Create creates a new unit in DynamoDB
func (r *DynamoDBUnitRepository) Create(ctx context.Context, unit *models.Unit) error {
	if unit == nil {
		return errors.New("unit cannot be nil")
	}

	// Validate required fields
	if unit.AccountID == "" {
		return errors.New("accountID is required")
	}
	if unit.UnitType == "" {
		return errors.New("unitType is required")
	}

	// Generate UUID if not already set
	if unit.ID == "" {
		unit.GenerateID()
	}

	// Set timestamps
	unit.SetTimestamps()

	// Set the computed SK field for DynamoDB
	unit.SortKey = unit.GetSortKey()

	// Marshal the unit to DynamoDB attribute map
	item, err := attributevalue.MarshalMap(unit)
	if err != nil {
		return fmt.Errorf("failed to marshal unit: %w", err)
	}

	// Create the item with condition that it doesn't already exist
	input := &dynamodb.PutItemInput{
		TableName:           aws.String(r.tableName),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(pk) AND attribute_not_exists(sk)"),
	}

	_, err = r.client.PutItem(ctx, input)
	if err != nil {
		var conditionalCheckFailedException *types.ConditionalCheckFailedException
		if errors.As(err, &conditionalCheckFailedException) {
			return fmt.Errorf("unit with id %s and type %s already exists for account %s", unit.ID, unit.UnitType, unit.AccountID)
		}
		return fmt.Errorf("failed to create unit: %w", err)
	}

	return nil
}

// GetByKey retrieves a unit by its composite primary key
func (r *DynamoDBUnitRepository) GetByKey(ctx context.Context, accountID, unitID, unitType string) (*models.Unit, error) {
	if accountID == "" {
		return nil, errors.New("accountID is required")
	}
	if unitID == "" {
		return nil, errors.New("unitID is required")
	}
	if unitType == "" {
		return nil, errors.New("unitType is required")
	}

	// Construct the sort key
	sk := unitID + "#" + unitType

	key := map[string]types.AttributeValue{
		"pk": &types.AttributeValueMemberS{Value: accountID},
		"sk": &types.AttributeValueMemberS{Value: sk},
	}

	input := &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key:       key,
	}

	result, err := r.client.GetItem(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get unit: %w", err)
	}

	if result.Item == nil {
		return nil, nil // Unit not found
	}

	var unit models.Unit
	err = attributevalue.UnmarshalMap(result.Item, &unit)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal unit: %w", err)
	}

	// Don't return soft deleted units
	if unit.IsDeleted() {
		return nil, nil
	}

	return &unit, nil
}

// Update updates an existing unit in DynamoDB
func (r *DynamoDBUnitRepository) Update(ctx context.Context, unit *models.Unit) error {
	if unit == nil {
		return errors.New("unit cannot be nil")
	}

	// Validate required fields
	if unit.AccountID == "" {
		return errors.New("accountID is required")
	}
	if unit.ID == "" {
		return errors.New("unit ID is required")
	}
	if unit.UnitType == "" {
		return errors.New("unitType is required")
	}

	// Update timestamp
	unit.SetTimestamps()

	// Set the computed SK field for DynamoDB
	unit.SortKey = unit.GetSortKey()

	// Marshal the unit to DynamoDB attribute map
	item, err := attributevalue.MarshalMap(unit)
	if err != nil {
		return fmt.Errorf("failed to marshal unit: %w", err)
	}

	// Update the item with condition that it exists and is not deleted
	input := &dynamodb.PutItemInput{
		TableName:           aws.String(r.tableName),
		Item:                item,
		ConditionExpression: aws.String("attribute_exists(pk) AND attribute_exists(sk) AND (attribute_not_exists(deletedAt) OR deletedAt = :zero)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":zero": &types.AttributeValueMemberN{Value: "0"},
		},
	}

	_, err = r.client.PutItem(ctx, input)
	if err != nil {
		var conditionalCheckFailedException *types.ConditionalCheckFailedException
		if errors.As(err, &conditionalCheckFailedException) {
			return fmt.Errorf("unit with id %s and type %s does not exist or is deleted for account %s", unit.ID, unit.UnitType, unit.AccountID)
		}
		return fmt.Errorf("failed to update unit: %w", err)
	}

	return nil
}

// Delete soft deletes a unit by setting deletedAt timestamp
func (r *DynamoDBUnitRepository) Delete(ctx context.Context, accountID, unitID, unitType string) error {
	if accountID == "" {
		return errors.New("accountID is required")
	}
	if unitID == "" {
		return errors.New("unitID is required")
	}
	if unitType == "" {
		return errors.New("unitType is required")
	}

	// First, check if the unit exists and get current data
	unit, err := r.GetByKey(ctx, accountID, unitID, unitType)
	if err != nil {
		return fmt.Errorf("failed to check unit existence: %w", err)
	}
	if unit == nil {
		return fmt.Errorf("unit with id %s and type %s not found for account %s", unitID, unitType, accountID)
	}

	// Mark as deleted
	unit.MarkDeleted()

	// Set the computed SK field for DynamoDB
	unit.SortKey = unit.GetSortKey()

	// Update the item
	item, err := attributevalue.MarshalMap(unit)
	if err != nil {
		return fmt.Errorf("failed to marshal unit: %w", err)
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      item,
	}

	_, err = r.client.PutItem(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to delete unit: %w", err)
	}

	return nil
}

// List retrieves a paginated list of units
func (r *DynamoDBUnitRepository) List(ctx context.Context, input *appsync.ListUnitsInput) (*appsync.ListUnitsResponse, error) {
	if input == nil {
		return nil, errors.New("input is required")
	}
	if input.AccountID == "" {
		return nil, errors.New("accountID is required")
	}

	// Default limit
	limit := int32(20)
	if input.Limit != nil && *input.Limit > 0 && *input.Limit <= 100 {
		limit = int32(*input.Limit)
	}

	// Build the query input to get all units for the account
	// Now that AccountID is the PK, we can query directly on the main table
	queryInput := &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		KeyConditionExpression: aws.String("pk = :accountId"),
		FilterExpression:       aws.String("attribute_not_exists(deletedAt) OR deletedAt = :zero"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":accountId": &types.AttributeValueMemberS{Value: input.AccountID},
			":zero":      &types.AttributeValueMemberN{Value: "0"},
		},
		Limit: aws.Int32(limit),
	}

	// Handle pagination with proper token decoding
	if input.NextToken != nil && *input.NextToken != "" {
		exclusiveStartKey, err := r.decodePaginationToken(*input.NextToken)
		if err != nil {
			return nil, fmt.Errorf("failed to decode pagination token: %w", err)
		}
		if exclusiveStartKey != nil {
			queryInput.ExclusiveStartKey = exclusiveStartKey
		}
	}

	result, err := r.client.Query(ctx, queryInput)
	if err != nil {
		return nil, fmt.Errorf("failed to list units: %w", err)
	}

	// Unmarshal the results
	// Initialize as empty slice to ensure it marshals to [] instead of null
	// This is critical for GraphQL schema compliance where items: [Unit!]! is non-nullable
	units := make([]models.Unit, 0)
	err = attributevalue.UnmarshalListOfMaps(result.Items, &units)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal units: %w", err)
	}

	// Build response
	response := &appsync.ListUnitsResponse{
		Items: units,
		Count: len(units),
	}

	// Handle next token for pagination with proper encoding
	if result.LastEvaluatedKey != nil {
		nextToken, err := r.encodePaginationToken(result.LastEvaluatedKey)
		if err != nil {
			return nil, fmt.Errorf("failed to encode pagination token: %w", err)
		}
		if nextToken != "" {
			response.NextToken = &nextToken
		}
	}

	return response, nil
}

// Exists checks if a unit exists by its primary key
func (r *DynamoDBUnitRepository) Exists(ctx context.Context, accountID, unitID, unitType string) (bool, error) {
	if accountID == "" {
		return false, errors.New("accountID is required")
	}
	if unitID == "" {
		return false, errors.New("unitID is required")
	}
	if unitType == "" {
		return false, errors.New("unitType is required")
	}

	// Construct the sort key
	sk := unitID + "#" + unitType

	key := map[string]types.AttributeValue{
		"pk": &types.AttributeValueMemberS{Value: accountID},
		"sk": &types.AttributeValueMemberS{Value: sk},
	}

	input := &dynamodb.GetItemInput{
		TableName:            aws.String(r.tableName),
		Key:                  key,
		ProjectionExpression: aws.String("pk, sk, deletedAt"),
	}

	result, err := r.client.GetItem(ctx, input)
	if err != nil {
		return false, fmt.Errorf("failed to check unit existence: %w", err)
	}

	if result.Item == nil {
		return false, nil
	}

	// Check if the unit is not soft deleted
	if deletedAtValue, exists := result.Item["deletedAt"]; exists {
		if nVal, ok := deletedAtValue.(*types.AttributeValueMemberN); ok {
			deletedAt, err := strconv.ParseInt(nVal.Value, 10, 64)
			if err == nil && deletedAt > 0 {
				return false, nil // Unit is soft deleted
			}
		}
	}

	return true, nil
}

// GetByUnitID retrieves units by unit ID using the GSI
// This will return all units with the given ID across all accounts and types
func (r *DynamoDBUnitRepository) GetByUnitID(ctx context.Context, unitID string) ([]models.Unit, error) {
	if unitID == "" {
		return nil, errors.New("unitID is required")
	}

	// Query the GSI on unitId
	queryInput := &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		IndexName:              aws.String("unit-id-index"), // GSI on unitId
		KeyConditionExpression: aws.String("id = :unitId"),
		FilterExpression:       aws.String("attribute_not_exists(deletedAt) OR deletedAt = :zero"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":unitId": &types.AttributeValueMemberS{Value: unitID},
			":zero":   &types.AttributeValueMemberN{Value: "0"},
		},
	}

	result, err := r.client.Query(ctx, queryInput)
	if err != nil {
		return nil, fmt.Errorf("failed to query units by ID: %w", err)
	}

	// Initialize as empty slice to ensure it marshals to [] instead of null
	units := make([]models.Unit, 0)
	err = attributevalue.UnmarshalListOfMaps(result.Items, &units)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal units: %w", err)
	}

	return units, nil
}
