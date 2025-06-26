package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	internalConfig "github.com/steverhoton/unt-units-svc/internal/config"
	"github.com/steverhoton/unt-units-svc/internal/handlers"
	"github.com/steverhoton/unt-units-svc/internal/repository"
	"github.com/steverhoton/unt-units-svc/pkg/appsync"
)

// Dependencies holds all the application dependencies
type Dependencies struct {
	Config    *internalConfig.Config
	DDBClient *dynamodb.Client
	Repo      repository.UnitRepository
	Handlers  *handlers.UnitHandlers
}

// Global dependencies - initialized once
var deps *Dependencies

// init initializes the lambda dependencies
func init() {
	var err error
	deps, err = initializeDependencies()
	if err != nil {
		log.Fatalf("Failed to initialize dependencies: %v", err)
	}
	log.Println("Lambda dependencies initialized successfully")
}

// initializeDependencies sets up all the required dependencies
func initializeDependencies() (*Dependencies, error) {
	// Load configuration
	cfg, err := internalConfig.New()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Configure AWS SDK
	awsCfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(cfg.Region),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS configuration: %w", err)
	}

	// Create DynamoDB client
	ddbClient := dynamodb.NewFromConfig(awsCfg)

	// Create repository
	repo := repository.NewDynamoDBUnitRepository(ddbClient, cfg.TableName)

	// Create handlers
	unitHandlers := handlers.NewUnitHandlers(repo)

	return &Dependencies{
		Config:    cfg,
		DDBClient: ddbClient,
		Repo:      repo,
		Handlers:  unitHandlers,
	}, nil
}

// handler is the main lambda handler function
func handler(ctx context.Context, event json.RawMessage) (*appsync.Response, error) {
	log.Printf("Lambda invoked with event: %s", string(event))

	// Parse the AppSync event
	var appSyncEvent appsync.AppSyncEvent
	if err := json.Unmarshal(event, &appSyncEvent); err != nil {
		log.Printf("Failed to parse AppSync event: %v", err)
		return appsync.NewErrorResponse("PARSE_ERROR", "Failed to parse event", err.Error()), nil
	}

	// Always dump the event for debugging (as requested)
	deps.Handlers.DumpEvent(ctx, &appSyncEvent)

	// Route to appropriate handler based on operation type
	switch appSyncEvent.GetOperationType() {
	case appsync.OperationTypeCreate:
		log.Println("Routing to Create handler")
		return deps.Handlers.HandleCreate(ctx, &appSyncEvent)

	case appsync.OperationTypeRead:
		log.Println("Routing to Read handler")
		return deps.Handlers.HandleRead(ctx, &appSyncEvent)

	case appsync.OperationTypeUpdate:
		log.Println("Routing to Update handler")
		return deps.Handlers.HandleUpdate(ctx, &appSyncEvent)

	case appsync.OperationTypeDelete:
		log.Println("Routing to Delete handler")
		return deps.Handlers.HandleDelete(ctx, &appSyncEvent)

	case appsync.OperationTypeList:
		log.Println("Routing to List handler")
		return deps.Handlers.HandleList(ctx, &appSyncEvent)

	default:
		log.Printf("Unknown operation type: %s", appSyncEvent.FieldName)
		return appsync.NewErrorResponse("UNKNOWN_OPERATION",
			fmt.Sprintf("Unknown operation: %s", appSyncEvent.FieldName),
			""), nil
	}
}

func main() {
	// Set log prefix
	log.SetPrefix("[UNT-UNITS-LAMBDA] ")
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Log startup info
	log.Printf("Starting UNT Units Service Lambda")
	log.Printf("Table Name: %s", deps.Config.TableName)
	log.Printf("Region: %s", deps.Config.Region)
	log.Printf("Log Level: %s", deps.Config.LogLevel)

	// Check if running in local development mode
	if os.Getenv("LOCAL_DEV") == "true" {
		log.Println("Running in local development mode")
		// You could add local testing code here
		return
	}

	// Start the lambda
	lambda.Start(handler)
}
