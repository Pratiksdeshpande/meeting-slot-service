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

# ALB Outputs
output "alb_dns_name" {
  description = "DNS name of the Application Load Balancer"
  value       = aws_lb.main.dns_name
}

output "alb_zone_id" {
  description = "Zone ID of the Application Load Balancer"
  value       = aws_lb.main.zone_id
}

output "alb_arn" {
  description = "ARN of the Application Load Balancer"
  value       = aws_lb.main.arn
}

output "alb_url" {
  description = "HTTP URL of the Application Load Balancer"
  value       = "http://${aws_lb.main.dns_name}"
}

# Auto Scaling Group Outputs
output "asg_name" {
  description = "Name of the Auto Scaling Group"
  value       = aws_autoscaling_group.app.name
}

output "asg_arn" {
  description = "ARN of the Auto Scaling Group"
  value       = aws_autoscaling_group.app.arn
}

output "asg_min_size" {
  description = "Minimum size of the Auto Scaling Group"
  value       = aws_autoscaling_group.app.min_size
}

output "asg_max_size" {
  description = "Maximum size of the Auto Scaling Group"
  value       = aws_autoscaling_group.app.max_size
}

output "asg_desired_capacity" {
  description = "Desired capacity of the Auto Scaling Group"
  value       = aws_autoscaling_group.app.desired_capacity
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

# CloudWatch Outputs
output "cloudwatch_log_groups" {
  description = "CloudWatch log group names"
  value = {
    application = aws_cloudwatch_log_group.application.name
    system      = aws_cloudwatch_log_group.system.name
    access      = aws_cloudwatch_log_group.access.name
    error       = aws_cloudwatch_log_group.error.name
  }
}

output "cloudwatch_dashboard_url" {
  description = "URL to CloudWatch Dashboard"
  value       = "https://console.aws.amazon.com/cloudwatch/home?region=${var.aws_region}#dashboards:name=${aws_cloudwatch_dashboard.main.dashboard_name}"
}

# Connection Information
output "connection_info" {
  description = "Connection information for the service"
  value = {
    # Primary access method
    application_url     = "http://${aws_lb.main.dns_name}"
    health_check_url    = "http://${aws_lb.main.dns_name}/health"
    
    # Monitoring
    cloudwatch_dashboard = "https://console.aws.amazon.com/cloudwatch/home?region=${var.aws_region}#dashboards:name=${aws_cloudwatch_dashboard.main.dashboard_name}"
    
    # Database
    database_endpoint   = aws_db_instance.main.endpoint
    db_secret_arn       = aws_secretsmanager_secret.db_credentials.arn
    
    # Infrastructure
    asg_name            = aws_autoscaling_group.app.name
    min_instances       = aws_autoscaling_group.app.min_size
    max_instances       = aws_autoscaling_group.app.max_size
  }
}

# Summary Output
output "deployment_summary" {
  description = "Summary of deployed resources"
  value = <<-EOT
    
    ========================================
    Meeting Slot Service - Deployment Summary
    ========================================
    
    Environment: ${var.environment}
    Region: ${var.aws_region}
    
    APPLICATION ACCESS:
    - Load Balancer URL: http://${aws_lb.main.dns_name}
    - Health Check: http://${aws_lb.main.dns_name}/health
    
    AUTO SCALING:
    - Min Instances: ${aws_autoscaling_group.app.min_size}
    - Max Instances: ${aws_autoscaling_group.app.max_size}
    - Desired: ${aws_autoscaling_group.app.desired_capacity}
    
    DATABASE:
    - RDS Endpoint: ${aws_db_instance.main.endpoint}
    - DB Name: ${aws_db_instance.main.db_name}
    - Credentials: Stored in AWS Secrets Manager
    
    MONITORING:
    - CloudWatch Dashboard: https://console.aws.amazon.com/cloudwatch/home?region=${var.aws_region}#dashboards:name=${aws_cloudwatch_dashboard.main.dashboard_name}
    - Log Groups: 
      * Application: ${aws_cloudwatch_log_group.application.name}
      * Errors: ${aws_cloudwatch_log_group.error.name}
      * System: ${aws_cloudwatch_log_group.system.name}
    
    ========================================
  EOT
}
