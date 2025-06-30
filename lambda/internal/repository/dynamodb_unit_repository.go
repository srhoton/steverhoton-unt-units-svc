package repository

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

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
	for k, value := range lastKey {
		switch v := value.(type) {
		case *types.AttributeValueMemberS:
			tokenData[k] = v.Value
		case *types.AttributeValueMemberN:
			tokenData[k] = v.Value
		default:
			// For other types, convert to string representation
			tokenData[k] = fmt.Sprintf("%v", value)
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
	for k, value := range tokenData {
		switch v := value.(type) {
		case string:
			// Try to determine if it's a number or string
			if k == "pk" || k == "sk" {
				// PK is accountID (string), SK is locationID (string)
				lastKey[k] = &types.AttributeValueMemberS{Value: v}
			} else {
				// For other fields, try to parse as number first
				if _, err := strconv.ParseFloat(v, 64); err == nil {
					lastKey[k] = &types.AttributeValueMemberN{Value: v}
				} else {
					lastKey[k] = &types.AttributeValueMemberS{Value: v}
				}
			}
		case float64:
			// JSON numbers become float64
			lastKey[k] = &types.AttributeValueMemberN{Value: fmt.Sprintf("%.0f", v)}
		default:
			// Fallback to string representation
			lastKey[k] = &types.AttributeValueMemberS{Value: fmt.Sprintf("%v", v)}
		}
	}

	return lastKey, nil
}

// Create creates a new unit in DynamoDB
func (r *DynamoDBUnitRepository) Create(ctx context.Context, unit *models.DynamicUnit) error {
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

	// Validate the unit data against schema
	if err := unit.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Convert to map for DynamoDB storage
	unitMap, err := unit.ToMap()
	if err != nil {
		return fmt.Errorf("failed to convert unit to map: %w", err)
	}

	// Marshal the unit to DynamoDB attribute map
	item, err := attributevalue.MarshalMap(unitMap)
	if err != nil {
		return fmt.Errorf("failed to marshal unit: %w", err)
	}

	// Create the item with condition that the key doesn't already exist
	input := &dynamodb.PutItemInput{
		TableName:           aws.String(r.tableName),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(pk) AND attribute_not_exists(sk)"),
	}

	_, err = r.client.PutItem(ctx, input)
	if err != nil {
		var conditionalCheckFailedException *types.ConditionalCheckFailedException
		if errors.As(err, &conditionalCheckFailedException) {
			return fmt.Errorf("unit already exists for account %s and unitType %s with ID %s", unit.AccountID, unit.UnitType, unit.ID)
		}
		return fmt.Errorf("failed to create unit: %w", err)
	}

	return nil
}

// GetByKey retrieves a unit by its composite primary key (accountID + sortKey)
func (r *DynamoDBUnitRepository) GetByKey(ctx context.Context, accountID, sortKey string) (*models.DynamicUnit, error) {
	if accountID == "" {
		return nil, errors.New("accountID is required")
	}
	if sortKey == "" {
		return nil, errors.New("sortKey is required")
	}

	key := map[string]types.AttributeValue{
		"pk": &types.AttributeValueMemberS{Value: accountID},
		"sk": &types.AttributeValueMemberS{Value: sortKey},
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

	// Convert DynamoDB item to map
	var itemMap map[string]interface{}
	err = attributevalue.UnmarshalMap(result.Item, &itemMap)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal unit: %w", err)
	}

	// Extract unitType from the sort key to create the dynamic unit
	parts := strings.SplitN(sortKey, "#", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid sort key format: %s", sortKey)
	}
	unitType := parts[0]

	unit, err := models.NewDynamicUnit(unitType)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic unit: %w", err)
	}

	// Populate unit from the data
	err = unit.FromMap(itemMap)
	if err != nil {
		return nil, fmt.Errorf("failed to populate unit from map: %w", err)
	}

	// Don't return soft deleted units
	if unit.IsDeleted() {
		return nil, nil
	}

	return unit, nil
}

// GetByID retrieves a unit by its unique ID within an account and unitType
func (r *DynamoDBUnitRepository) GetByID(ctx context.Context, accountID, unitType, unitID string) (*models.DynamicUnit, error) {
	if accountID == "" {
		return nil, errors.New("accountID is required")
	}
	if unitType == "" {
		return nil, errors.New("unitType is required")
	}
	if unitID == "" {
		return nil, errors.New("unitID is required")
	}

	// Build the sort key
	sortKey := fmt.Sprintf("%s#%s", unitType, unitID)
	
	// Use GetByKey since we have the full composite key
	return r.GetByKey(ctx, accountID, sortKey)
}

// Update updates an existing unit in DynamoDB
func (r *DynamoDBUnitRepository) Update(ctx context.Context, unit *models.DynamicUnit) error {
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
	if unit.ID == "" {
		return errors.New("ID is required")
	}

	// Update timestamp
	unit.SetTimestamps()

	// Validate the unit data against schema
	if err := unit.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Convert to map for DynamoDB storage
	unitMap, err := unit.ToMap()
	if err != nil {
		return fmt.Errorf("failed to convert unit to map: %w", err)
	}

	// Marshal the unit to DynamoDB attribute map
	item, err := attributevalue.MarshalMap(unitMap)
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
			return fmt.Errorf("unit does not exist or is deleted for account %s, unitType %s, and ID %s", unit.AccountID, unit.UnitType, unit.ID)
		}
		return fmt.Errorf("failed to update unit: %w", err)
	}

	return nil
}

// Delete soft deletes a unit by setting deletedAt timestamp
func (r *DynamoDBUnitRepository) Delete(ctx context.Context, accountID, unitType, unitID string) error {
	if accountID == "" {
		return errors.New("accountID is required")
	}
	if unitType == "" {
		return errors.New("unitType is required")
	}
	if unitID == "" {
		return errors.New("unitID is required")
	}

	// Build the sort key
	sortKey := fmt.Sprintf("%s#%s", unitType, unitID)

	// First, check if the unit exists and get current data
	unit, err := r.GetByKey(ctx, accountID, sortKey)
	if err != nil {
		return fmt.Errorf("failed to check unit existence: %w", err)
	}
	if unit == nil {
		return fmt.Errorf("unit not found for account %s, unitType %s, and ID %s", accountID, unitType, unitID)
	}

	// Mark as deleted and update timestamp
	unit.MarkDeleted()
	unit.SetTimestamps()

	// Convert to map for DynamoDB storage
	unitMap, err := unit.ToMap()
	if err != nil {
		return fmt.Errorf("failed to convert unit to map: %w", err)
	}

	// Update the existing item with deletedAt set
	item, err := attributevalue.MarshalMap(unitMap)
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

	// Build the query input to get units for the account
	var keyConditionExpression string
	expressionAttributeValues := map[string]types.AttributeValue{
		":accountId": &types.AttributeValueMemberS{Value: input.AccountID},
		":zero":      &types.AttributeValueMemberN{Value: "0"},
	}

	if input.UnitType != nil && *input.UnitType != "" {
		// Filter by specific unit type
		keyConditionExpression = "pk = :accountId AND begins_with(sk, :unitType)"
		expressionAttributeValues[":unitType"] = &types.AttributeValueMemberS{Value: *input.UnitType + "#"}
	} else {
		// Get all units for the account
		keyConditionExpression = "pk = :accountId"
	}

	queryInput := &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		KeyConditionExpression: aws.String(keyConditionExpression),
		FilterExpression:       aws.String("attribute_not_exists(deletedAt) OR deletedAt = :zero"),
		ExpressionAttributeValues: expressionAttributeValues,
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

	// Convert results to map array
	var unitMaps []map[string]interface{}
	for _, item := range result.Items {
		var unitMap map[string]interface{}
		err = attributevalue.UnmarshalMap(item, &unitMap)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal unit: %w", err)
		}
		unitMaps = append(unitMaps, unitMap)
	}

	// Build response
	response := &appsync.ListUnitsResponse{
		Items: unitMaps,
		Count: len(unitMaps),
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
func (r *DynamoDBUnitRepository) Exists(ctx context.Context, accountID, sortKey string) (bool, error) {
	if accountID == "" {
		return false, errors.New("accountID is required")
	}
	if sortKey == "" {
		return false, errors.New("sortKey is required")
	}

	// Use GetByKey which handles the logic
	unit, err := r.GetByKey(ctx, accountID, sortKey)
	if err != nil {
		return false, fmt.Errorf("failed to check unit existence: %w", err)
	}

	return unit != nil, nil
}
