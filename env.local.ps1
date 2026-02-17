# Local development environment variables for PowerShell
# Usage: . .\env.local.ps1

$env:DB_HOST = "localhost"
$env:DB_PORT = "3306"
$env:DB_USER = "appuser"
$env:DB_PASSWORD = "password"
$env:DB_NAME = "meetingslots"

$env:SERVER_HOST = "0.0.0.0"
$env:SERVER_PORT = "8080"

$env:ENV = "development"
$env:LOG_LEVEL = "debug"
$env:AWS_REGION = "us-east-1"

Write-Host "Environment variables loaded for local development" -ForegroundColor Green
