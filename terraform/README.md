# UNT Units Service Infrastructure

This Terraform configuration deploys the production infrastructure for the UNT Units Service, including:

- DynamoDB table for storing unit data
- Lambda function for handling AppSync events
- IAM roles and policies for secure access
- CloudWatch log group for monitoring
- Remote state management with S3

## Prerequisites

- Terraform >= 1.0
- AWS CLI configured with appropriate credentials
- Go 1.24+ (for building the Lambda function)
- Access to the `steve-rhoton-tfstate` S3 bucket

## Quick Start

1. **Initialize Terraform (first time only):**
   ```bash
   cd terraform
   terraform init
   ```

2. **Configure variables:**
   ```bash
   cp terraform.tfvars.example terraform.tfvars
   # Edit terraform.tfvars as needed
   ```

3. **Plan the deployment:**
   ```bash
   terraform plan
   ```

4. **Apply the configuration:**
   ```bash
   terraform apply
   ```

## AppSync Integration

See [APPSYNC_RESOLVER_GUIDE.md](./APPSYNC_RESOLVER_GUIDE.md) for detailed instructions on:
- Setting up AppSync resolvers
- GraphQL schema definitions
- Example queries and mutations
- Error handling and best practices

## Configuration

### Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `aws_region` | AWS region for deployment | `us-west-2` | No |
| `environment` | Environment name (dev/staging/prod/sandbox) | `sandbox` | No |
| `project_name` | Project name | `sr-unt-units-svc` | No |
| `lambda_timeout` | Lambda timeout in seconds | `30` | No |
| `lambda_memory_size` | Lambda memory in MB | `256` | No |
| `lambda_architecture` | Lambda architecture (x86_64/arm64) | `arm64` | No |
| `log_level` | Log level (DEBUG/INFO/WARN/ERROR) | `INFO` | No |
| `dynamodb_billing_mode` | DynamoDB billing mode | `PAY_PER_REQUEST` | No |
| `enable_point_in_time_recovery` | Enable PITR for DynamoDB | `true` | No |
| `enable_deletion_protection` | Enable deletion protection | `false` | No |
| `tags` | Additional tags | `{}` | No |

### Example terraform.tfvars

```hcl
aws_region    = "us-west-2"
environment   = "sandbox"
project_name  = "sr-unt-units-svc"
lambda_timeout = 60
lambda_memory_size = 512
log_level     = "INFO"
enable_deletion_protection = true

tags = {
  Owner       = "platform-team"
  CostCenter  = "engineering"
}
```

## Architecture

### DynamoDB Table Schema

- **Table Name:** `{project_name}-{environment}-units`
- **Primary Key:** `pk` (String) - Unit ID (UUID)
- **Sort Key:** `sk` (String) - Account ID
- **Global Secondary Index:** `sk-index` - Allows querying by Account ID

### Lambda Function

- **Runtime:** `provided.al2` (Go custom runtime)
- **Handler:** `bootstrap`
- **Environment Variables:**
  - `TABLE_NAME`: DynamoDB table name
  - `AWS_REGION`: AWS region
  - `LOG_LEVEL`: Logging level

### IAM Permissions

The Lambda function has permissions to:
- Read/write/query DynamoDB table and its indexes
- Write to CloudWatch logs

## Build Process

The Terraform configuration automatically:

1. Builds the Go binary for the target architecture
2. Creates a deployment ZIP package
3. Updates the Lambda function when source code changes

## Monitoring

- CloudWatch logs are automatically created at `/aws/lambda/{project_name}-{environment}-lambda`
- Log retention is set to 14 days by default

## Security

- Follows least privilege principle for IAM policies
- Enables point-in-time recovery for DynamoDB by default
- All resources are tagged for compliance and cost tracking

## Cleanup

To destroy all resources:

```bash
terraform destroy
```

**Warning:** This will permanently delete all data in the DynamoDB table unless point-in-time recovery is enabled.