# AppSync Resolver Integration Guide

This guide explains how to integrate the UNT Units Service Lambda function with AWS AppSync as a resolver.

## Overview

The Lambda function is designed to handle AppSync events and supports full CRUD operations:
- **Create** - Create new units
- **Read** - Get units by ID
- **Update** - Update existing units  
- **Delete** - Soft delete units
- **List** - Query units by account ID with pagination

## AppSync Schema Definition

First, define your GraphQL schema with the following types:

```graphql
# Unit type definition
type Unit {
  id: ID!
  accountId: String!
  suggestedVin: String!
  errorCode: String
  possibleValues: String
  errorText: String
  vehicleDescriptor: String
  make: String!
  manufacturerName: String!
  model: String!
  modelYear: String!
  series: String!
  vehicleType: String!
  # ... add other fields as needed
  createdAt: AWSTimestamp!
  updatedAt: AWSTimestamp!
  deletedAt: AWSTimestamp
}

# Input types
input CreateUnitInput {
  accountId: String!
  suggestedVin: String!
  make: String!
  manufacturerName: String!
  model: String!
  modelYear: String!
  series: String!
  vehicleType: String!
  # ... add other required/optional fields
}

input UpdateUnitInput {
  id: ID!
  accountId: String!
  suggestedVin: String
  make: String
  model: String
  # ... add other updatable fields
}

input ListUnitsInput {
  accountId: String!
  limit: Int
  nextToken: String
}

type ListUnitsResponse {
  items: [Unit!]!
  count: Int!
  nextToken: String
}

# Query and Mutation definitions
type Query {
  getUnit(id: ID!, accountId: String!): Unit
  listUnits(input: ListUnitsInput!): ListUnitsResponse!
}

type Mutation {
  createUnit(input: CreateUnitInput!): Unit!
  updateUnit(input: UpdateUnitInput!): Unit!
  deleteUnit(id: ID!, accountId: String!): Boolean!
}
```

## Resolver Configuration

### 1. Create Data Source

Create a Lambda data source in AppSync pointing to your deployed Lambda function:

```json
{
  "name": "UntUnitsLambdaDataSource",
  "type": "AWS_LAMBDA",
  "lambdaConfig": {
    "lambdaFunctionArn": "arn:aws:lambda:us-east-1:ACCOUNT:function:unt-units-svc-prod-lambda"
  },
  "serviceRoleArn": "arn:aws:iam::ACCOUNT:role/appsync-lambda-role"
}
```

### 2. Resolver Templates

The Lambda function expects AppSync events in a specific format. Use these resolver templates:

#### Query: getUnit

**Request Template:**
```vtl
{
  "version": "2018-05-29",
  "operation": "Invoke",
  "payload": {
    "fieldName": "getUnit",
    "arguments": $util.toJson($context.arguments),
    "identity": $util.toJson($context.identity),
    "source": $util.toJson($context.source),
    "request": $util.toJson($context.request),
    "prev": $util.toJson($context.prev)
  }
}
```

**Response Template:**
```vtl
#if($context.error)
  $util.error($context.error.message, $context.error.type)
#end

#if($context.result.errorType)
  $util.error($context.result.errorMessage, $context.result.errorType)
#end

$util.toJson($context.result.data)
```

#### Query: listUnits

**Request Template:**
```vtl
{
  "version": "2018-05-29",
  "operation": "Invoke",
  "payload": {
    "fieldName": "listUnits",
    "arguments": $util.toJson($context.arguments),
    "identity": $util.toJson($context.identity),
    "source": $util.toJson($context.source),
    "request": $util.toJson($context.request),
    "prev": $util.toJson($context.prev)
  }
}
```

**Response Template:** (Same as getUnit)

#### Mutation: createUnit

**Request Template:**
```vtl
{
  "version": "2018-05-29",
  "operation": "Invoke",
  "payload": {
    "fieldName": "createUnit",
    "arguments": $util.toJson($context.arguments),
    "identity": $util.toJson($context.identity),
    "source": $util.toJson($context.source),
    "request": $util.toJson($context.request),
    "prev": $util.toJson($context.prev)
  }
}
```

**Response Template:** (Same as getUnit)

#### Mutation: updateUnit

**Request Template:**
```vtl
{
  "version": "2018-05-29",
  "operation": "Invoke",
  "payload": {
    "fieldName": "updateUnit",
    "arguments": $util.toJson($context.arguments),
    "identity": $util.toJson($context.identity),
    "source": $util.toJson($context.source),
    "request": $util.toJson($context.request),
    "prev": $util.toJson($context.prev)
  }
}
```

**Response Template:** (Same as getUnit)

#### Mutation: deleteUnit

**Request Template:**
```vtl
{
  "version": "2018-05-29",
  "operation": "Invoke",
  "payload": {
    "fieldName": "deleteUnit",
    "arguments": $util.toJson($context.arguments),
    "identity": $util.toJson($context.identity),
    "source": $util.toJson($context.source),
    "request": $util.toJson($context.request),
    "prev": $util.toJson($context.prev)
  }
}
```

**Response Template:**
```vtl
#if($context.error)
  $util.error($context.error.message, $context.error.type)
#end

#if($context.result.errorType)
  $util.error($context.result.errorMessage, $context.result.errorType)
#end

$context.result.data
```

## Example GraphQL Operations

### Create a Unit

```graphql
mutation CreateUnit($input: CreateUnitInput!) {
  createUnit(input: $input) {
    id
    accountId
    suggestedVin
    make
    model
    modelYear
    createdAt
    updatedAt
  }
}
```

**Variables:**
```json
{
  "input": {
    "accountId": "account-123",
    "suggestedVin": "1HGBH41JXMN109186",
    "make": "Honda",
    "manufacturerName": "Honda Motor Co.",
    "model": "Civic",
    "modelYear": "2021",
    "series": "Sport",
    "vehicleType": "Passenger Car",
    "bodyClass": "Sedan",
    "doors": "4",
    "grossVehicleWeightRatingFrom": "3000",
    "grossVehicleWeightRatingTo": "3500",
    "wheelBaseInchesFrom": "106.3",
    "bedType": "Not Applicable",
    "cabType": "Not Applicable",
    "trailerTypeConnection": "Not Applicable",
    "trailerBodyType": "Not Applicable",
    "customMotorcycleType": "Not Applicable",
    "motorcycleSuspensionType": "Not Applicable",
    "motorcycleChassisType": "Not Applicable",
    "busFloorConfigurationType": "Not Applicable",
    "busType": "Not Applicable",
    "engineNumberOfCylinders": "4",
    "displacementCc": "1996",
    "displacementCi": "121.8",
    "displacementL": "2.0",
    "engineBrakeHpFrom": "158",
    "fuelTypePrimary": "Gasoline",
    "seatBeltType": "Manual",
    "otherRestraintSystemInfo": "Standard airbag system",
    "frontAirBagLocations": "1st Row (Driver & Passenger)"
  }
}
```

### Get a Unit

```graphql
query GetUnit($id: ID!, $accountId: String!) {
  getUnit(id: $id, accountId: $accountId) {
    id
    accountId
    suggestedVin
    make
    model
    modelYear
    createdAt
    updatedAt
  }
}
```

**Variables:**
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "accountId": "account-123"
}
```

### List Units

```graphql
query ListUnits($input: ListUnitsInput!) {
  listUnits(input: $input) {
    items {
      id
      accountId
      suggestedVin
      make
      model
      modelYear
      createdAt
    }
    count
    nextToken
  }
}
```

**Variables:**
```json
{
  "input": {
    "accountId": "account-123",
    "limit": 20
  }
}
```

### Update a Unit

```graphql
mutation UpdateUnit($input: UpdateUnitInput!) {
  updateUnit(input: $input) {
    id
    accountId
    suggestedVin
    make
    model
    updatedAt
  }
}
```

**Variables:**
```json
{
  "input": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "accountId": "account-123",
    "make": "Honda",
    "model": "Accord"
  }
}
```

### Delete a Unit

```graphql
mutation DeleteUnit($id: ID!, $accountId: String!) {
  deleteUnit(id: $id, accountId: $accountId)
}
```

**Variables:**
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "accountId": "account-123"
}
```

## Error Handling

The Lambda function returns standardized error responses:

```json
{
  "errorType": "VALIDATION_ERROR",
  "errorMessage": "Invalid input provided",
  "errorInfo": "accountId is required"
}
```

Common error types:
- `VALIDATION_ERROR` - Invalid input data
- `NOT_FOUND` - Resource not found
- `ALREADY_EXISTS` - Resource already exists
- `INTERNAL_ERROR` - Server error

## Authentication & Authorization

The Lambda function receives the AppSync context including:
- `identity` - Contains user authentication information
- `request` - Contains headers and other request metadata

You can access the user's identity in your Lambda function via:
```go
// The identity information is available in the AppSync event
identity := appSyncEvent.Identity
```

## Performance Considerations

1. **Pagination**: Always use pagination for list operations to avoid timeouts
2. **Projection**: Only request fields you need in your GraphQL queries
3. **Caching**: Consider implementing AppSync caching for frequently accessed data
4. **Batch Operations**: For multiple operations, consider implementing batch resolvers

## Monitoring & Debugging

- Lambda logs are available in CloudWatch at `/aws/lambda/unt-units-svc-prod-lambda`
- Use the `DumpEvent` function in your Lambda to log incoming AppSync events for debugging
- Monitor DynamoDB metrics for read/write capacity and throttling
- Set up CloudWatch alarms for Lambda errors and duration if needed

## Security Best Practices

1. **Input Validation**: Always validate input data in your GraphQL schema and Lambda function
2. **Authorization**: Implement proper authorization checks based on the user's identity
3. **Rate Limiting**: Consider implementing rate limiting at the AppSync level
4. **Field-Level Security**: Use AppSync field-level authorization if needed