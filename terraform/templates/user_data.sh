#!/bin/bash
set -e

# Log output to file
exec > >(tee /var/log/user-data.log|logger -t user-data -s 2>/dev/console) 2>&1

echo "=== Meeting Slot Service EC2 Setup ==="
echo "Region: ${aws_region}"
echo "Environment: ${environment}"
echo "App Port: ${app_port}"

# Update system
echo "Updating system packages..."
dnf update -y

# Install required packages
echo "Installing required packages..."
dnf install -y golang git jq amazon-cloudwatch-agent

# Create application user
echo "Creating application user..."
useradd -m -s /bin/bash appuser || true

# Create application directory
echo "Setting up application directory..."
mkdir -p /opt/meeting-slot-service
chown appuser:appuser /opt/meeting-slot-service

# Create environment file
echo "Creating environment configuration..."
cat > /opt/meeting-slot-service/.env << 'EOF'
AWS_REGION=${aws_region}
DB_SECRET_ARN=${db_secret_arn}
SERVER_PORT=${app_port}
ENV=${environment}
EOF

chown appuser:appuser /opt/meeting-slot-service/.env

# Configure CloudWatch Agent
echo "Configuring CloudWatch Agent..."
cat > /opt/aws/amazon-cloudwatch-agent/etc/config.json << 'CW_EOF'
{
  "agent": {
    "metrics_collection_interval": 60,
    "run_as_user": "root"
  },
  "logs": {
    "logs_collected": {
      "files": {
        "collect_list": [
          {
            "file_path": "/var/log/meeting-slot-service/app.log",
            "log_group_name": "/aws/ec2/${name_prefix}/application",
            "log_stream_name": "{instance_id}",
            "retention_in_days": 7,
            "timezone": "UTC"
          },
          {
            "file_path": "/var/log/meeting-slot-service/error.log",
            "log_group_name": "/aws/ec2/${name_prefix}/error",
            "log_stream_name": "{instance_id}",
            "retention_in_days": 7,
            "timezone": "UTC"
          },
          {
            "file_path": "/var/log/messages",
            "log_group_name": "/aws/ec2/${name_prefix}/system",
            "log_stream_name": "{instance_id}",
            "retention_in_days": 7,
            "timezone": "UTC"
          },
          {
            "file_path": "/var/log/user-data.log",
            "log_group_name": "/aws/ec2/${name_prefix}/system",
            "log_stream_name": "{instance_id}-user-data",
            "retention_in_days": 7,
            "timezone": "UTC"
          }
        ]
      }
    }
  },
  "metrics": {
    "namespace": "${name_prefix}/Application",
    "metrics_collected": {
      "cpu": {
        "measurement": [
          {
            "name": "cpu_usage_idle",
            "rename": "CPU_IDLE",
            "unit": "Percent"
          },
          {
            "name": "cpu_usage_iowait",
            "rename": "CPU_IOWAIT",
            "unit": "Percent"
          }
        ],
        "metrics_collection_interval": 60,
        "totalcpu": false
      },
      "disk": {
        "measurement": [
          {
            "name": "used_percent",
            "rename": "DISK_USED",
            "unit": "Percent"
          }
        ],
        "metrics_collection_interval": 60,
        "resources": [
          "*"
        ]
      },
      "mem": {
        "measurement": [
          {
            "name": "mem_used_percent",
            "rename": "MEM_USED",
            "unit": "Percent"
          }
        ],
        "metrics_collection_interval": 60
      },
      "netstat": {
        "measurement": [
          {
            "name": "tcp_established",
            "rename": "TCP_ESTABLISHED",
            "unit": "Count"
          }
        ],
        "metrics_collection_interval": 60
      }
    }
  }
}
CW_EOF

# Create log directory for application
mkdir -p /var/log/meeting-slot-service
chown appuser:appuser /var/log/meeting-slot-service

# Start CloudWatch Agent
echo "Starting CloudWatch Agent..."
/opt/aws/amazon-cloudwatch-agent/bin/amazon-cloudwatch-agent-ctl \
  -a fetch-config \
  -m ec2 \
  -s \
  -c file:/opt/aws/amazon-cloudwatch-agent/etc/config.json

# Create systemd service
echo "Creating systemd service..."
cat > /etc/systemd/system/meeting-slot-service.service << 'EOF'
[Unit]
Description=Meeting Slot Service API
After=network.target

[Service]
Type=simple
User=appuser
Group=appuser
WorkingDirectory=/opt/meeting-slot-service
EnvironmentFile=/opt/meeting-slot-service/.env
ExecStart=/opt/meeting-slot-service/server
Restart=always
RestartSec=5
StandardOutput=append:/var/log/meeting-slot-service/app.log
StandardError=append:/var/log/meeting-slot-service/error.log

[Install]
WantedBy=multi-user.target
EOF

# Reload systemd
systemctl daemon-reload

# Create deployment script
echo "Creating deployment script..."
cat > /opt/meeting-slot-service/deploy.sh << 'DEPLOY_EOF'
#!/bin/bash
set -e

echo "Deploying Meeting Slot Service..."

# Stop service if running
systemctl stop meeting-slot-service || true

# Build application (assumes code is in /opt/meeting-slot-service/src)
cd /opt/meeting-slot-service/src
go build -o /opt/meeting-slot-service/server ./cmd/server/main.go

# Set permissions
chown appuser:appuser /opt/meeting-slot-service/server
chmod +x /opt/meeting-slot-service/server

# Start service
systemctl start meeting-slot-service
systemctl enable meeting-slot-service

echo "Deployment complete!"
systemctl status meeting-slot-service
DEPLOY_EOF

chmod +x /opt/meeting-slot-service/deploy.sh

# Create health check script
cat > /opt/meeting-slot-service/health-check.sh << 'HEALTH_EOF'
#!/bin/bash
curl -sf http://localhost:${app_port}/health || exit 1
HEALTH_EOF

chmod +x /opt/meeting-slot-service/health-check.sh

echo "=== EC2 Setup Complete ==="
echo "To deploy the application:"
echo "1. Clone/copy your code to /opt/meeting-slot-service/src"
echo "2. Run: /opt/meeting-slot-service/deploy.sh"
