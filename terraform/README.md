# Meeting Slot Service - Terraform Infrastructure

This folder contains Terraform configurations to deploy the Meeting Slot Service on AWS with production-ready infrastructure.

## 📋 Table of Contents

- [Architecture Overview](#architecture-overview)
- [Infrastructure Components](#infrastructure-components)
- [Prerequisites](#prerequisites)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [Deployment](#deployment)
- [Monitoring and Logs](#monitoring-and-logs)
- [Scaling](#scaling)
- [Troubleshooting](#troubleshooting)
- [Cost Estimation](#cost-estimation)
- [Security Best Practices](#security-best-practices)

## 🏗️ Architecture Overview

```
Internet
    │
    ▼
Application Load Balancer (ALB)
    │
    ├─────────┬─────────┐
    ▼         ▼         ▼
  EC2       EC2       EC2    (Auto Scaling Group)
    │         │         │
    └─────────┴─────────┘
           │
           ▼
      RDS MySQL (Multi-AZ for prod)
           │
           ▼
    AWS Secrets Manager (DB Credentials)
           │
           ▼
    CloudWatch (Logs & Metrics)
```

### Key Features

- **High Availability**: Multi-AZ deployment with Auto Scaling
- **Horizontal Scalability**: Automatic scaling based on CPU utilization
- **Security**: Private subnets for RDS, encrypted volumes, security groups
- **Monitoring**: CloudWatch logs, metrics, and dashboard
- **Cost Optimized**: T3 instances with auto-scaling

## 🧩 Infrastructure Components

### Network Layer (vpc.tf)
- **VPC**: Isolated virtual network (10.0.0.0/16)
- **Public Subnets** (2): For ALB and NAT Gateways
- **Private Subnets** (2): For RDS MySQL
- **Internet Gateway**: Outbound internet access
- **NAT Gateways**: Private subnet internet access
- **Route Tables**: Routing configuration

### Load Balancing (alb.tf)
- **Application Load Balancer**: Distributes HTTP traffic
- **Target Group**: Health checks on `/health` endpoint
- **HTTP Listener**: Port 80 forwarding
- **HTTPS Support**: Ready for SSL termination (requires ACM certificate)

### Compute Layer (autoscaling.tf)
- **Launch Template**: EC2 instance configuration with latest Amazon Linux 2023
- **Auto Scaling Group**: Horizontal scaling (1-4 instances by default)
- **Scaling Policies**:
  - CPU > 70% → Scale Up
  - CPU < 20% → Scale Down
  - Target Tracking: Maintain 50% average CPU
- **Instance Refresh**: Zero-downtime deployments

### Database Layer (rds.tf)
- **RDS MySQL 8.0**: Fully managed database
- **Multi-AZ**: High availability for production
- **Encrypted Storage**: At-rest encryption enabled
- **Auto Scaling Storage**: 20GB initial, scales to 100GB
- **Performance Insights**: Query performance monitoring (prod only)
- **Automated Backups**: 7-day retention

### Security (security_groups.tf)
- **ALB Security Group**: Allow HTTP/HTTPS from internet
- **EC2 Security Group**: Allow traffic from ALB only + SSH
- **RDS Security Group**: Allow MySQL from EC2 only

### IAM Permissions (ec2.tf)
EC2 instances have permissions to:
- Read Secrets Manager (database credentials)
- Describe RDS instances
- Write to CloudWatch Logs
- Access SSM Parameter Store
- EC2 instance metadata

### Monitoring (cloudwatch.tf)
- **Log Groups**:
  - `/aws/ec2/{env}/application` - Application logs
  - `/aws/ec2/{env}/error` - Error logs  
  - `/aws/ec2/{env}/system` - System logs
  - `/aws/ec2/{env}/access` - Access logs
- **Metrics Dashboard**: ALB, EC2, and RDS metrics
- **Alarms**:
  - High/Low CPU (for auto-scaling)
  - High error rate
- **ALB Access Logs**: S3 bucket with lifecycle policies

## 📦 Prerequisites

1. **AWS Account** with appropriate permissions
2. **Terraform** >= 1.0 installed ([Install Guide](https://learn.hashicorp.com/terraform/getting-started/install))
3. **AWS CLI** configured ([Install Guide](https://docs.aws.amazon.com/cli/latest/userguide/install-cliv2.html))
4. **SSH Key Pair** (optional, for EC2 SSH access)
5. **Database Password** for RDS MySQL

### Required AWS Permissions
Your AWS user/role needs permissions for:
- VPC, Subnets, Internet Gateway, NAT Gateway
- EC2, Auto Scaling, Launch Templates
- ALB, Target Groups, Listeners
- RDS, DB Subnet Groups
- IAM Roles and Instance Profiles
- Secrets Manager
- CloudWatch Logs and Alarms
- S3 (for ALB logs)

## 🚀 Quick Start

### 1. Clone Repository
```bash
cd terraform/
```

### 2. Initialize Terraform
```bash
terraform init
```

### 3. Create terraform.tfvars
```hcl
# terraform.tfvars
aws_region  = "us-east-1"
environment = "dev"

# Database credentials
db_password = "YourSecurePassword123!"  # Change this!

# EC2 Configuration
ec2_instance_type = "t3.micro"
ec2_key_name      = "your-key-pair"  # Optional

# Auto Scaling
asg_min_size         = 1
asg_max_size         = 4
asg_desired_capacity = 2

# CloudWatch
cloudwatch_log_retention_days = 7
```

### 4. Review Plan
```bash
terraform plan
```

### 5. Deploy Infrastructure
```bash
terraform apply
```

### 6. Get ALB URL
```bash
terraform output alb_url
```

## ⚙️ Configuration

### Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `aws_region` | AWS region | `us-east-1` | No |
| `environment` | Environment name (dev/staging/prod) | `dev` | No |
| `db_password` | RDS MySQL password | - | **Yes** |
| `ec2_instance_type` | EC2 instance type | `t3.micro` | No |
| `ec2_key_name` | SSH key pair name | `""` | No |
| `asg_min_size` | Minimum instances | `1` | No |
| `asg_max_size` | Maximum instances | `4` | No |
| `asg_desired_capacity` | Desired instances | `2` | No |
| `cloudwatch_log_retention_days` | Log retention | `7` | No |

### Network Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `vpc_cidr` | VPC CIDR block | `10.0.0.0/16` |
| `public_subnet_cidrs` | Public subnet CIDRs | `["10.0.1.0/24", "10.0.2.0/24"]` |
| `private_subnet_cidrs` | Private subnet CIDRs | `["10.0.10.0/24", "10.0.11.0/24"]` |

### Database Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `db_instance_class` | RDS instance type | `db.t3.micro` |
| `db_name` | Database name | `meetingslotdb` |
| `db_username` | Database username | `admin` |
| `db_allocated_storage` | Storage in GB | `20` |
| `db_engine_version` | MySQL version | `8.0` |

## 📤 Deployment

### Initial Deployment

```bash
# 1. Initialize
terraform init

# 2. Plan (review changes)
terraform plan -out=tfplan

# 3. Apply
terraform apply tfplan
```

### Application Deployment

After infrastructure is created, deploy your application:

```bash
# 1. SSH into an EC2 instance (via SSM or SSH)
aws ssm start-session --target <instance-id>

# 2. Clone your application code
sudo su - appuser
cd /opt/meeting-slot-service
git clone <your-repo-url> src

# 3. Deploy
sudo /opt/meeting-slot-service/deploy.sh
```

### Update Deployment (Zero Downtime)

```bash
# 1. Update Launch Template (modify code in autoscaling.tf or user_data.sh)
# 2. Apply changes
terraform apply

# 3. Trigger instance refresh (automatic with instance_refresh configuration)
# Or manually:
aws autoscaling start-instance-refresh \
  --auto-scaling-group-name <asg-name> \
  --preferences MinHealthyPercentage=50
```

## 📊 Monitoring and Logs

### CloudWatch Dashboard

Access your custom dashboard:
```bash
terraform output cloudwatch_dashboard_url
```

### View Logs

```bash
# Application logs
aws logs tail /aws/ec2/dev-meeting-slot/application --follow

# Error logs
aws logs tail /aws/ec2/dev-meeting-slot/error --follow

# System logs
aws logs tail /aws/ec2/dev-meeting-slot/system --follow
```

### Key Metrics

- **ALB Metrics**: Request count, response time, HTTP codes
- **EC2 Metrics**: CPU utilization, network in/out
- **RDS Metrics**: CPU, connections, read/write latency
- **Custom Metrics**: Application-specific metrics via CloudWatch Agent

### Alarms

- **CPU High**: Triggers scale-up when CPU > 70%
- **CPU Low**: Triggers scale-down when CPU < 20%
- **High Error Rate**: Alerts when error count > 10 in 5 minutes

## 📈 Scaling

### Auto Scaling Policies

1. **Simple Scaling**:
   - Scale up: +1 instance when CPU > 70% for 2 minutes
   - Scale down: -1 instance when CPU < 20% for 2 minutes
   - Cooldown: 5 minutes

2. **Target Tracking** (Recommended):
   - Target: 50% average CPU utilization
   - Automatically adds/removes instances to maintain target

### Manual Scaling

```bash
# Update desired capacity
aws autoscaling set-desired-capacity \
  --auto-scaling-group-name <asg-name> \
  --desired-capacity 3

# Update min/max sizes
aws autoscaling update-auto-scaling-group \
  --auto-scaling-group-name <asg-name> \
  --min-size 2 \
  --max-size 6
```

### Scaling Configuration

Modify in `terraform.tfvars`:
```hcl
asg_min_size         = 2  # Minimum instances
asg_max_size         = 8  # Maximum instances
asg_desired_capacity = 3  # Desired instances
```

## 🔧 Troubleshooting

### Common Issues

#### 1. Instances Not Healthy

```bash
# Check target group health
aws elbv2 describe-target-health --target-group-arn <tg-arn>

# Check instance system logs
aws ec2 get-console-output --instance-id <instance-id>

# Check user data logs
aws logs tail /aws/ec2/dev-meeting-slot/system --follow --filter-pattern "user-data"
```

#### 2. Application Not Starting

```bash
# SSH to instance and check service status
sudo systemctl status meeting-slot-service

# Check application logs
sudo journalctl -u meeting-slot-service -f

# Check CloudWatch logs
aws logs tail /aws/ec2/dev-meeting-slot/error --follow
```

#### 3. Database Connection Issues

```bash
# Verify security group allows EC2 -> RDS
# Check Secrets Manager secret
aws secretsmanager get-secret-value --secret-id <secret-arn>

# Test from EC2 instance
mysql -h <rds-endpoint> -u admin -p
```

#### 4. High Costs

```bash
# Check running resources
terraform state list

# Terminate unused instances
terraform destroy -target=aws_autoscaling_group.app

# Review CloudWatch costs
aws ce get-cost-and-usage --time-period Start=2024-01-01,End=2024-01-31 \
  --granularity MONTHLY --metrics BlendedCost
```

### Debug Mode

Enable detailed Terraform logging:
```bash
export TF_LOG=DEBUG
terraform apply
```

## 💰 Cost Estimation

### Monthly Cost Breakdown (us-east-1, dev environment)

| Service | Configuration | Approx. Cost |
|---------|---------------|--------------|
| **EC2** | 2x t3.micro (2 vCPU, 1GB RAM) | ~$12 |
| **ALB** | Application Load Balancer | ~$22 |
| **RDS** | db.t3.micro MySQL | ~$15 |
| **NAT Gateway** | 2x NAT Gateway | ~$65 |
| **EBS** | 40GB gp3 (2x 20GB) | ~$4 |
| **CloudWatch** | Logs (7 days retention) | ~$5 |
| **Data Transfer** | Internet egress | ~$5 |
| **S3** | ALB logs | <$1 |
| **Total** | | **~$129/month** |

### Cost Optimization Tips

1. **Use Single NAT Gateway**: Save ~$32/month (reduces HA)
2. **Reduce Instances**: Use `asg_min_size = 1` for dev
3. **Stop in Non-Business Hours**: Use Lambda to stop ASG overnight
4. **Use Reserved Instances**: 40% savings for production
5. **Reduce Log Retention**: Use 3 days instead of 7
6. **Use t3.nano**: ~$3.80/month per instance (for low traffic)

### Production Cost Estimate

| Service | Configuration | Approx. Cost |
|---------|---------------|--------------|
| **EC2** | 3x t3.small (2 vCPU, 2GB RAM) | ~$45 |
| **ALB** | Application Load Balancer | ~$25 |
| **RDS** | db.t3.small MySQL Multi-AZ | ~$60 |
| **NAT Gateway** | 2x NAT Gateway | ~$65 |
| **EBS** | 60GB gp3 (3x 20GB) | ~$6 |
| **CloudWatch** | Logs (30 days retention) | ~$15 |
| **Data Transfer** | Internet egress | ~$20 |
| **Total** | | **~$236/month** |

## 🔐 Security Best Practices

### Implemented Security Features

1. ✅ **Encryption at Rest**: All EBS and RDS volumes encrypted
2. ✅ **Encryption in Transit**: HTTPS ready (requires ACM certificate)
3. ✅ **Private Subnets**: RDS in private subnets only
4. ✅ **Security Groups**: Least privilege access
5. ✅ **Secrets Management**: Database credentials in Secrets Manager
6. ✅ **IAM Roles**: Instance profiles, no hard-coded credentials
7. ✅ **IMDSv2**: Instance metadata service v2 required
8. ✅ **S3 Block Public Access**: ALB logs bucket secured

### Additional Recommendations

1. **Enable MFA Delete** on S3 buckets
2. **Use ACM for SSL/TLS** certificates
3. **Restrict SSH Access** to specific IP ranges
4. **Enable VPC Flow Logs** for network monitoring
5. **Use AWS WAF** for application firewall
6. **Enable GuardDuty** for threat detection
7. **Implement Backup Strategy** for RDS snapshots
8. **Use Parameter Store** for application secrets
9. **Enable CloudTrail** for audit logging
10. **Regular Security Audits** using AWS Security Hub

### SSH Hardening

If you need SSH access:

```bash
# In security_groups.tf, restrict SSH to your IP
ingress {
  description = "SSH from office"
  from_port   = 22
  to_port     = 22
  protocol    = "tcp"
  cidr_blocks = ["203.0.113.0/32"]  # Your public IP
}
```

Or use **AWS Systems Manager Session Manager** (no SSH required):
```bash
aws ssm start-session --target <instance-id>
```

## 📁 File Structure

```
terraform/
├── main.tf                 # Main configuration (provider, data sources)
├── variables.tf           # Input variables
├── outputs.tf             # Output values
├── vpc.tf                 # VPC, subnets, routing
├── security_groups.tf     # Security groups
├── alb.tf                 # Application Load Balancer
├── autoscaling.tf         # Launch Template, ASG, scaling policies
├── ec2.tf                 # IAM roles and instance profile
├── rds.tf                 # RDS MySQL, Secrets Manager
├── cloudwatch.tf          # Logs, metrics, dashboard, alarms
├── templates/
│   └── user_data.sh       # EC2 initialization script
├── terraform.tfvars       # Your variable values (gitignored)
└── README.md              # This file
```

## 🔄 CI/CD Integration

### GitHub Actions Example

```yaml
name: Terraform Deploy

on:
  push:
    branches: [main]

jobs:
  terraform:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      
      - name: Setup Terraform
        uses: hashicorp/setup-terraform@v1
        
      - name: Terraform Init
        run: terraform init
        working-directory: ./terraform
        
      - name: Terraform Plan
        run: terraform plan
        working-directory: ./terraform
        env:
          AWS_ACCESS_KEY_ID: ${{ secrets.AWS_ACCESS_KEY_ID }}
          AWS_SECRET_ACCESS_KEY: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
          TF_VAR_db_password: ${{ secrets.DB_PASSWORD }}
          
      - name: Terraform Apply
        if: github.ref == 'refs/heads/main'
        run: terraform apply -auto-approve
        working-directory: ./terraform
```

## 🧪 Testing

### Validate Terraform Configuration

```bash
# Check syntax
terraform fmt -check

# Validate configuration
terraform validate

# Security scan with tfsec
docker run --rm -v $(pwd):/src aquasec/tfsec /src
```

### Test Deployment

```bash
# Get ALB URL
ALB_URL=$(terraform output -raw alb_url)

# Test health endpoint
curl ${ALB_URL}/health

# Load test
ab -n 1000 -c 10 ${ALB_URL}/health
```

## 📚 Additional Resources

- [Terraform AWS Provider Docs](https://registry.terraform.io/providers/hashicorp/aws/latest/docs)
- [AWS Well-Architected Framework](https://aws.amazon.com/architecture/well-architected/)
- [Terraform Best Practices](https://www.terraform-best-practices.com/)
- [AWS Auto Scaling Guide](https://docs.aws.amazon.com/autoscaling/ec2/userguide/what-is-amazon-ec2-auto-scaling.html)

## 🤝 Support

For issues or questions:
1. Check the [Troubleshooting](#troubleshooting) section
2. Review CloudWatch logs
3. Open an issue in the repository
4. Contact the infrastructure team

## 📝 License

[Your License Here]

---

**Happy Deploying! 🚀**
