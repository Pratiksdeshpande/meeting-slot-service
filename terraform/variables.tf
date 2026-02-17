# Variables for Meeting Slot Service Infrastructure

variable "aws_region" {
  description = "AWS region to deploy resources"
  type        = string
  default     = "us-east-1"
}

variable "environment" {
  description = "Environment name (dev, staging, prod)"
  type        = string
  default     = "dev"
}

variable "vpc_cidr" {
  description = "CIDR block for VPC"
  type        = string
  default     = "10.0.0.0/16"
}

variable "public_subnet_cidrs" {
  description = "CIDR blocks for public subnets"
  type        = list(string)
  default     = ["10.0.1.0/24", "10.0.2.0/24"]
}

variable "private_subnet_cidrs" {
  description = "CIDR blocks for private subnets"
  type        = list(string)
  default     = ["10.0.10.0/24", "10.0.11.0/24"]
}

# EC2 Variables
variable "ec2_instance_type" {
  description = "EC2 instance type for API server"
  type        = string
  default     = "t3.micro"
}

variable "ec2_key_name" {
  description = "Name of the EC2 key pair for SSH access"
  type        = string
  default     = ""
}

variable "ec2_ami_id" {
  description = "AMI ID for EC2 instance (Amazon Linux 2023)"
  type        = string
  default     = "" # Will use latest Amazon Linux 2023 if empty
}

# RDS Variables
variable "db_instance_class" {
  description = "RDS instance class"
  type        = string
  default     = "db.t3.micro"
}

variable "db_name" {
  description = "Database name"
  type        = string
  default     = "meetingslotdb"
}

variable "db_username" {
  description = "Database master username"
  type        = string
  default     = "admin"
  sensitive   = true
}

variable "db_password" {
  description = "Database master password"
  type        = string
  sensitive   = true
}

variable "db_allocated_storage" {
  description = "Allocated storage for RDS in GB"
  type        = number
  default     = 20
}

variable "db_engine_version" {
  description = "MySQL engine version"
  type        = string
  default     = "8.0"
}

# API Gateway Variables
variable "api_stage_name" {
  description = "API Gateway stage name"
  type        = string
  default     = "v1"
}

# Application Variables
variable "app_port" {
  description = "Port the application runs on"
  type        = number
  default     = 8080
}
