# CloudWatch Logs Configuration
# Provides centralized logging for application and system logs

# Application Log Group
resource "aws_cloudwatch_log_group" "application" {
  name              = "/aws/ec2/${local.name_prefix}/application"
  retention_in_days = var.cloudwatch_log_retention_days

  tags = {
    Name        = "${local.name_prefix}-app-logs"
    Environment = var.environment
  }
}

# System Logs
resource "aws_cloudwatch_log_group" "system" {
  name              = "/aws/ec2/${local.name_prefix}/system"
  retention_in_days = var.cloudwatch_log_retention_days

  tags = {
    Name        = "${local.name_prefix}-system-logs"
    Environment = var.environment
  }
}

# Access Logs
resource "aws_cloudwatch_log_group" "access" {
  name              = "/aws/ec2/${local.name_prefix}/access"
  retention_in_days = var.cloudwatch_log_retention_days

  tags = {
    Name        = "${local.name_prefix}-access-logs"
    Environment = var.environment
  }
}

# Error Logs
resource "aws_cloudwatch_log_group" "error" {
  name              = "/aws/ec2/${local.name_prefix}/error"
  retention_in_days = var.cloudwatch_log_retention_days

  tags = {
    Name        = "${local.name_prefix}-error-logs"
    Environment = var.environment
  }
}

# ALB Access Logs S3 Bucket
resource "aws_s3_bucket" "alb_logs" {
  bucket = "${local.name_prefix}-alb-logs-${data.aws_caller_identity.current.account_id}"

  tags = {
    Name        = "${local.name_prefix}-alb-logs"
    Environment = var.environment
  }
}

# Block public access to ALB logs bucket
resource "aws_s3_bucket_public_access_block" "alb_logs" {
  bucket = aws_s3_bucket.alb_logs.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# S3 bucket lifecycle policy for ALB logs
resource "aws_s3_bucket_lifecycle_configuration" "alb_logs" {
  bucket = aws_s3_bucket.alb_logs.id

  rule {
    id     = "delete-old-logs"
    status = "Enabled"

    expiration {
      days = var.cloudwatch_log_retention_days
    }

    noncurrent_version_expiration {
      noncurrent_days = 1
    }
  }
}

# S3 bucket policy for ALB to write logs
data "aws_elb_service_account" "main" {}

resource "aws_s3_bucket_policy" "alb_logs" {
  bucket = aws_s3_bucket.alb_logs.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          AWS = data.aws_elb_service_account.main.arn
        }
        Action   = "s3:PutObject"
        Resource = "${aws_s3_bucket.alb_logs.arn}/*"
      },
      {
        Effect = "Allow"
        Principal = {
          Service = "elasticloadbalancing.amazonaws.com"
        }
        Action   = "s3:PutObject"
        Resource = "${aws_s3_bucket.alb_logs.arn}/*"
      }
    ]
  })
}

# Enable ALB access logs
resource "aws_lb" "main_with_logs" {
  # This is handled in alb.tf, but access_logs block can be added there
  # Placing this configuration here as documentation
  # access_logs {
  #   bucket  = aws_s3_bucket.alb_logs.bucket
  #   enabled = true
  # }
}

# CloudWatch Dashboard for Monitoring
resource "aws_cloudwatch_dashboard" "main" {
  dashboard_name = "${local.name_prefix}-dashboard"

  dashboard_body = jsonencode({
    widgets = [
      {
        type = "metric"
        properties = {
          metrics = [
            ["AWS/ApplicationELB", "TargetResponseTime", { stat = "Average" }],
            [".", "RequestCount", { stat = "Sum" }],
            [".", "HTTPCode_Target_2XX_Count", { stat = "Sum" }],
            [".", "HTTPCode_Target_4XX_Count", { stat = "Sum" }],
            [".", "HTTPCode_Target_5XX_Count", { stat = "Sum" }]
          ]
          period = 300
          region = var.aws_region
          title  = "ALB Metrics"
        }
      },
      {
        type = "metric"
        properties = {
          metrics = [
            ["AWS/EC2", "CPUUtilization", { stat = "Average" }],
            [".", "NetworkIn", { stat = "Sum" }],
            [".", "NetworkOut", { stat = "Sum" }]
          ]
          period = 300
          region = var.aws_region
          title  = "EC2 Metrics"
        }
      },
      {
        type = "metric"
        properties = {
          metrics = [
            ["AWS/RDS", "CPUUtilization", { stat = "Average", label = "RDS CPU" }],
            [".", "DatabaseConnections", { stat = "Average" }],
            [".", "ReadLatency", { stat = "Average" }],
            [".", "WriteLatency", { stat = "Average" }]
          ]
          period = 300
          region = var.aws_region
          title  = "RDS Metrics"
        }
      },
      {
        type = "log"
        properties = {
          query   = "SOURCE '${aws_cloudwatch_log_group.error.name}' | fields @timestamp, @message | sort @timestamp desc | limit 20"
          region  = var.aws_region
          title   = "Recent Error Logs"
        }
      }
    ]
  })
}

# CloudWatch Log Metric Filter - Error Count
resource "aws_cloudwatch_log_metric_filter" "error_count" {
  name           = "${local.name_prefix}-error-count"
  log_group_name = aws_cloudwatch_log_group.error.name
  pattern        = "[ERROR]"

  metric_transformation {
    name      = "ErrorCount"
    namespace = "${local.name_prefix}/Application"
    value     = "1"
  }
}

# CloudWatch Alarm - High Error Rate
resource "aws_cloudwatch_metric_alarm" "high_error_rate" {
  alarm_name          = "${local.name_prefix}-high-error-rate"
  alarm_description   = "Alert when error rate is high"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 2
  metric_name         = "ErrorCount"
  namespace           = "${local.name_prefix}/Application"
  period              = 300
  statistic           = "Sum"
  threshold           = 10
  treat_missing_data  = "notBreaching"

  tags = {
    Name = "${local.name_prefix}-high-error-rate"
  }
}
