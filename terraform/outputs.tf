# Outputs

# VPC Outputs
output "vpc_id" {
  description = "VPC ID"
  value       = aws_vpc.main.id
}

output "public_subnet_ids" {
  description = "Public subnet IDs"
  value       = aws_subnet.public[*].id
}

output "private_subnet_ids" {
  description = "Private subnet IDs"
  value       = aws_subnet.private[*].id
}

# EC2 Outputs
output "ec2_instance_id" {
  description = "EC2 instance ID"
  value       = aws_instance.api_server.id
}

output "ec2_public_ip" {
  description = "EC2 public IP (Elastic IP)"
  value       = aws_eip.api_server.public_ip
}

output "ec2_private_ip" {
  description = "EC2 private IP"
  value       = aws_instance.api_server.private_ip
}

# RDS Outputs
output "rds_endpoint" {
  description = "RDS endpoint"
  value       = aws_db_instance.main.endpoint
}

output "rds_address" {
  description = "RDS address (hostname)"
  value       = aws_db_instance.main.address
}

output "rds_port" {
  description = "RDS port"
  value       = aws_db_instance.main.port
}

output "db_credentials_secret_arn" {
  description = "ARN of the Secrets Manager secret containing DB credentials"
  value       = aws_secretsmanager_secret.db_credentials.arn
}

# API Gateway Outputs
output "api_gateway_id" {
  description = "API Gateway REST API ID"
  value       = aws_api_gateway_rest_api.main.id
}

output "api_gateway_url" {
  description = "API Gateway invoke URL"
  value       = aws_api_gateway_stage.main.invoke_url
}

output "api_gateway_stage" {
  description = "API Gateway stage name"
  value       = aws_api_gateway_stage.main.stage_name
}

# Connection Information
output "connection_info" {
  description = "Connection information for the service"
  value = {
    api_url         = aws_api_gateway_stage.main.invoke_url
    ec2_direct_url  = "http://${aws_eip.api_server.public_ip}:${var.app_port}"
    health_check    = "${aws_api_gateway_stage.main.invoke_url}/health"
    ssh_command     = var.ec2_key_name != "" ? "ssh -i ${var.ec2_key_name}.pem ec2-user@${aws_eip.api_server.public_ip}" : "Use SSM Session Manager"
  }
}
