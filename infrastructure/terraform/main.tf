terraform {
  required_version = ">= 1.5"
  
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }

  backend "s3" {
    bucket         = "fintech-platform-terraform-state"
    key            = "production/terraform.tfstate"
    region         = "us-east-1"
    encrypt        = true
    dynamodb_table = "terraform-state-lock"
  }
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Project     = "FinTech Platform"
      Environment = var.environment
      ManagedBy   = "Terraform"
    }
  }
}

# VPC
module "vpc" {
  source = "./modules/vpc"

  project_name       = var.project_name
  environment        = var.environment
  vpc_cidr           = var.vpc_cidr
  availability_zones = var.availability_zones
}

# EKS Cluster
module "eks" {
  source = "./modules/eks"

  project_name       = var.project_name
  environment        = var.environment
  vpc_id             = module.vpc.vpc_id
  private_subnet_ids = module.vpc.private_subnet_ids
  cluster_version    = var.eks_cluster_version
}

# RDS PostgreSQL
module "rds" {
  source = "./modules/rds"

  project_name           = var.project_name
  environment            = var.environment
  vpc_id                 = module.vpc.vpc_id
  database_subnet_ids    = module.vpc.database_subnet_ids
  instance_class         = var.rds_instance_class
  allocated_storage      = var.rds_allocated_storage
  engine_version         = "15.5"
  multi_az               = true
  backup_retention_period = 7
}

# ElastiCache Redis
module "elasticache" {
  source = "./modules/elasticache"

  project_name        = var.project_name
  environment         = var.environment
  vpc_id              = module.vpc.vpc_id
  cache_subnet_ids    = module.vpc.database_subnet_ids
  node_type           = var.elasticache_node_type
  num_cache_nodes     = 2
  engine_version      = "7.0"
}

# Application Load Balancer
module "alb" {
  source = "./modules/alb"

  project_name       = var.project_name
  environment        = var.environment
  vpc_id             = module.vpc.vpc_id
  public_subnet_ids  = module.vpc.public_subnet_ids
}

# Secrets Manager
resource "aws_secretsmanager_secret" "database_credentials" {
  name = "${var.project_name}-${var.environment}-db-credentials"
  
  recovery_window_in_days = 7
}

resource "aws_secretsmanager_secret_version" "database_credentials" {
  secret_id = aws_secretsmanager_secret.database_credentials.id
  secret_string = jsonencode({
    username = module.rds.db_username
    password = module.rds.db_password
    host     = module.rds.db_endpoint
    port     = module.rds.db_port
    dbname   = module.rds.db_name
  })
}

# S3 Bucket for Application Data
resource "aws_s3_bucket" "app_data" {
  bucket = "${var.project_name}-${var.environment}-app-data"
}

resource "aws_s3_bucket_versioning" "app_data" {
  bucket = aws_s3_bucket.app_data.id
  
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "app_data" {
  bucket = aws_s3_bucket.app_data.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm     = "aws:kms"
      kms_master_key_id = aws_kms_key.app_data.arn
    }
  }
}

# KMS Key for Encryption
resource "aws_kms_key" "app_data" {
  description             = "KMS key for ${var.project_name} ${var.environment}"
  deletion_window_in_days = 10
  enable_key_rotation     = true
}

resource "aws_kms_alias" "app_data" {
  name          = "alias/${var.project_name}-${var.environment}"
  target_key_id = aws_kms_key.app_data.key_id
}

# CloudWatch Log Group
resource "aws_cloudwatch_log_group" "application" {
  name              = "/aws/fintech-platform/${var.environment}"
  retention_in_days = 30
}

# IAM Role for EKS Pods
resource "aws_iam_role" "eks_pod_role" {
  name = "${var.project_name}-${var.environment}-pod-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRoleWithWebIdentity"
      Effect = "Allow"
      Principal = {
        Federated = module.eks.oidc_provider_arn
      }
    }]
  })
}

# IAM Policy for Secrets Access
resource "aws_iam_role_policy" "secrets_access" {
  name = "secrets-access"
  role = aws_iam_role.eks_pod_role.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Action = [
        "secretsmanager:GetSecretValue",
        "secretsmanager:DescribeSecret"
      ]
      Resource = [
        aws_secretsmanager_secret.database_credentials.arn
      ]
    }]
  })
}
