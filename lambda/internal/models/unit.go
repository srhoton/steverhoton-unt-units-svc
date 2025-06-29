package models

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
)

// ExtendedAttribute represents a key-value pair for extended attributes
type ExtendedAttribute struct {
	AttributeName  string `json:"attributeName" dynamodbav:"attributeName"`
	AttributeValue string `json:"attributeValue" dynamodbav:"attributeValue"`
}

// AcesAttribute represents an ACES attribute with additional key field
type AcesAttribute struct {
	AttributeName  string `json:"attributeName" dynamodbav:"attributeName"`
	AttributeValue string `json:"attributeValue" dynamodbav:"attributeValue"`
	AttributeKey   string `json:"attributeKey" dynamodbav:"attributeKey"`
}

// Unit represents a commercial vehicle type unit in DynamoDB
type Unit struct {
	// DynamoDB key fields
	ID         string `json:"id" dynamodbav:"id"`         // Unit UUID
	AccountID  string `json:"accountId" dynamodbav:"pk"`  // Primary Key
	LocationID string `json:"locationId" dynamodbav:"sk"` // Sort Key

	// Core identification fields
	SuggestedVin      string `json:"suggestedVin" dynamodbav:"suggestedVin"`
	ErrorCode         string `json:"errorCode" dynamodbav:"errorCode"`
	PossibleValues    string `json:"possibleValues" dynamodbav:"possibleValues"`
	ErrorText         string `json:"errorText" dynamodbav:"errorText"`
	VehicleDescriptor string `json:"vehicleDescriptor" dynamodbav:"vehicleDescriptor"`

	// Optional text fields
	AdditionalErrorText *string `json:"additionalErrorText,omitempty" dynamodbav:"additionalErrorText,omitempty"`
	DestinationMarket   *string `json:"destinationMarket,omitempty" dynamodbav:"destinationMarket,omitempty"`
	Note                string  `json:"note" dynamodbav:"note"`

	// Vehicle basic info
	Make             string `json:"make" dynamodbav:"make"`
	ManufacturerName string `json:"manufacturerName" dynamodbav:"manufacturerName"`
	Model            string `json:"model" dynamodbav:"model"`
	ModelYear        string `json:"modelYear" dynamodbav:"modelYear"`
	Series           string `json:"series" dynamodbav:"series"`
	VehicleType      string `json:"vehicleType" dynamodbav:"vehicleType"`

	// Plant information
	PlantCity        string  `json:"plantCity" dynamodbav:"plantCity"`
	PlantCountry     string  `json:"plantCountry" dynamodbav:"plantCountry"`
	PlantState       *string `json:"plantState,omitempty" dynamodbav:"plantState,omitempty"`
	PlantCompanyName *string `json:"plantCompanyName,omitempty" dynamodbav:"plantCompanyName,omitempty"`

	// Trim and series
	Trim    *string `json:"trim,omitempty" dynamodbav:"trim,omitempty"`
	Trim2   *string `json:"trim2,omitempty" dynamodbav:"trim2,omitempty"`
	Series2 *string `json:"series2,omitempty" dynamodbav:"series2,omitempty"`

	// Pricing
	BasePrice  *string `json:"basePrice,omitempty" dynamodbav:"basePrice,omitempty"`
	NonLandUse *string `json:"nonLandUse,omitempty" dynamodbav:"nonLandUse,omitempty"`

	// Body specifications
	BodyClass        string  `json:"bodyClass" dynamodbav:"bodyClass"`
	Doors            string  `json:"doors" dynamodbav:"doors"`
	Windows          *string `json:"windows,omitempty" dynamodbav:"windows,omitempty"`
	WheelBaseType    *string `json:"wheelBaseType,omitempty" dynamodbav:"wheelBaseType,omitempty"`
	TrackWidthInches *string `json:"trackWidthInches,omitempty" dynamodbav:"trackWidthInches,omitempty"`

	// Weight specifications
	GrossVehicleWeightRatingFrom     string  `json:"grossVehicleWeightRatingFrom" dynamodbav:"grossVehicleWeightRatingFrom"`
	GrossVehicleWeightRatingTo       string  `json:"grossVehicleWeightRatingTo" dynamodbav:"grossVehicleWeightRatingTo"`
	GrossCombinationWeightRatingFrom *string `json:"grossCombinationWeightRatingFrom,omitempty" dynamodbav:"grossCombinationWeightRatingFrom,omitempty"`
	GrossCombinationWeightRatingTo   *string `json:"grossCombinationWeightRatingTo,omitempty" dynamodbav:"grossCombinationWeightRatingTo,omitempty"`
	CurbWeightPounds                 *string `json:"curbWeightPounds,omitempty" dynamodbav:"curbWeightPounds,omitempty"`

	// Dimensions
	BedLengthInches     *string `json:"bedLengthInches,omitempty" dynamodbav:"bedLengthInches,omitempty"`
	WheelBaseInchesFrom string  `json:"wheelBaseInchesFrom" dynamodbav:"wheelBaseInchesFrom"`
	WheelBaseInchesTo   *string `json:"wheelBaseInchesTo,omitempty" dynamodbav:"wheelBaseInchesTo,omitempty"`

	// Truck/Trailer specific
	BedType               string  `json:"bedType" dynamodbav:"bedType"`
	CabType               string  `json:"cabType" dynamodbav:"cabType"`
	TrailerTypeConnection string  `json:"trailerTypeConnection" dynamodbav:"trailerTypeConnection"`
	TrailerBodyType       string  `json:"trailerBodyType" dynamodbav:"trailerBodyType"`
	TrailerLengthFeet     *string `json:"trailerLengthFeet,omitempty" dynamodbav:"trailerLengthFeet,omitempty"`
	OtherTrailerInfo      *string `json:"otherTrailerInfo,omitempty" dynamodbav:"otherTrailerInfo,omitempty"`

	// Wheels
	NumberOfWheels       *string `json:"numberOfWheels,omitempty" dynamodbav:"numberOfWheels,omitempty"`
	WheelSizeFrontInches *string `json:"wheelSizeFrontInches,omitempty" dynamodbav:"wheelSizeFrontInches,omitempty"`
	WheelSizeRearInches  *string `json:"wheelSizeRearInches,omitempty" dynamodbav:"wheelSizeRearInches,omitempty"`

	// Motorcycle specific
	CustomMotorcycleType     string  `json:"customMotorcycleType" dynamodbav:"customMotorcycleType"`
	MotorcycleSuspensionType string  `json:"motorcycleSuspensionType" dynamodbav:"motorcycleSuspensionType"`
	MotorcycleChassisType    string  `json:"motorcycleChassisType" dynamodbav:"motorcycleChassisType"`
	OtherMotorcycleInfo      *string `json:"otherMotorcycleInfo,omitempty" dynamodbav:"otherMotorcycleInfo,omitempty"`

	// Fuel tank
	FuelTankType     *string `json:"fuelTankType,omitempty" dynamodbav:"fuelTankType,omitempty"`
	FuelTankMaterial *string `json:"fuelTankMaterial,omitempty" dynamodbav:"fuelTankMaterial,omitempty"`

	// Motorcycle braking
	CombinedBrakingSystem *string `json:"combinedBrakingSystem,omitempty" dynamodbav:"combinedBrakingSystem,omitempty"`
	WheelieMitigation     *string `json:"wheelieMitigation,omitempty" dynamodbav:"wheelieMitigation,omitempty"`

	// Bus specific
	BusLengthFeet             *string `json:"busLengthFeet,omitempty" dynamodbav:"busLengthFeet,omitempty"`
	BusFloorConfigurationType string  `json:"busFloorConfigurationType" dynamodbav:"busFloorConfigurationType"`
	BusType                   string  `json:"busType" dynamodbav:"busType"`
	OtherBusInfo              *string `json:"otherBusInfo,omitempty" dynamodbav:"otherBusInfo,omitempty"`

	// Interior
	EntertainmentSystem *string `json:"entertainmentSystem,omitempty" dynamodbav:"entertainmentSystem,omitempty"`
	SteeringLocation    *string `json:"steeringLocation,omitempty" dynamodbav:"steeringLocation,omitempty"`
	NumberOfSeats       *string `json:"numberOfSeats,omitempty" dynamodbav:"numberOfSeats,omitempty"`
	NumberOfSeatRows    *string `json:"numberOfSeatRows,omitempty" dynamodbav:"numberOfSeatRows,omitempty"`

	// Drivetrain
	TransmissionStyle  *string `json:"transmissionStyle,omitempty" dynamodbav:"transmissionStyle,omitempty"`
	TransmissionSpeeds *string `json:"transmissionSpeeds,omitempty" dynamodbav:"transmissionSpeeds,omitempty"`
	DriveType          *string `json:"driveType,omitempty" dynamodbav:"driveType,omitempty"`
	Axles              *string `json:"axles,omitempty" dynamodbav:"axles,omitempty"`
	AxleConfiguration  *string `json:"axleConfiguration,omitempty" dynamodbav:"axleConfiguration,omitempty"`

	// Braking system
	BrakeSystemType        *string `json:"brakeSystemType,omitempty" dynamodbav:"brakeSystemType,omitempty"`
	BrakeSystemDescription *string `json:"brakeSystemDescription,omitempty" dynamodbav:"brakeSystemDescription,omitempty"`

	// Battery/Electric
	OtherBatteryInfo               *string `json:"otherBatteryInfo,omitempty" dynamodbav:"otherBatteryInfo,omitempty"`
	BatteryType                    *string `json:"batteryType,omitempty" dynamodbav:"batteryType,omitempty"`
	NumberOfBatteryCellsPerModule  *string `json:"numberOfBatteryCellsPerModule,omitempty" dynamodbav:"numberOfBatteryCellsPerModule,omitempty"`
	BatteryCurrentAmpsFrom         *string `json:"batteryCurrentAmpsFrom,omitempty" dynamodbav:"batteryCurrentAmpsFrom,omitempty"`
	BatteryVoltageVoltsFrom        *string `json:"batteryVoltageVoltsFrom,omitempty" dynamodbav:"batteryVoltageVoltsFrom,omitempty"`
	BatteryEnergyKwhFrom           *string `json:"batteryEnergyKwhFrom,omitempty" dynamodbav:"batteryEnergyKwhFrom,omitempty"`
	EVDriveUnit                    *string `json:"evDriveUnit,omitempty" dynamodbav:"evDriveUnit,omitempty"`
	BatteryCurrentAmpsTo           *string `json:"batteryCurrentAmpsTo,omitempty" dynamodbav:"batteryCurrentAmpsTo,omitempty"`
	BatteryVoltageVoltsTo          *string `json:"batteryVoltageVoltsTo,omitempty" dynamodbav:"batteryVoltageVoltsTo,omitempty"`
	BatteryEnergyKwhTo             *string `json:"batteryEnergyKwhTo,omitempty" dynamodbav:"batteryEnergyKwhTo,omitempty"`
	NumberOfBatteryModulesPerPack  *string `json:"numberOfBatteryModulesPerPack,omitempty" dynamodbav:"numberOfBatteryModulesPerPack,omitempty"`
	NumberOfBatteryPacksPerVehicle *string `json:"numberOfBatteryPacksPerVehicle,omitempty" dynamodbav:"numberOfBatteryPacksPerVehicle,omitempty"`
	ChargerLevel                   *string `json:"chargerLevel,omitempty" dynamodbav:"chargerLevel,omitempty"`
	ChargerPowerKw                 *string `json:"chargerPowerKw,omitempty" dynamodbav:"chargerPowerKw,omitempty"`

	// Engine specifications
	EngineNumberOfCylinders string  `json:"engineNumberOfCylinders" dynamodbav:"engineNumberOfCylinders"`
	DisplacementCc          string  `json:"displacementCc" dynamodbav:"displacementCc"`
	DisplacementCi          string  `json:"displacementCi" dynamodbav:"displacementCi"`
	DisplacementL           string  `json:"displacementL" dynamodbav:"displacementL"`
	EngineStrokeCycles      *string `json:"engineStrokeCycles,omitempty" dynamodbav:"engineStrokeCycles,omitempty"`
	EngineModel             *string `json:"engineModel,omitempty" dynamodbav:"engineModel,omitempty"`
	EnginePowerKw           *string `json:"enginePowerKw,omitempty" dynamodbav:"enginePowerKw,omitempty"`
	FuelTypePrimary         string  `json:"fuelTypePrimary" dynamodbav:"fuelTypePrimary"`
	ValveTrainDesign        *string `json:"valveTrainDesign,omitempty" dynamodbav:"valveTrainDesign,omitempty"`
	EngineConfiguration     *string `json:"engineConfiguration,omitempty" dynamodbav:"engineConfiguration,omitempty"`
	FuelTypeSecondary       *string `json:"fuelTypeSecondary,omitempty" dynamodbav:"fuelTypeSecondary,omitempty"`

	// Fuel delivery
	FuelDeliveryFuelInjectionType *string `json:"fuelDeliveryFuelInjectionType,omitempty" dynamodbav:"fuelDeliveryFuelInjectionType,omitempty"`
	EngineBrakeHpFrom             string  `json:"engineBrakeHpFrom" dynamodbav:"engineBrakeHpFrom"`
	CoolingType                   *string `json:"coolingType,omitempty" dynamodbav:"coolingType,omitempty"`
	EngineBrakeHpTo               *string `json:"engineBrakeHpTo,omitempty" dynamodbav:"engineBrakeHpTo,omitempty"`
	ElectrificationLevel          *string `json:"electrificationLevel,omitempty" dynamodbav:"electrificationLevel,omitempty"`
	OtherEngineInfo               *string `json:"otherEngineInfo,omitempty" dynamodbav:"otherEngineInfo,omitempty"`
	Turbo                         *string `json:"turbo,omitempty" dynamodbav:"turbo,omitempty"`
	TopSpeedMph                   *string `json:"topSpeedMph,omitempty" dynamodbav:"topSpeedMph,omitempty"`
	EngineManufacturer            *string `json:"engineManufacturer,omitempty" dynamodbav:"engineManufacturer,omitempty"`

	// Safety - Restraints
	Pretensioner             *string `json:"pretensioner,omitempty" dynamodbav:"pretensioner,omitempty"`
	SeatBeltType             string  `json:"seatBeltType" dynamodbav:"seatBeltType"`
	OtherRestraintSystemInfo string  `json:"otherRestraintSystemInfo" dynamodbav:"otherRestraintSystemInfo"`

	// Airbags
	CurtainAirBagLocations     *string `json:"curtainAirBagLocations,omitempty" dynamodbav:"curtainAirBagLocations,omitempty"`
	SeatCushionAirBagLocations *string `json:"seatCushionAirBagLocations,omitempty" dynamodbav:"seatCushionAirBagLocations,omitempty"`
	FrontAirBagLocations       string  `json:"frontAirBagLocations" dynamodbav:"frontAirBagLocations"`
	KneeAirBagLocations        *string `json:"kneeAirBagLocations,omitempty" dynamodbav:"kneeAirBagLocations,omitempty"`
	SideAirBagLocations        *string `json:"sideAirBagLocations,omitempty" dynamodbav:"sideAirBagLocations,omitempty"`

	// Electronic safety systems
	AntiLockBrakingSystem            *string `json:"antiLockBrakingSystem,omitempty" dynamodbav:"antiLockBrakingSystem,omitempty"`
	ElectronicStabilityControl       *string `json:"electronicStabilityControl,omitempty" dynamodbav:"electronicStabilityControl,omitempty"`
	TractionControl                  *string `json:"tractionControl,omitempty" dynamodbav:"tractionControl,omitempty"`
	TirePressureMonitoringSystemType *string `json:"tirePressureMonitoringSystemType,omitempty" dynamodbav:"tirePressureMonitoringSystemType,omitempty"`

	// Active safety
	ActiveSafetySystemNote                 *string `json:"activeSafetySystemNote,omitempty" dynamodbav:"activeSafetySystemNote,omitempty"`
	AutoReverseSystemForWindowsAndSunroofs *string `json:"autoReverseSystemForWindowsAndSunroofs,omitempty" dynamodbav:"autoReverseSystemForWindowsAndSunroofs,omitempty"`
	AutomaticPedestrianAlertingSound       *string `json:"automaticPedestrianAlertingSound,omitempty" dynamodbav:"automaticPedestrianAlertingSound,omitempty"`
	EventDataRecorder                      *string `json:"eventDataRecorder,omitempty" dynamodbav:"eventDataRecorder,omitempty"`
	KeylessIgnition                        *string `json:"keylessIgnition,omitempty" dynamodbav:"keylessIgnition,omitempty"`
	SAEAutomationLevelFrom                 *string `json:"saeAutomationLevelFrom,omitempty" dynamodbav:"saeAutomationLevelFrom,omitempty"`
	SAEAutomationLevelTo                   *string `json:"saeAutomationLevelTo,omitempty" dynamodbav:"saeAutomationLevelTo,omitempty"`

	// ADAS features
	AdaptiveCruiseControl               *string `json:"adaptiveCruiseControl,omitempty" dynamodbav:"adaptiveCruiseControl,omitempty"`
	CrashImminentBraking                *string `json:"crashImminentBraking,omitempty" dynamodbav:"crashImminentBraking,omitempty"`
	ForwardCollisionWarning             *string `json:"forwardCollisionWarning,omitempty" dynamodbav:"forwardCollisionWarning,omitempty"`
	DynamicBrakeSupport                 *string `json:"dynamicBrakeSupport,omitempty" dynamodbav:"dynamicBrakeSupport,omitempty"`
	PedestrianAutomaticEmergencyBraking *string `json:"pedestrianAutomaticEmergencyBraking,omitempty" dynamodbav:"pedestrianAutomaticEmergencyBraking,omitempty"`
	BlindSpotWarning                    *string `json:"blindSpotWarning,omitempty" dynamodbav:"blindSpotWarning,omitempty"`
	LaneDepartureWarning                *string `json:"laneDepartureWarning,omitempty" dynamodbav:"laneDepartureWarning,omitempty"`
	LaneKeepingAssistance               *string `json:"laneKeepingAssistance,omitempty" dynamodbav:"laneKeepingAssistance,omitempty"`
	BlindSpotIntervention               *string `json:"blindSpotIntervention,omitempty" dynamodbav:"blindSpotIntervention,omitempty"`
	LaneCenteringAssistance             *string `json:"laneCenteringAssistance,omitempty" dynamodbav:"laneCenteringAssistance,omitempty"`

	// Parking assistance
	BackupCamera                  *string `json:"backupCamera,omitempty" dynamodbav:"backupCamera,omitempty"`
	ParkingAssist                 *string `json:"parkingAssist,omitempty" dynamodbav:"parkingAssist,omitempty"`
	RearCrossTrafficAlert         *string `json:"rearCrossTrafficAlert,omitempty" dynamodbav:"rearCrossTrafficAlert,omitempty"`
	RearAutomaticEmergencyBraking *string `json:"rearAutomaticEmergencyBraking,omitempty" dynamodbav:"rearAutomaticEmergencyBraking,omitempty"`

	// Communication systems
	AutomaticCrashNotification *string `json:"automaticCrashNotification,omitempty" dynamodbav:"automaticCrashNotification,omitempty"`

	// Lighting
	DaytimeRunningLight                *string `json:"daytimeRunningLight,omitempty" dynamodbav:"daytimeRunningLight,omitempty"`
	HeadlampLightSource                *string `json:"headlampLightSource,omitempty" dynamodbav:"headlampLightSource,omitempty"`
	SemiautomaticHeadlampBeamSwitching *string `json:"semiautomaticHeadlampBeamSwitching,omitempty" dynamodbav:"semiautomaticHeadlampBeamSwitching,omitempty"`
	AdaptiveDrivingBeam                *string `json:"adaptiveDrivingBeam,omitempty" dynamodbav:"adaptiveDrivingBeam,omitempty"`

	// Timestamp fields
	CreatedAt int64 `json:"createdAt" dynamodbav:"createdAt"`
	UpdatedAt int64 `json:"updatedAt" dynamodbav:"updatedAt"`
	DeletedAt int64 `json:"deletedAt" dynamodbav:"deletedAt"`

	// Extended data
	ExtendedAttributes []ExtendedAttribute `json:"extendedAttributes,omitempty" dynamodbav:"extendedAttributes,omitempty"`
	AcesAttributes     []AcesAttribute     `json:"acesAttributes,omitempty" dynamodbav:"acesAttributes,omitempty"`
}

// GetKey returns the composite primary key for DynamoDB operations (PK + SK)
func (u *Unit) GetKey() map[string]types.AttributeValue {
	key, _ := attributevalue.MarshalMap(map[string]interface{}{
		"pk": u.AccountID,  // Primary Key (AccountID)
		"sk": u.LocationID, // Sort Key (LocationID)
	})
	return key
}

// SetTimestamps sets created and updated timestamps
func (u *Unit) SetTimestamps() {
	now := time.Now().Unix()
	if u.CreatedAt == 0 {
		u.CreatedAt = now
	}
	u.UpdatedAt = now
}

// IsDeleted checks if the unit is soft deleted
func (u *Unit) IsDeleted() bool {
	return u.DeletedAt > 0
}

// MarkDeleted marks the unit as soft deleted
func (u *Unit) MarkDeleted() {
	u.DeletedAt = time.Now().Unix()
}

// GenerateID generates a new UUID for the unit's primary key
func (u *Unit) GenerateID() {
	u.ID = uuid.New().String()
}

// ValidateLocationID validates that the LocationID is a valid UUID
func (u *Unit) ValidateLocationID() error {
	if u.LocationID == "" {
		return fmt.Errorf("locationId is required")
	}
	if _, err := uuid.Parse(u.LocationID); err != nil {
		return fmt.Errorf("locationId must be a valid UUID: %w", err)
	}
	return nil
}
