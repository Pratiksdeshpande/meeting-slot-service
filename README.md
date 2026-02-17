# Meeting Slot Service

A REST API service built in Go that helps geographically distributed teams find optimal meeting times by analyzing participant availability and recommending time slots that work for everyone.

## 🎯 Problem Statement

In distributed teams across multiple timezones, finding a common meeting time is challenging. This service solves that by:
- Allowing organizers to propose multiple time slots
- Collecting availability from all participants
- Recommending optimal times that work for all (or most) participants
- Providing fallback options when perfect alignment isn't possible

## ✨ Features

- **Event Management**: Create, update, and delete meeting events
- **Availability Tracking**: Participants submit their available time slots
- **Smart Recommendations**: Algorithm finds best meeting times considering all constraints
- **Timezone Support**: Built-in handling of multiple timezones
- **RESTful API**: Clean, well-documented REST endpoints
- **AWS Native**: Deployed on AWS with EC2, RDS, and API Gateway

## 🏗️ Architecture

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   AWS API       │────▶│     EC2         │────▶│    AWS RDS      │
│   Gateway       │     │  (Go Service)   │     │    (MySQL)      │
└─────────────────┘     └─────────────────┘     └─────────────────┘
                               │
                               ▼
                        ┌─────────────────┐
                        │ AWS Secrets     │
                        │ Manager         │
                        └─────────────────┘
```

## 🚀 Quick Start

### Prerequisites
- Go 1.24+ (with toolchain set to go1.24.5)
- MySQL 8.0+ (local installation or Docker)
- Docker (optional, for running MySQL in container)
- AWS CLI configured (for production deployment)
- Terraform 1.0+ (for AWS infrastructure)
- Postman (optional, for API testing)

### Local Development Setup

#### Option 1: Local MySQL Server (Windows)

If you have MySQL Server installed locally (e.g., MySQL Community Server):

```powershell
# Clone the repository
git clone https://github.com/yourusername/meeting-slot-service.git
cd meeting-slot-service

# Install dependencies
go mod download

# Connect to MySQL as root and set up the database
mysql -u root -p
```

Run these SQL commands in the MySQL prompt:
```sql
CREATE DATABASE IF NOT EXISTS meetingslots;
CREATE USER IF NOT EXISTS 'appuser'@'localhost' IDENTIFIED BY 'password';
GRANT ALL PRIVILEGES ON meetingslots.* TO 'appuser'@'localhost';
FLUSH PRIVILEGES;
EXIT;
```

Then start the application:
```powershell
# Load environment variables
. .\env.local.ps1

# Run the service
go run cmd/server/main.go
```

#### Option 2: Using Make with Docker (Linux/Mac)

```bash
# Clone the repository
git clone https://github.com/yourusername/meeting-slot-service.git
cd meeting-slot-service

# Install dependencies
make init

# Start MySQL in Docker
make docker-mysql

# Run the service with environment variables loaded
make run-local

# When done, stop MySQL
make docker-mysql-stop
```

#### Option 3: Docker MySQL (Windows PowerShell)

```powershell
# Clone the repository
git clone https://github.com/yourusername/meeting-slot-service.git
cd meeting-slot-service

# Install dependencies
go mod download

# Start MySQL in Docker
docker run --name mysql-meeting `
  -e MYSQL_ROOT_PASSWORD=password `
  -e MYSQL_DATABASE=meetingslots `
  -e MYSQL_USER=appuser `
  -e MYSQL_PASSWORD=password `
  -p 3306:3306 `
  -d mysql:8.0

# Wait ~30 seconds for MySQL to initialize

# Load environment variables
. .\env.local.ps1

# Run the service
go run cmd/server/main.go

# When done, stop MySQL
docker stop mysql-meeting; docker rm mysql-meeting
```

#### Option 4: Docker MySQL (Linux/Mac Bash)

```bash
# Clone the repository
git clone https://github.com/yourusername/meeting-slot-service.git
cd meeting-slot-service

# Install dependencies
go mod download

# Start MySQL in Docker
docker run --name mysql-meeting \
  -e MYSQL_ROOT_PASSWORD=password \
  -e MYSQL_DATABASE=meetingslots \
  -e MYSQL_USER=appuser \
  -e MYSQL_PASSWORD=password \
  -p 3306:3306 \
  -d mysql:8.0

# Wait ~30 seconds for MySQL to initialize

# Load environment variables
source env.local.sh

# Run the service
go run cmd/server/main.go

# When done, stop MySQL
docker stop mysql-meeting && docker rm mysql-meeting
```

The server will start on `http://localhost:8080`

### Make Commands Reference

| Command | Description |
|---------|-------------|
| `make help` | Show all available commands |
| `make init` | Download Go dependencies |
| `make docker-mysql` | Start MySQL in Docker container |
| `make docker-mysql-stop` | Stop and remove MySQL container |
| `make docker-mysql-logs` | View MySQL container logs |
| `make run` | Run the application |
| `make run-local` | Load env.local.sh and run (Linux/Mac) |
| `make build` | Build binary to `bin/server` |
| `make test` | Run all tests |
| `make test-coverage` | Run tests with coverage report |
| `make clean` | Clean build artifacts |

### Environment Files

| File | Platform | Usage |
|------|----------|-------|
| `env.local.sh` | Linux/Mac | `source env.local.sh` |
| `env.local.ps1` | Windows | `. .\env.local.ps1` |

### Testing with Postman

#### Import API Collection
1. Open Postman
2. Click **Import** → Select `docs/swagger.yaml`
3. This creates a collection with all 16 endpoints

#### Sample API Flow

**1. Health Check:**
```
GET http://localhost:8080/health
```

**2. Create a User:**
```
POST http://localhost:8080/api/v1/users
Content-Type: application/json

{
  "name": "John Doe",
  "email": "john@example.com"
}
```
Save the returned `id` (e.g., `usr_abc123`)

**3. Create an Event:**
```
POST http://localhost:8080/api/v1/events
Content-Type: application/json

{
  "title": "Team Standup",
  "description": "Daily sync meeting",
  "organizer_id": "usr_abc123",
  "duration_minutes": 30,
  "proposed_slots": [
    {
      "start_time": "2026-02-18T09:00:00Z",
      "end_time": "2026-02-18T10:00:00Z",
      "timezone": "UTC"
    },
    {
      "start_time": "2026-02-18T14:00:00Z",
      "end_time": "2026-02-18T15:00:00Z",
      "timezone": "UTC"
    }
  ]
}
```
Save the returned `id` (e.g., `evt_xyz789`)

**4. Add a Participant:**
```
POST http://localhost:8080/api/v1/events/evt_xyz789/participants
Content-Type: application/json

{
  "user_id": "usr_abc123"
}
```

**5. Submit Availability:**
```
POST http://localhost:8080/api/v1/events/evt_xyz789/participants/usr_abc123/availability
Content-Type: application/json

{
  "slots": [
    {
      "start_time": "2026-02-18T09:00:00Z",
      "end_time": "2026-02-18T10:00:00Z",
      "timezone": "UTC"
    }
  ]
}
```

**6. Get Recommendations:**
```
GET http://localhost:8080/api/v1/events/evt_xyz789/recommendations
```

### AWS Deployment

```bash
# Navigate to terraform directory
cd terraform

# Copy and configure variables
cp terraform.tfvars.example terraform.tfvars
# Edit terraform.tfvars with your values

# Initialize Terraform
terraform init

# Plan deployment
terraform plan

# Apply infrastructure
terraform apply

# Get outputs
terraform output
```

### Testing the API

```bash
# Create a user
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"name": "John Doe", "email": "john@example.com"}'

# Create an event (use the user ID from above response)
curl -X POST http://localhost:8080/api/v1/events \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Team Meeting",
    "organizer_id": "usr_...",
    "duration_minutes": 60,
    "proposed_slots": [{
      "start_time": "2025-02-20T14:00:00Z",
      "end_time": "2025-02-20T16:00:00Z",
      "timezone": "UTC"
    }]
  }'

# Get recommendations
curl http://localhost:8080/api/v1/events/evt_.../recommendations
```

## 🔗 API Documentation

Full API documentation is available in OpenAPI/Swagger format:
- **Swagger File**: [docs/swagger.yaml](./docs/swagger.yaml)

### Core Endpoints

```
# Users
POST   /api/v1/users                              # Create user
GET    /api/v1/users/{id}                         # Get user
PUT    /api/v1/users/{id}                         # Update user
DELETE /api/v1/users/{id}                         # Delete user

# Events
POST   /api/v1/events                             # Create event
GET    /api/v1/events                             # List events
GET    /api/v1/events/{id}                        # Get event details
PUT    /api/v1/events/{id}                        # Update event
DELETE /api/v1/events/{id}                        # Delete event

# Participants  
POST   /api/v1/events/{id}/participants           # Add participant
GET    /api/v1/events/{id}/participants           # List participants
DELETE /api/v1/events/{id}/participants/{user_id} # Remove participant

# Availability
POST   /api/v1/events/{id}/participants/{user_id}/availability  # Submit
PUT    /api/v1/events/{id}/participants/{user_id}/availability  # Update
GET    /api/v1/events/{id}/participants/{user_id}/availability  # Get

# Recommendations
GET    /api/v1/events/{id}/recommendations        # Get recommended slots

# Health
GET    /health                                    # Health check
```

See [ARCHITECTURE.md](./ARCHITECTURE.md) for system architecture and design decisions.

## 🛠️ Technology Stack

- **Language**: Go 1.25+
- **Web Framework**: Gorilla Mux
- **Database**: AWS RDS MySQL
- **Database Access**: AWS SDK for Go v2, database/sql
- **Configuration**: Environment variables + AWS Secrets Manager
- **Infrastructure**: Terraform
- **Cloud**: AWS (EC2, RDS, API Gateway, Secrets Manager, VPC)
- **Testing**: testify, httptest

## ☁️ AWS Infrastructure

The Terraform configuration provisions:

| Resource | Description |
|----------|-------------|
| **VPC** | Isolated network with public/private subnets |
| **EC2** | Application server running the Go service |
| **RDS MySQL** | Managed database in private subnet |
| **API Gateway** | Public REST API endpoint with CORS |
| **Secrets Manager** | Database credential storage |
| **IAM** | Roles and policies for EC2 |
| **NAT Gateway** | Outbound internet for private subnets |

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVER_HOST` | Server bind address | `0.0.0.0` |
| `SERVER_PORT` | Server port | `8080` |
| `DB_HOST` | Database hostname | `localhost` |
| `DB_PORT` | Database port | `3306` |
| `DB_USER` | Database username | - |
| `DB_PASSWORD` | Database password | - |
| `DB_NAME` | Database name | `meetingslots` |
| `DB_SECRET_ARN` | AWS Secrets Manager ARN | - |
| `AWS_REGION` | AWS region | `us-east-1` |
| `ENV` | Environment | `development` |

## 📊 Project Status

**Current Phase**: ✅ Core + AWS Infrastructure Complete

Completed:
- ✅ Project setup and configuration
- ✅ Database models and migrations (MySQL)
- ✅ Repository layer with AWS SDK
- ✅ Service layer with business logic
- ✅ **Core recommendation algorithm**
- ✅ HTTP handlers with Gorilla Mux
- ✅ Middleware (CORS, logging, recovery)
- ✅ Unit tests for critical components
- ✅ RESTful API with all CRUD operations
- ✅ **Terraform infrastructure (VPC, EC2, RDS, API Gateway)**
- ✅ **AWS Secrets Manager integration**
- ✅ **Swagger/OpenAPI documentation**

Pending:
- ⏳ GitLab CI/CD pipeline setup
- ⏳ Integration tests
- ⏳ Production hardening



## 🧪 Testing

```bash
# Run all tests
make test
# or: go test -v ./...

# Run tests with coverage
make test-coverage
# or: go test -v -coverprofile=coverage.out ./...

# Run specific package tests
go test -v ./internal/service/...
go test -v ./internal/utils/...

# Run specific test
go test -v -run TestRecommendationService_AllParticipantsAvailable ./internal/service/
```

## 🚢 Deployment

### Local Development
Run with local MySQL following the Quick Start guide above.

### AWS Deployment with Terraform

```bash
cd terraform

# Configure variables
cp terraform.tfvars.example terraform.tfvars
vim terraform.tfvars  # Edit with your values

# Set RDS password securely
export TF_VAR_rds_password="your-secure-password"

# Deploy infrastructure
terraform init
terraform plan
terraform apply

# Get API Gateway URL
terraform output api_gateway_url
```

### Post-Deployment
1. Deploy your Go binary to EC2 via the deployment script
2. Verify health check: `curl <api_gateway_url>/health`
3. Access CloudWatch for API Gateway logs

## 📖 Documentation

- **[ARCHITECTURE.md](./ARCHITECTURE.md)** - System architecture and design decisions
- **[docs/swagger.yaml](./docs/swagger.yaml)** - OpenAPI/Swagger API documentation

## 🏗 Project Structure

```
meeting-slot-service/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── docs/
│   └── swagger.yaml             # OpenAPI/Swagger documentation
├── internal/
│   ├── config/                  # Configuration management
│   │   └── config.go
│   ├── database/                # AWS SDK database connection
│   │   └── database.go
│   ├── handler/                 # HTTP request handlers
│   │   ├── user_handler.go
│   │   ├── event_handler.go
│   │   └── availability_handler.go
│   ├── middleware/              # HTTP middleware
│   │   ├── logger.go
│   │   ├── recovery.go
│   │   └── cors.go
│   ├── models/                  # Data models
│   │   ├── user.go
│   │   ├── event.go
│   │   ├── slot.go
│   │   ├── participant.go
│   │   └── recommendation.go
│   ├── repository/              # Data access layer (AWS SDK)
│   │   ├── interface.go
│   │   ├── user_repository.go
│   │   ├── event_repository.go
│   │   ├── availability_repository.go
│   │   └── participant_repository.go
│   ├── service/                 # Business logic
│   │   ├── user_service.go
│   │   ├── event_service.go
│   │   ├── availability_service.go
│   │   └── recommendation_service.go
│   └── utils/                   # Utility functions
│       ├── id_generator.go
│       ├── time_utils.go
│       └── response.go
├── terraform/                   # AWS Infrastructure as Code
│   ├── main.tf                  # Provider and locals
│   ├── variables.tf             # Input variables
│   ├── outputs.tf               # Output values
│   ├── vpc.tf                   # VPC and networking
│   ├── security_groups.tf       # Security groups
│   ├── rds.tf                   # RDS MySQL + Secrets Manager
│   ├── ec2.tf                   # EC2 instance + IAM
│   ├── api_gateway.tf           # API Gateway REST API
│   ├── terraform.tfvars.example # Variable template
│   └── templates/
│       └── user_data.sh         # EC2 bootstrap script
├── go.mod                       # Go dependencies
├── Makefile                     # Build automation
├── ARCHITECTURE.md              # Architecture documentation
└── README.md
```

## 🧮 Core Algorithm

The recommendation service implements a sliding window algorithm to find optimal meeting slots:

### Algorithm Steps
1. **Normalize times to UTC** - All times converted for consistent comparison
2. **Generate candidate slots** - 15-minute sliding windows within proposed time ranges
3. **Check participant availability** - For each candidate, verify overlap with user availability
4. **Calculate availability rate** - Percentage of participants available for each slot
5. **Sort and rank** - Order by availability rate, participant count, and time
6. **Return top recommendations** - Up to 10 best options with full participant details

### Performance
- **Time Complexity**: O(P × C × U × S)
  - P = proposed slots (~10)
  - C = candidates per slot (~8, constant)
  - U = participants (~20)
  - S = availability slots per user (~5)
- **Typical Performance**: ~1,000 operations for standard use case
- **Optimizations**: UTC normalization, efficient interval checking

See [ARCHITECTURE.md](./ARCHITECTURE.md) for detailed algorithm explanation and system design.

## 🎓 Key Technical Highlights

This project demonstrates:
- ✅ **Clean Architecture** - Separation of concerns (handler → service → repository)
- ✅ **REST API Design** - Following REST conventions and best practices
- ✅ **Complex Algorithm** - Slot matching with timezone handling
- ✅ **Database Design** - Normalized schema with relationships and constraints
- ✅ **AWS SDK for Go v2** - Native AWS integration for RDS and Secrets Manager
- ✅ **Infrastructure as Code** - Complete AWS infrastructure with Terraform
- ✅ **Dependency Injection** - Testable, loosely coupled components
- ✅ **Error Handling** - Consistent error responses and logging
- ✅ **Middleware Pattern** - CORS, logging, panic recovery
- ✅ **Unit Testing** - Mocks, table-driven tests, test coverage
- ✅ **12-Factor App** - Environment-based configuration
- ✅ **API Documentation** - OpenAPI/Swagger specification

## 🤝 Contributing

This is a personal coding exercise project, but suggestions and feedback are welcome!

## 📝 License

MIT License - feel free to use this as a reference for your own projects.

---

**Status**: Core + AWS Infrastructure complete ✅ | Next: CI/CD pipeline and production hardening