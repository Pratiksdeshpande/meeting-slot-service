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
dnf install -y golang git jq

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
StandardOutput=journal
StandardError=journal

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
