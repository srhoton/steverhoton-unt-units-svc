# Local values for resource naming and common configurations
locals {
  name_prefix = "${var.project_name}-${var.environment}"

  common_tags = merge(var.tags, {
    Environment = var.environment
    Project     = var.project_name
    ManagedBy   = "terraform"
    Service     = "unt-units-service"
  })

  lambda_source_dir = "${path.module}/../lambda"
  lambda_build_dir  = "${path.module}/build"
  lambda_zip_path   = "${local.lambda_build_dir}/lambda.zip"
}

# DynamoDB table for storing unit data
resource "aws_dynamodb_table" "units_table" {
  name           = local.name_prefix
  billing_mode   = var.dynamodb_billing_mode
  hash_key       = "pk"
  range_key      = "sk"
  stream_enabled = false

  # Only set capacity if using PROVISIONED billing mode
  read_capacity  = var.dynamodb_billing_mode == "PROVISIONED" ? var.dynamodb_read_capacity : null
  write_capacity = var.dynamodb_billing_mode == "PROVISIONED" ? var.dynamodb_write_capacity : null

  attribute {
    name = "pk"
    type = "S"
  }

  attribute {
    name = "sk"
    type = "S"
  }

  attribute {
    name = "id"
    type = "S"
  }

  # Global Secondary Index for querying by unit ID across accounts
  global_secondary_index {
    name            = "unit-id-index"
    hash_key        = "id"
    projection_type = "ALL"

    # Only set capacity if using PROVISIONED billing mode
    read_capacity  = var.dynamodb_billing_mode == "PROVISIONED" ? var.dynamodb_read_capacity : null
    write_capacity = var.dynamodb_billing_mode == "PROVISIONED" ? var.dynamodb_write_capacity : null
  }

  point_in_time_recovery {
    enabled = var.enable_point_in_time_recovery
  }

  deletion_protection_enabled = var.enable_deletion_protection

  tags = merge(local.common_tags, {
    Name = "${local.name_prefix}-units-table"
  })
}

# IAM role for Lambda function
resource "aws_iam_role" "lambda_role" {
  name = "${local.name_prefix}-lambda-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      }
    ]
  })

  tags = merge(local.common_tags, {
    Name = "${local.name_prefix}-lambda-role"
  })
}

# IAM policy for DynamoDB access
resource "aws_iam_policy" "lambda_dynamodb_policy" {
  name        = "${local.name_prefix}-lambda-dynamodb-policy"
  description = "IAM policy for Lambda function to access DynamoDB"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "dynamodb:GetItem",
          "dynamodb:PutItem",
          "dynamodb:UpdateItem",
          "dynamodb:DeleteItem",
          "dynamodb:Query",
          "dynamodb:Scan"
        ]
        Resource = [
          aws_dynamodb_table.units_table.arn,
          "${aws_dynamodb_table.units_table.arn}/index/*"
        ]
      }
    ]
  })

  tags = merge(local.common_tags, {
    Name = "${local.name_prefix}-lambda-dynamodb-policy"
  })
}

# Attach DynamoDB policy to Lambda role
resource "aws_iam_role_policy_attachment" "lambda_dynamodb_policy_attachment" {
  role       = aws_iam_role.lambda_role.name
  policy_arn = aws_iam_policy.lambda_dynamodb_policy.arn
}

# Attach AWS managed policy for basic Lambda execution
resource "aws_iam_role_policy_attachment" "lambda_basic_execution" {
  role       = aws_iam_role.lambda_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

# CloudWatch log group for Lambda function
resource "aws_cloudwatch_log_group" "lambda_log_group" {
  name              = "/aws/lambda/${local.name_prefix}"
  retention_in_days = 14

  tags = merge(local.common_tags, {
    Name = "${local.name_prefix}-lambda-logs"
  })
}

# Build the Go binary for Lambda
resource "null_resource" "lambda_build" {
  triggers = {
    # Rebuild when any Go file changes
    source_hash = data.archive_file.lambda_source_hash.output_base64sha256
  }

  provisioner "local-exec" {
    command = <<-EOT
      echo "Building Go Lambda function..."
      mkdir -p ${local.lambda_build_dir}
      cd ${local.lambda_source_dir}
      GOOS=linux GOARCH=${var.lambda_architecture} CGO_ENABLED=0 go build -o ${abspath(local.lambda_build_dir)}/bootstrap ./cmd/lambda/
      echo "Build completed successfully"
    EOT
  }

  depends_on = [data.archive_file.lambda_source_hash]
}

# Create hash of source files to trigger rebuilds
data "archive_file" "lambda_source_hash" {
  type        = "zip"
  output_path = "/tmp/lambda_source_hash.zip"

  source_dir = local.lambda_source_dir
  excludes = [
    "*.zip",
    "build/",
    ".git/",
    "*.test",
    "coverage.out"
  ]
}

# Create Lambda deployment package
data "archive_file" "lambda_zip" {
  type        = "zip"
  output_path = local.lambda_zip_path
  source_file = "${local.lambda_build_dir}/bootstrap"

  depends_on = [null_resource.lambda_build]
}

# Lambda function
resource "aws_lambda_function" "units_lambda" {
  filename         = local.lambda_zip_path
  function_name    = local.name_prefix
  role             = aws_iam_role.lambda_role.arn
  handler          = "bootstrap"
  source_code_hash = data.archive_file.lambda_zip.output_base64sha256
  runtime          = "provided.al2"
  architectures    = [var.lambda_architecture]
  timeout          = var.lambda_timeout
  memory_size      = var.lambda_memory_size

  environment {
    variables = {
      TABLE_NAME = aws_dynamodb_table.units_table.name
      LOG_LEVEL  = var.log_level
    }
  }

  depends_on = [
    aws_iam_role_policy_attachment.lambda_basic_execution,
    aws_iam_role_policy_attachment.lambda_dynamodb_policy_attachment,
    aws_cloudwatch_log_group.lambda_log_group,
    data.archive_file.lambda_zip
  ]

  tags = merge(local.common_tags, {
    Name = "${local.name_prefix}-lambda"
  })
}