# EC2 Instance for API Server

# Get latest Amazon Linux 2023 AMI
data "aws_ami" "amazon_linux_2023" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["al2023-ami-*-x86_64"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
}

# IAM Role for EC2
resource "aws_iam_role" "ec2" {
  name = "${local.name_prefix}-ec2-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      }
    ]
  })

  tags = {
    Name = "${local.name_prefix}-ec2-role"
  }
}

# IAM Policy for EC2 to access RDS and Secrets Manager
resource "aws_iam_role_policy" "ec2_policy" {
  name = "${local.name_prefix}-ec2-policy"
  role = aws_iam_role.ec2.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "rds:DescribeDBInstances",
          "rds:DescribeDBClusters"
        ]
        Resource = "*"
      },
      {
        Effect = "Allow"
        Action = [
          "secretsmanager:GetSecretValue",
          "secretsmanager:DescribeSecret"
        ]
        Resource = [
          aws_secretsmanager_secret.db_credentials.arn
        ]
      },
      {
        Effect = "Allow"
        Action = [
          "logs:CreateLogGroup",
          "logs:CreateLogStream",
          "logs:PutLogEvents"
        ]
        Resource = "*"
      },
      {
        Effect = "Allow"
        Action = [
          "ssm:GetParameter",
          "ssm:GetParameters",
          "ssm:GetParametersByPath"
        ]
        Resource = "arn:aws:ssm:${var.aws_region}:${data.aws_caller_identity.current.account_id}:parameter/${local.name_prefix}/*"
      }
    ]
  })
}

# Attach SSM managed policy for Session Manager access
resource "aws_iam_role_policy_attachment" "ec2_ssm" {
  role       = aws_iam_role.ec2.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"
}

# IAM Instance Profile
resource "aws_iam_instance_profile" "ec2" {
  name = "${local.name_prefix}-ec2-profile"
  role = aws_iam_role.ec2.name
}

# EC2 Instance
resource "aws_instance" "api_server" {
  ami                    = var.ec2_ami_id != "" ? var.ec2_ami_id : data.aws_ami.amazon_linux_2023.id
  instance_type          = var.ec2_instance_type
  key_name               = var.ec2_key_name != "" ? var.ec2_key_name : null
  subnet_id              = aws_subnet.public[0].id
  vpc_security_group_ids = [aws_security_group.ec2.id]
  iam_instance_profile   = aws_iam_instance_profile.ec2.name

  root_block_device {
    volume_type           = "gp3"
    volume_size           = 20
    encrypted             = true
    delete_on_termination = true
  }

  user_data = base64encode(templatefile("${path.module}/templates/user_data.sh", {
    aws_region         = var.aws_region
    db_secret_arn      = aws_secretsmanager_secret.db_credentials.arn
    app_port           = var.app_port
    environment        = var.environment
  }))

  tags = {
    Name = "${local.name_prefix}-api-server"
  }

  depends_on = [
    aws_db_instance.main,
    aws_secretsmanager_secret_version.db_credentials
  ]
}

# Elastic IP for EC2
resource "aws_eip" "api_server" {
  instance = aws_instance.api_server.id
  domain   = "vpc"

  tags = {
    Name = "${local.name_prefix}-api-eip"
  }
}
