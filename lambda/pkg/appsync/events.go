package appsync

import (
	"encoding/json"
)

// AppSyncEvent represents the event structure from AppSync
type AppSyncEvent struct {
	TypeName  string          `json:"typeName"`
	FieldName string          `json:"fieldName"`
	Arguments json.RawMessage `json:"arguments"`
	Identity  Identity        `json:"identity"`
	Source    json.RawMessage `json:"source,omitempty"`
	Request   RequestHeaders  `json:"request"`
	Info      Info            `json:"info"`
}

// Identity represents the caller identity from AppSync
type Identity struct {
	Sub                   string   `json:"sub"`
	Issuer                string   `json:"issuer"`
	Username              string   `json:"username"`
	Claims                Claims   `json:"claims"`
	SourceIP              []string `json:"sourceIp"`
	DefaultAuthStrategy   string   `json:"defaultAuthStrategy"`
	Groups                []string `json:"groups,omitempty"`
	UserArn               string   `json:"userArn,omitempty"`
	AccountID             string   `json:"accountId,omitempty"`
	CognitoIdentityPoolID string   `json:"cognitoIdentityPoolId,omitempty"`
	CognitoIdentityID     string   `json:"cognitoIdentityId,omitempty"`
}

// Claims represents the JWT claims
type Claims struct {
	Sub           string `json:"sub"`
	EventID       string `json:"event_id"`
	TokenUse      string `json:"token_use"`
	Scope         string `json:"scope"`
	AuthTime      int64  `json:"auth_time"`
	Iss           string `json:"iss"`
	Exp           int64  `json:"exp"`
	Iat           int64  `json:"iat"`
	Version       int    `json:"version"`
	Jti           string `json:"jti"`
	ClientID      string `json:"client_id"`
	Username      string `json:"username"`
	Email         string `json:"email,omitempty"`
	EmailVerified bool   `json:"email_verified,omitempty"`
}

// RequestHeaders represents HTTP request headers
type RequestHeaders struct {
	Headers map[string]string `json:"headers"`
}

// Info represents GraphQL execution info
type Info struct {
	FieldName           string                 `json:"fieldName"`
	ParentTypeName      string                 `json:"parentTypeName"`
	Variables           map[string]interface{} `json:"variables"`
	SelectionSetList    []string               `json:"selectionSetList"`
	SelectionSetGraphQL string                 `json:"selectionSetGraphQL"`
}

// OperationType represents the type of CRUD operation
type OperationType string

const (
	OperationTypeCreate OperationType = "CREATE"
	OperationTypeRead   OperationType = "READ"
	OperationTypeUpdate OperationType = "UPDATE"
	OperationTypeDelete OperationType = "DELETE"
	OperationTypeList   OperationType = "LIST"
)

// CreateUnitInput represents input for creating a unit
type CreateUnitInput struct {
	AccountID string                 `json:"accountId"`
	UnitType  string                 `json:"unitType"`
	Data      map[string]interface{} `json:"data"`
}

// UpdateUnitInput represents input for updating a unit
type UpdateUnitInput struct {
	ID        string                 `json:"id"`
	AccountID string                 `json:"accountId"`
	UnitType  string                 `json:"unitType"`
	Data      map[string]interface{} `json:"data"`
}

// DeleteUnitInput represents input for deleting a unit
type DeleteUnitInput struct {
	ID        string `json:"id"`
	AccountID string `json:"accountId"`
	UnitType  string `json:"unitType"`
}

// GetUnitInput represents input for getting a single unit
type GetUnitInput struct {
	ID        string `json:"id"`
	AccountID string `json:"accountId"`
	UnitType  string `json:"unitType"`
}

// ListUnitsInput represents input for listing units
type ListUnitsInput struct {
	AccountID string  `json:"accountId"`
	UnitType  *string `json:"unitType,omitempty"`
	Limit     *int    `json:"limit,omitempty"`
	NextToken *string `json:"nextToken,omitempty"`
	Filter    *string `json:"filter,omitempty"`
}

// Response represents a standard response structure
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

// ErrorInfo represents error information
type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// ListUnitsResponse represents the response for list operations
type ListUnitsResponse struct {
	Items     []map[string]interface{} `json:"items"`
	NextToken *string                  `json:"nextToken,omitempty"`
	Count     int                      `json:"count"`
}

// GetOperationType determines the operation type based on the field name
func (e *AppSyncEvent) GetOperationType() OperationType {
	switch e.FieldName {
	case "createUnit":
		return OperationTypeCreate
	case "getUnit":
		return OperationTypeRead
	case "updateUnit":
		return OperationTypeUpdate
	case "deleteUnit":
		return OperationTypeDelete
	case "listUnits":
		return OperationTypeList
	default:
		return OperationTypeRead
	}
}

// ParseArguments parses the arguments based on operation type
func (e *AppSyncEvent) ParseArguments() (interface{}, error) {
	switch e.GetOperationType() {
	case OperationTypeCreate:
		var wrapper struct {
			Input CreateUnitInput `json:"input"`
		}
		if err := json.Unmarshal(e.Arguments, &wrapper); err != nil {
			return nil, err
		}
		return wrapper.Input, nil
	case OperationTypeRead:
		var wrapper struct {
			Input GetUnitInput `json:"input"`
		}
		if err := json.Unmarshal(e.Arguments, &wrapper); err != nil {
			return nil, err
		}
		return wrapper.Input, nil
	case OperationTypeUpdate:
		var wrapper struct {
			Input UpdateUnitInput `json:"input"`
		}
		if err := json.Unmarshal(e.Arguments, &wrapper); err != nil {
			return nil, err
		}
		return wrapper.Input, nil
	case OperationTypeDelete:
		var wrapper struct {
			Input DeleteUnitInput `json:"input"`
		}
		if err := json.Unmarshal(e.Arguments, &wrapper); err != nil {
			return nil, err
		}
		return wrapper.Input, nil
	case OperationTypeList:
		var wrapper struct {
			Input ListUnitsInput `json:"input"`
		}
		if err := json.Unmarshal(e.Arguments, &wrapper); err != nil {
			return nil, err
		}
		return wrapper.Input, nil
	default:
		return nil, nil
	}
}

// NewSuccessResponse creates a successful response
func NewSuccessResponse(data interface{}, message string) *Response {
	return &Response{
		Success: true,
		Data:    data,
		Message: message,
	}
}

// NewErrorResponse creates an error response
func NewErrorResponse(code, message, details string) *Response {
	return &Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
}
