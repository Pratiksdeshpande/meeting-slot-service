# RDS MySQL Database

# DB Subnet Group
resource "aws_db_subnet_group" "main" {
  name        = "${local.name_prefix}-db-subnet-group"
  description = "Database subnet group for Meeting Slot Service"
  subnet_ids  = aws_subnet.private[*].id

  tags = {
    Name = "${local.name_prefix}-db-subnet-group"
  }
}

# RDS MySQL Instance
resource "aws_db_instance" "main" {
  identifier = "${local.name_prefix}-mysql"

  # Engine configuration
  engine               = "mysql"
  engine_version       = var.db_engine_version
  instance_class       = var.db_instance_class
  allocated_storage    = var.db_allocated_storage
  max_allocated_storage = var.db_allocated_storage * 2 # Auto-scaling

  # Database configuration
  db_name  = var.db_name
  username = var.db_username
  password = var.db_password
  port     = 3306

  # Network configuration
  db_subnet_group_name   = aws_db_subnet_group.main.name
  vpc_security_group_ids = [aws_security_group.rds.id]
  publicly_accessible    = false
  multi_az               = var.environment == "prod" ? true : false

  # Storage configuration
  storage_type          = "gp3"
  storage_encrypted     = true

  # Backup configuration
  backup_retention_period = var.environment == "prod" ? 7 : 1
  backup_window           = "03:00-04:00"
  maintenance_window      = "Mon:04:00-Mon:05:00"

  # Performance and monitoring
  performance_insights_enabled = var.environment == "prod" ? true : false
  monitoring_interval          = var.environment == "prod" ? 60 : 0

  # Other settings
  auto_minor_version_upgrade = true
  deletion_protection        = var.environment == "prod" ? true : false
  skip_final_snapshot        = var.environment == "prod" ? false : true
  final_snapshot_identifier  = var.environment == "prod" ? "${local.name_prefix}-final-snapshot" : null

  tags = {
    Name = "${local.name_prefix}-mysql"
  }
}

# Store DB credentials in Secrets Manager
resource "aws_secretsmanager_secret" "db_credentials" {
  name        = "${local.name_prefix}/db-credentials"
  description = "Database credentials for Meeting Slot Service"

  tags = {
    Name = "${local.name_prefix}-db-credentials"
  }
}

resource "aws_secretsmanager_secret_version" "db_credentials" {
  secret_id = aws_secretsmanager_secret.db_credentials.id
  secret_string = jsonencode({
    username = var.db_username
    password = var.db_password
    host     = aws_db_instance.main.address
    port     = aws_db_instance.main.port
    dbname   = var.db_name
    engine   = "mysql"
  })
}
