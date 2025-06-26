package models

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnit_GetKey(t *testing.T) {
	tests := []struct {
		name string
		unit Unit
		want map[string]string
	}{
		{
			name: "Valid unit with required fields",
			unit: Unit{
				ID:        "123e4567-e89b-12d3-a456-426614174000",
				AccountID: "account-123",
			},
			want: map[string]string{
				"pk": "123e4567-e89b-12d3-a456-426614174000",
				"sk": "account-123",
			},
		},
		{
			name: "Unit with empty fields",
			unit: Unit{
				ID:        "",
				AccountID: "",
			},
			want: map[string]string{
				"pk": "",
				"sk": "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := tt.unit.GetKey()

			// Verify the key contains the expected fields
			require.Contains(t, key, "pk")
			require.Contains(t, key, "sk")

			// Extract string values from AttributeValue
			pkAttr, ok := key["pk"].(*types.AttributeValueMemberS)
			require.True(t, ok)
			assert.Equal(t, tt.want["pk"], pkAttr.Value)

			skAttr, ok := key["sk"].(*types.AttributeValueMemberS)
			require.True(t, ok)
			assert.Equal(t, tt.want["sk"], skAttr.Value)
		})
	}
}

func TestUnit_GenerateID(t *testing.T) {
	unit := Unit{}
	
	// Initially no ID
	assert.Empty(t, unit.ID)
	
	// Generate ID
	unit.GenerateID()
	
	// Should have a UUID
	assert.NotEmpty(t, unit.ID)
	assert.Len(t, unit.ID, 36) // Standard UUID length
	
	// Generate another ID for a different unit
	unit2 := Unit{}
	unit2.GenerateID()
	
	// Should be different
	assert.NotEqual(t, unit.ID, unit2.ID)
}

func TestUnit_SetTimestamps(t *testing.T) {
	tests := []struct {
		name   string
		unit   Unit
		setup  func(*Unit)
		verify func(*testing.T, *Unit, int64)
	}{
		{
			name: "New unit with zero timestamps",
			unit: Unit{},
			setup: func(u *Unit) {
				// No setup needed - timestamps are zero
			},
			verify: func(t *testing.T, u *Unit, beforeTime int64) {
				assert.GreaterOrEqual(t, u.CreatedAt, beforeTime)
				assert.GreaterOrEqual(t, u.UpdatedAt, beforeTime)
				assert.Equal(t, u.CreatedAt, u.UpdatedAt)
			},
		},
		{
			name: "Existing unit with CreatedAt set",
			unit: Unit{},
			setup: func(u *Unit) {
				u.CreatedAt = time.Now().Unix() - 3600 // 1 hour ago
			},
			verify: func(t *testing.T, u *Unit, beforeTime int64) {
				assert.Less(t, u.CreatedAt, beforeTime)           // CreatedAt should not change
				assert.GreaterOrEqual(t, u.UpdatedAt, beforeTime) // UpdatedAt should be recent
				assert.Greater(t, u.UpdatedAt, u.CreatedAt)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(&tt.unit)
			beforeTime := time.Now().Unix()

			tt.unit.SetTimestamps()

			tt.verify(t, &tt.unit, beforeTime)
		})
	}
}

func TestUnit_IsDeleted(t *testing.T) {
	tests := []struct {
		name      string
		deletedAt int64
		want      bool
	}{
		{
			name:      "Unit not deleted (zero timestamp)",
			deletedAt: 0,
			want:      false,
		},
		{
			name:      "Unit deleted (positive timestamp)",
			deletedAt: time.Now().Unix(),
			want:      true,
		},
		{
			name:      "Unit deleted (past timestamp)",
			deletedAt: time.Now().Unix() - 3600,
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			unit := Unit{DeletedAt: tt.deletedAt}
			assert.Equal(t, tt.want, unit.IsDeleted())
		})
	}
}

func TestUnit_MarkDeleted(t *testing.T) {
	unit := Unit{DeletedAt: 0}
	beforeTime := time.Now().Unix()

	unit.MarkDeleted()

	assert.GreaterOrEqual(t, unit.DeletedAt, beforeTime)
	assert.True(t, unit.IsDeleted())
}

func TestExtendedAttribute(t *testing.T) {
	attr := ExtendedAttribute{
		AttributeName:  "testAttribute",
		AttributeValue: "testValue",
	}

	assert.Equal(t, "testAttribute", attr.AttributeName)
	assert.Equal(t, "testValue", attr.AttributeValue)
}

func TestAcesAttribute(t *testing.T) {
	attr := AcesAttribute{
		AttributeName:  "testAttribute",
		AttributeValue: "testValue",
		AttributeKey:   "testKey",
	}

	assert.Equal(t, "testAttribute", attr.AttributeName)
	assert.Equal(t, "testValue", attr.AttributeValue)
	assert.Equal(t, "testKey", attr.AttributeKey)
}

func TestUnit_FullStructValidation(t *testing.T) {
	unit := Unit{
		ID:                           "123e4567-e89b-12d3-a456-426614174000",
		AccountID:                    "account-123",
		SuggestedVin:                 "1HGBH41JXMN109186",
		ErrorCode:                    "E001",
		PossibleValues:               "Value1,Value2",
		ErrorText:                    "Test error",
		VehicleDescriptor:            "Test Vehicle",
		Make:                         "Honda",
		ManufacturerName:             "Honda Motor Co.",
		Model:                        "Civic",
		ModelYear:                    "2023",
		Series:                       "LX",
		VehicleType:                  "Passenger Car",
		PlantCity:                    "Marysville",
		PlantCountry:                 "USA",
		Note:                         "Test note",
		BodyClass:                    "Sedan",
		Doors:                        "4",
		GrossVehicleWeightRatingFrom: "3000",
		GrossVehicleWeightRatingTo:   "3500",
		WheelBaseInchesFrom:          "106.3",
		BedType:                      "N/A",
		CabType:                      "N/A",
		TrailerTypeConnection:        "N/A",
		TrailerBodyType:              "N/A",
		CustomMotorcycleType:         "N/A",
		MotorcycleSuspensionType:     "N/A",
		MotorcycleChassisType:        "N/A",
		BusFloorConfigurationType:    "N/A",
		BusType:                      "N/A",
		EngineNumberOfCylinders:      "4",
		DisplacementCc:               "1998",
		DisplacementCi:               "122",
		DisplacementL:                "2.0",
		FuelTypePrimary:              "Gasoline",
		EngineBrakeHpFrom:            "158",
		SeatBeltType:                 "3-Point",
		OtherRestraintSystemInfo:     "Standard",
		FrontAirBagLocations:         "Driver,Passenger",
		ExtendedAttributes: []ExtendedAttribute{
			{
				AttributeName:  "Color",
				AttributeValue: "Blue",
			},
		},
		AcesAttributes: []AcesAttribute{
			{
				AttributeName:  "Engine",
				AttributeValue: "2.0L I4",
				AttributeKey:   "ENG001",
			},
		},
	}

	// Test that all required fields are set
	assert.NotEmpty(t, unit.ID)
	assert.NotEmpty(t, unit.AccountID)
	assert.NotEmpty(t, unit.SuggestedVin)
	assert.NotEmpty(t, unit.ErrorCode)
	assert.NotEmpty(t, unit.Make)
	assert.NotEmpty(t, unit.Model)

	// Test arrays
	assert.Len(t, unit.ExtendedAttributes, 1)
	assert.Len(t, unit.AcesAttributes, 1)

	// Test GetKey works with full unit
	key := unit.GetKey()
	assert.Contains(t, key, "pk")
	assert.Contains(t, key, "sk")

	// Test timestamps
	unit.SetTimestamps()
	assert.Greater(t, unit.CreatedAt, int64(0))
	assert.Greater(t, unit.UpdatedAt, int64(0))

	// Test deletion
	assert.False(t, unit.IsDeleted())
	unit.MarkDeleted()
	assert.True(t, unit.IsDeleted())
}
