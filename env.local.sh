#!/bin/bash
# Local development environment variables
# Usage: source env.local.sh

export DB_HOST=localhost
export DB_PORT=3306
export DB_USER=appuser
export DB_PASSWORD=password
export DB_NAME=meetingslots

export SERVER_HOST=0.0.0.0
export SERVER_PORT=8080

export ENV=development
export LOG_LEVEL=debug
export AWS_REGION=us-east-1

echo "Environment variables loaded for local development"
