# Meeting Slot Service

A REST API service built in Go that helps geographically distributed teams find optimal meeting times by analyzing participant availability and recommending time slots that work for everyone.

## 🎯 Problem Statement

In distributed teams across multiple timezones, finding a common meeting time is challenging:
- Team members span continents (e.g., US, Europe, Asia)
- Manual scheduling via email chains is time-consuming and error-prone
- Timezone conversions lead to mistakes and missed meetings
- Finding slots that work for everyone often requires multiple iterations

---

## Solution

This service solves the meeting scheduling problem by:
1. **Proposing Time Slots** - Organizers propose multiple potential meeting windows
2. **Collecting Availability** - Participants submit their available times in their local timezone
3. **Smart Recommendations** - Algorithm finds optimal times that work for all (or most) participants
4. **Fallback Options** - Provides ranked alternatives when perfect alignment isn't possible

---

## ✨ Features

- **Event Management** - Create, update, and delete meeting events with proposed time slots
- **Availability Tracking** - Participants submit their available time windows
- **Smart Recommendations** - Algorithm calculates best meeting times with availability percentages
- **Timezone Support** - Built-in handling of multiple timezones (all stored/compared in UTC)
- **RESTful API** - Clean, well-documented REST endpoints
- **AWS Native** - Deployed on AWS with ALB, Auto Scaling, RDS, and CloudWatch
- **Horizontal Scalability** - Auto-scaling group (1-4 instances) handles variable load
- **Production Monitoring** - CloudWatch logs, metrics, dashboards, and alarms

---

## 🚀 Quick Start

### Prerequisites

- **Go 1.24+** (with toolchain set to `go1.24.5`)
- **MySQL 8.0+** (local installation or Docker)
- **Docker** (optional, for running MySQL in container)
- **Postman** (optional, for API testing)

### Local Development Setup

#### Option 1: Local MySQL Server (Windows)

If you have MySQL Server installed locally (e.g., MySQL Community Server):

```powershell
# Clone the repository
git clone https://github.com/Pratiksdeshpande/meeting-slot-service.git
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

#### Option 2: Docker MySQL (Windows PowerShell)

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

The server will start on `http://localhost:8080`

---

## Environment Files

| File | Platform | Usage |
|------|----------|-------|
| `env.local.ps1` | Windows PowerShell | `. .\env.local.ps1` |
| `env.local.sh` | Linux/Mac | `source env.local.sh` |

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

---

## 🧪 Testing with Postman

For detailed API testing instructions with a complete step-by-step scenario, see:

📖 **[TEST-APIs.md](TEST-APIs.md)** - Complete API testing guide with example requests and responses

---

## 🚢 AWS Deployment

### Prerequisites
- AWS CLI configured with appropriate credentials
- Terraform 1.0+ installed

### Deploy Infrastructure

```powershell
# Navigate to terraform directory
cd terraform

# Copy and configure variables
cp terraform.tfvars.example terraform.tfvars
# Edit terraform.tfvars with your values

# Set RDS password securely
$env:TF_VAR_db_password = "your-secure-password"

# Initialize Terraform
terraform init

# Plan deployment
terraform plan

# Apply infrastructure
terraform apply

# Get Application Load Balancer URL
terraform output alb_url
```

### Post-Deployment Verification

```powershell
# Get the ALB URL
$ALB_URL = terraform output -raw alb_url

# Health check
curl.exe "$ALB_URL/health"

# Create a test user
curl.exe -X POST "$ALB_URL/api/v1/users" `
  -H "Content-Type: application/json" `
  -d '{"name":"John Doe","email":"john@example.com"}'
```

### Infrastructure Components

| Resource | Description |
|----------|-------------|
| **Application Load Balancer** | Distributes traffic across EC2 instances |
| **Auto Scaling Group** | 1-4 EC2 instances, scales based on CPU |
| **RDS MySQL** | Managed database with Multi-AZ support |
| **CloudWatch** | Logs, metrics, dashboards, and alarms |
| **Secrets Manager** | Secure database credential storage |
| **VPC** | Isolated network with public/private subnets |

For detailed infrastructure documentation, see [terraform/README.md](./terraform/README.md)

---

## 📖 API Documentation

Full API documentation is available in OpenAPI/Swagger format:

📖 **[docs/swagger.yaml](./docs/swagger.yaml)** - OpenAPI 3.0 specification

### API Endpoints Overview

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | Health check |
| `/api/v1/users` | POST, GET | Create/list users |
| `/api/v1/users/{id}` | GET, PUT, DELETE | User operations |
| `/api/v1/events` | POST, GET | Create/list events |
| `/api/v1/events/{id}` | GET, PUT, DELETE | Event operations |
| `/api/v1/events/{id}/participants` | POST, GET | Manage participants |
| `/api/v1/events/{id}/participants/{user_id}` | DELETE | Remove participant |
| `/api/v1/events/{id}/participants/{user_id}/availability` | POST, PUT, GET | Availability operations |
| `/api/v1/events/{id}/recommendations` | GET | Get meeting recommendations |

---

## 🛠️ Technology Stack

| Component | Technology |
|-----------|------------|
| **Language** | Go 1.24+ |
| **Web Framework** | Gorilla Mux |
| **Database** | MySQL 8.0 (AWS RDS) |
| **Database Access** | database/sql + AWS SDK for Go v2 |
| **Infrastructure** | Terraform |
| **Cloud Provider** | AWS (ALB, EC2, RDS, CloudWatch, Secrets Manager) |
| **Testing** | testify, httptest, go-sqlmock |
| **API Documentation** | OpenAPI/Swagger 3.0 |

---

## 🎯 Algorithm Visualization

The **Slot Matching Algorithm** lives in `internal/service/recommendation_service.go` and uses a sliding window approach to find the optimal meeting time across all proposed slots.

### Algorithm Steps

1. **Normalize to UTC** - All times converted for consistent comparison
2. **Generate Candidates** - 15-minute sliding windows within proposed time ranges
3. **Check Availability** - For each candidate, verify overlap with user availability
4. **Calculate Rate** - Percentage of participants available for each slot
5. **Rank Results** - Order by availability rate, then by time
6. **Return Best** - Top recommendation with participant details


### Step-by-Step Walkthrough

```
═══════════════════════════════════════════════════════════════════
  INPUTS
═══════════════════════════════════════════════════════════════════

  Proposed Window:  [════════════════════════]  16:00 → 19:00 UTC
  Meeting Duration: 90 minutes
  Participants:     4 (Sarah, Raj, Emma, Carlos)

═══════════════════════════════════════════════════════════════════
  STEP 1: Generate Candidate Slots (15-min sliding window)
═══════════════════════════════════════════════════════════════════

  16:00 ──────────────────────────────────────────── 19:00
    │                                                   │
    [══════════════]                    16:00 → 17:30  Candidate A
        [══════════════]                16:15 → 17:45  Candidate B
            [══════════════]            16:30 → 18:00  Candidate C
                [══════════════]        16:45 → 18:15  Candidate D
                    [══════════════]    17:00 → 18:30  Candidate E
                        [══════════════]17:15 → 18:45  Candidate F
                            [══════════]17:30 → 19:00  Candidate G

═══════════════════════════════════════════════════════════════════
  STEP 2: Map Participant Availability Windows (UTC)
═══════════════════════════════════════════════════════════════════

  15:00          16:00          17:00          18:00          19:00
    │              │              │              │              │
    Sarah:         [══════════════════════════]  16:00 → 19:00
    Raj:      [═══════════════════════]          15:30 → 18:30
    Emma:  [══════════════════════]              15:00 → 18:00
    Carlos:[════════════════════════════════════]15:00 → 20:00

═══════════════════════════════════════════════════════════════════
  STEP 3: Score Each Candidate
          User is available if candidate window ⊆ user's window
═══════════════════════════════════════════════════════════════════

  Candidate        Sarah    Raj     Emma   Carlos    Score
  ─────────────────────────────────────────────────────────
  A: 16:00–17:30     ✓       ✓       ✓       ✓      4/4 = 100% ★ BEST
  B: 16:15–17:45     ✓       ✓       ✓       ✓      4/4 = 100%
  C: 16:30–18:00     ✓       ✓       ✓       ✓      4/4 = 100%
  D: 16:45–18:15     ✓       ✓       ✗       ✓      3/4 =  75%
  E: 17:00–18:30     ✓       ✓       ✗       ✓      3/4 =  75%
  F: 17:15–18:45     ✓       ✗       ✗       ✓      2/4 =  50%
  G: 17:30–19:00     ✓       ✗       ✗       ✓      2/4 =  50%

═══════════════════════════════════════════════════════════════════
  STEP 4: Rank & Return Best
          Sort by: availability rate DESC → time ASC
═══════════════════════════════════════════════════════════════════

  ★ WINNER: Candidate A — 16:00–17:30 UTC
    available_participants: 4
    availability_rate:      1.0  (100%)
    available_users:        [sarah, raj, emma, carlos]
    unavailable_users:      []

═══════════════════════════════════════════════════════════════════
  COMPLEXITY
═══════════════════════════════════════════════════════════════════

  O(P × C × U × S)
    P = proposed slots      (~3–10)
    C = candidates/slot     (fixed: window_minutes / 15 - duration/15 + 1)
    U = participants        (~4–20)
    S = avail. slots/user   (~1–5)

  Example: 3 slots × 7 candidates × 4 users × 3 slots/user = 252 ops
  Runs in microseconds even for large teams
```

---

## Documentation

| Document                                     | Description                                      |
|----------------------------------------------|--------------------------------------------------|
| [ARCHITECTURE.md](./ARCHITECTURE.md)         | System architecture, data flow, component design |
| [TEST-APIs.md](TEST-APIs.md)                 | Step-by-step API testing guide                   |
| [docs/swagger.yaml](./docs/swagger.yaml)     | OpenAPI specification                            |

---

