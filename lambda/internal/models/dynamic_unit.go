package models

import (
	_ "embed"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"github.com/xeipuuv/gojsonschema"
)

//go:embed commercialVehicleType.json
var commercialVehicleTypeSchema []byte

// schemaCache caches compiled JSON schemas
var schemaCache = make(map[string]*gojsonschema.Schema)
var schemaCacheMutex sync.RWMutex

// DynamicUnit represents a unit with dynamic structure based on schema
type DynamicUnit struct {
	// Core fields present in all units
	ID        string `json:"id" dynamodbav:"id"`         // Unit UUID
	AccountID string `json:"accountId" dynamodbav:"pk"`  // Primary Key
	UnitType  string `json:"unitType" dynamodbav:"sk"`   // Sort Key (unitType#id)
	
	// Timestamp fields
	CreatedAt int64 `json:"createdAt" dynamodbav:"createdAt"`
	UpdatedAt int64 `json:"updatedAt" dynamodbav:"updatedAt"`
	DeletedAt int64 `json:"deletedAt" dynamodbav:"deletedAt"`
	
	// Dynamic data based on schema
	Data map[string]interface{} `json:"data" dynamodbav:"data"`
	
	// Internal fields for processing
	schema *gojsonschema.Schema `json:"-" dynamodbav:"-"`
}

// NewDynamicUnit creates a new dynamic unit with the specified unit type
func NewDynamicUnit(unitType string) (*DynamicUnit, error) {
	schema, err := loadSchema(unitType)
	if err != nil {
		return nil, fmt.Errorf("failed to load schema for unit type %s: %w", unitType, err)
	}
	
	return &DynamicUnit{
		UnitType: unitType,
		Data:     make(map[string]interface{}),
		schema:   schema,
	}, nil
}

// ValidateAndSetData validates the provided data against the schema and sets it
func (du *DynamicUnit) ValidateAndSetData(data map[string]interface{}) error {
	if du.schema == nil {
		schema, err := loadSchema(du.UnitType)
		if err != nil {
			return fmt.Errorf("failed to load schema for validation: %w", err)
		}
		du.schema = schema
	}
	
	// Ensure accountId is present as it's required in all schemas
	if du.AccountID != "" {
		data["accountId"] = du.AccountID
	}
	
	// Inject ID if not present
	if du.ID != "" {
		data["id"] = du.ID
	}
	
	// Validate against schema
	documentLoader := gojsonschema.NewGoLoader(data)
	result, err := du.schema.Validate(documentLoader)
	if err != nil {
		return fmt.Errorf("schema validation error: %w", err)
	}
	
	if !result.Valid() {
		var errorMessages []string
		for _, err := range result.Errors() {
			errorMessages = append(errorMessages, err.String())
		}
		return fmt.Errorf("data validation failed: %s", strings.Join(errorMessages, "; "))
	}
	
	du.Data = data
	return nil
}

// loadSchema loads and caches a JSON schema from the embedded config
func loadSchema(unitType string) (*gojsonschema.Schema, error) {
	schemaCacheMutex.RLock()
	schema, exists := schemaCache[unitType]
	schemaCacheMutex.RUnlock()
	
	if exists {
		return schema, nil
	}
	
	// For now, we only support commercialVehicleType
	if unitType != "commercialVehicleType" {
		return nil, fmt.Errorf("unsupported unit type: %s", unitType)
	}
	
	// Load schema from embedded data
	schemaData := commercialVehicleTypeSchema
	
	// Parse and compile schema
	schemaLoader := gojsonschema.NewBytesLoader(schemaData)
	compiledSchema, err := gojsonschema.NewSchema(schemaLoader)
	if err != nil {
		return nil, fmt.Errorf("failed to compile schema: %w", err)
	}
	
	// Cache the compiled schema
	schemaCacheMutex.Lock()
	schemaCache[unitType] = compiledSchema
	schemaCacheMutex.Unlock()
	
	return compiledSchema, nil
}

// GetAvailableUnitTypes returns a list of available unit types from embedded config
func GetAvailableUnitTypes() ([]string, error) {
	// For now, we only support commercialVehicleType
	// In the future, we can expand this by embedding multiple files or using a directory pattern
	return []string{"commercialVehicleType"}, nil
}

// GenerateID generates a new UUID for the unit
func (du *DynamicUnit) GenerateID() {
	du.ID = uuid.New().String()
}

// SetTimestamps sets created and updated timestamps
func (du *DynamicUnit) SetTimestamps() {
	now := time.Now().Unix()
	if du.CreatedAt == 0 {
		du.CreatedAt = now
	}
	du.UpdatedAt = now
}

// IsDeleted checks if the unit is soft deleted
func (du *DynamicUnit) IsDeleted() bool {
	return du.DeletedAt > 0
}

// MarkDeleted marks the unit as soft deleted
func (du *DynamicUnit) MarkDeleted() {
	du.DeletedAt = time.Now().Unix()
}

// GetSortKey returns the composite sort key (unitType#id)
func (du *DynamicUnit) GetSortKey() string {
	return fmt.Sprintf("%s#%s", du.UnitType, du.ID)
}

// GetKey returns the composite primary key for DynamoDB operations (PK + SK)
func (du *DynamicUnit) GetKey() map[string]types.AttributeValue {
	key, _ := attributevalue.MarshalMap(map[string]interface{}{
		"pk": du.AccountID,      // Primary Key (AccountID)
		"sk": du.GetSortKey(),   // Sort Key (unitType#id)
	})
	return key
}

// ToMap converts the dynamic unit to a map for DynamoDB storage
func (du *DynamicUnit) ToMap() (map[string]interface{}, error) {
	result := map[string]interface{}{
		"id":        du.ID,
		"pk":        du.AccountID,
		"sk":        du.GetSortKey(),
		"unitType":  du.UnitType,
		"createdAt": du.CreatedAt,
		"updatedAt": du.UpdatedAt,
		"deletedAt": du.DeletedAt,
	}
	
	// Merge dynamic data
	for key, value := range du.Data {
		// Skip core fields that we manage separately
		if key != "id" && key != "accountId" && key != "unitType" && 
		   key != "createdAt" && key != "updatedAt" && key != "deletedAt" {
			result[key] = value
		}
	}
	
	return result, nil
}

// FromMap populates the dynamic unit from a map (from DynamoDB)
func (du *DynamicUnit) FromMap(data map[string]interface{}) error {
	// Extract core fields
	if id, ok := data["id"].(string); ok {
		du.ID = id
	}
	if pk, ok := data["pk"].(string); ok {
		du.AccountID = pk
	}
	if sk, ok := data["sk"].(string); ok {
		// Extract unitType from SK (format: unitType#id)
		parts := strings.SplitN(sk, "#", 2)
		if len(parts) == 2 {
			du.UnitType = parts[0]
		}
	}
	if unitType, ok := data["unitType"].(string); ok {
		du.UnitType = unitType
	}
	if createdAt, ok := data["createdAt"].(int64); ok {
		du.CreatedAt = createdAt
	}
	if updatedAt, ok := data["updatedAt"].(int64); ok {
		du.UpdatedAt = updatedAt
	}
	if deletedAt, ok := data["deletedAt"].(int64); ok {
		du.DeletedAt = deletedAt
	}
	
	// Initialize data map and copy remaining fields
	du.Data = make(map[string]interface{})
	for key, value := range data {
		if key != "pk" && key != "sk" && key != "createdAt" && key != "updatedAt" && key != "deletedAt" {
			du.Data[key] = value
		}
	}
	
	return nil
}

// Validate validates the current data against the schema
func (du *DynamicUnit) Validate() error {
	if du.schema == nil {
		schema, err := loadSchema(du.UnitType)
		if err != nil {
			return fmt.Errorf("failed to load schema for validation: %w", err)
		}
		du.schema = schema
	}
	
	// Prepare data for validation
	validationData := make(map[string]interface{})
	for k, v := range du.Data {
		validationData[k] = v
	}
	
	// Ensure core fields are included
	validationData["id"] = du.ID
	validationData["accountId"] = du.AccountID
	validationData["unitType"] = du.UnitType
	validationData["createdAt"] = du.CreatedAt
	validationData["updatedAt"] = du.UpdatedAt
	validationData["deletedAt"] = du.DeletedAt
	
	documentLoader := gojsonschema.NewGoLoader(validationData)
	result, err := du.schema.Validate(documentLoader)
	if err != nil {
		return fmt.Errorf("schema validation error: %w", err)
	}
	
	if !result.Valid() {
		var errorMessages []string
		for _, err := range result.Errors() {
			errorMessages = append(errorMessages, err.String())
		}
		return fmt.Errorf("data validation failed: %s", strings.Join(errorMessages, "; "))
	}
	
	return nil
}