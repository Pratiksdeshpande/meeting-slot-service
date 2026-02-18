# Architecture Overview

Visual representation of the meeting slot service architecture and data flow.

## 🏛️ System Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         Client Layer                            │
│  (Web Apps, Mobile Apps, CLI Tools, Postman, etc.)              │
└────────────────────────────┬────────────────────────────────────┘
                             │ HTTP/HTTPS
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│          AWS Application Load Balancer (ALB)                    │
│  • Health Checks (/health every 30s)                            │
│  • SSL/TLS Termination (optional)                               │
│  • Cross-zone Load Balancing                                    │
└────────────────────────────┬────────────────────────────────────┘
                             │
         ┌───────────────────┼───────────────────┐
         │                   │                   │
         ▼                   ▼                   ▼
┌──────────────┐    ┌──────────────┐    ┌──────────────┐
│ EC2 Instance │    │ EC2 Instance │    │ EC2 Instance │
│  (Go App)    │    │  (Go App)    │    │  (Go App)    │
└──────────────┘    └──────────────┘    └──────────────┘
       │                   │                   │
       └───────────────────┼───────────────────┘
                           │
         Auto Scaling Group (1-4 instances)
         • Scale up: CPU > 70%
         • Scale down: CPU < 20%
                           │
┌──────────────────────────┼─────────────────────────────────────┐
│     Meeting Slot Service │(Go) - Application Container         │
│  ┌───────────────────────┼───────────────────────────────┐     │
│  │          HTTP Handler │Layer                          │     │
│  │  • Event Handlers  • Availability Handlers            │     │
│  │  • Health Check    • Middleware (CORS, Logging)       │     │
│  └───────────────────────┬───────────────────────────────┘     │
│                          │                                     │
│  ┌───────────────────────▼───────────────────────────────┐     │
│  │                   Service Layer                       │     │
│  │  • Event Service   • Availability Service             │     │
│  │  • Recommendation Service (Algorithm)                 │     │
│  └───────────────────────┬───────────────────────────────┘     │
│                          │                                     │
│  ┌───────────────────────▼───────────────────────────────┐     │
│  │                 Repository Layer                      │     │
│  │  • Event Repository  • User Repository                │     │
│  │  • Availability Repository                            │     │
│  └───────────────────────┬───────────────────────────────┘     │
└──────────────────────────┼─────────────────────────────────────┘
                           │
            ┌──────────────┼──────────────┐
            │              │              │
            ▼              ▼              ▼
    ┌──────────────┐ ┌──────────────┐ ┌──────────────┐
    │  AWS RDS     │ │ AWS Secrets  │ │ CloudWatch   │
    │  MySQL 8.0   │ │  Manager     │ │ • Logs       │
    │  (Multi-AZ)  │ │ (DB Creds)   │ │ • Metrics    │
    │              │ │              │ │ • Dashboard  │
    └──────────────┘ └──────────────┘ │ • Alarms     │
                                      └──────────────┘
```

## 🔄 Request Flow

### 1. Create Event Flow

```
Client                  API                Service              Repository         Database
  │                      │                    │                     │                 │
  │ POST /events         │                    │                     │                 │
  ├──────────────────────>                    │                     │                 │
  │                      │ CreateEvent()      │                     │                 │
  │                      ├────────────────────>                     │                 │
  │                      │                    │ event.Create()      │                 │
  │                      │                    ├─────────────────────>                 │
  │                      │                    │                     │ INSERT          │
  │                      │                    │                     ├─────────────────>
  │                      │                    │                     │<─────────────────
  │                      │                    │<─────────────────────                 │
  │                      │<────────────────────                     │                 │
  │ 201 Created          │                    │                     │                 │
  │<──────────────────────                    │                     │                 │
  │ {event_id: "evt_123"}│                    │                     │                 │
```

### 2. Submit Availability Flow

```
Client                  API                Service              Repository         Database
  │                      │                    │                     │                 │
  │ POST /events/{id}/   │                    │                     │                 │
  │   availability       │                    │                     │                 │
  ├──────────────────────>                    │                     │                 │
  │                      │ SubmitAvailability()                     │                 │
  │                      ├────────────────────>                     │                 │
  │                      │                    │ ValidateTimeSlots() │                 │
  │                      │                    │ NormalizeTimezones()│                 │
  │                      │                    │ availability.Create()                 │
  │                      │                    ├─────────────────────>                 │
  │                      │                    │                     │ INSERT          │
  │                      │                    │                     ├─────────────────>
  │                      │                    │                     │<─────────────────
  │                      │                    │ InvalidateCache()   │                 │
  │                      │                    │<─────────────────────                 │
  │                      │<────────────────────                     │                 │
  │ 200 OK               │                    │                     │                 │
  │<──────────────────────                    │                     │                 │
```

### 3. Get Recommendations Flow

```
Client         API            Service         Repository       Database
  │             │                 │              │                 │               │
  │ GET /events/ │                 │              │                 │               │
  │  {id}/recom- │                 │              │                 │               │
  │  mendations  │                 │              │                 │               │
  ├──────────────>                 │              │                 │               │
  │             │ GetRecommendations()            │                 │               │
  │             ├─────────────────>               │                 │               │
  │             │                 │ LoadEventData()                 │               │
  │             │                 ├─────────────────────────────────>               │
  │             │                 │                 │               │ SELECT        │
  │             │                 │                 │               ├───────────────>
  │             │                 │                 │               │<───────────────
  │             │                 │<─────────────────────────────────               │
  │             │                 │ LoadAvailabilities()            │               │
  │             │                 ├─────────────────────────────────>               │
  │             │                 │                 │               │ SELECT        │
  │             │                 │                 │               ├───────────────>
  │             │                 │                 │               │<───────────────
  │             │                 │<─────────────────────────────────               │
  │             │                 │ RunAlgorithm() │                │               │
  │             │<─────────────────                │                │               │
  │ 200 OK      │                 │                │                │               │
  │<────────────                  │                │                │               │
  │ {recommendations}             │                │                │               │
```

## 🧩 Component Breakdown

### 1. HTTP Handler Layer
**Responsibility**: HTTP request/response handling, validation, routing

```go
// Handles:
- Request parsing and validation
- Response formatting
- HTTP status codes
- Middleware execution (logging, CORS, recovery)
- Error handling
```

**Key Files**:
- `internal/handler/event_handler.go`
- `internal/handler/availability_handler.go`

### 2. Service Layer
**Responsibility**: Business logic, orchestration, algorithm

```go
// Handles:
- Business rules validation
- Multi-repository coordination
- Algorithm execution
- Transaction management
```

**Key Files**:
- `internal/service/event_service.go`
- `internal/service/availability_service.go`
- `internal/service/recommendation_service.go`

### 3. Repository Layer
**Responsibility**: Data access, database operations

```go
// Handles:
- CRUD operations
- Query building
- Database transactions
- Data mapping (DB <-> Models)
```

**Key Files**:
- `internal/repository/event_repo.go`
- `internal/repository/availability_repo.go`
- `internal/repository/user_repo.go`

### 4. Algorithm Package
**Responsibility**: Core slot matching logic

```go
// Handles:
- Time slot overlap detection
- Candidate generation
- Availability calculation
- Recommendation ranking
```

**Key Files**:
- `pkg/algorithm/slot_matcher.go`
- `pkg/algorithm/interval_tree.go`

## 📊 Data Model

### Entity Relationship Diagram

```
┌─────────────┐
│    Users    │
│─────────────│
│ id (PK)     │
│ name        │
│ email       │◄────────┐
│ created_at  │         │
└─────────────┘         │
                        │
                        │ organizer_id
                        │
                   ┌────┴──────────┐
                   │    Events     │
                   │───────────────│
                   │ id (PK)       │
                   │ title         │
                   │ organizer_id  │
                   │ duration_min  │
                   │ status        │
                   │ created_at    │
                   └───┬───────────┘
                       │
         ┌─────────────┼─────────────┐
         │             │             │
         │             │             │
    ┌────▼──────┐ ┌───▼────────┐ ┌──▼────────────┐
    │ Proposed  │ │   Event    │ │ Availability  │
    │   Slots   │ │Participants│ │    Slots      │
    │───────────│ │────────────│ │───────────────│
    │ id (PK)   │ │ id (PK)    │ │ id (PK)       │
    │ event_id  │ │ event_id   │ │ event_id      │
    │ start_time│ │ user_id    │ │ user_id       │
    │ end_time  │ │ status     │ │ start_time    │
    │ timezone  │ └────────────┘ │ end_time      │
    └───────────┘                │ timezone      │
                                 └───────────────┘
```

### Data Flow Example

**Scenario**: Finding best meeting time for 3 people

```
1. Event Created:
   ┌────────────────────────┐
   │ Event: "Q1 Planning"   │
   │ Duration: 60 min       │
   │ Proposed Slots:        │
   │  - Jan 12, 2-4PM EST   │
   │  - Jan 14, 6-8PM EST   │
   └────────────────────────┘

2. Users Submit Availability:
   ┌──────────────────────────────────────────────────────┐
   │ User A: Jan 12 2-3:30PM EST, Jan 14 6-8PM EST       │
   │ User B: Jan 12 2-4PM EST, Jan 14 7-8PM EST          │
   │ User C: Jan 12 3-4PM EST, Jan 14 6-7:30PM EST       │
   └──────────────────────────────────────────────────────┘

3. Algorithm Processes:
   ┌────────────────────────────────────────────┐
   │ Normalize to UTC                           │
   │ Generate candidates (15-min intervals)     │
   │ Check each candidate against all users     │
   │ Calculate availability rates               │
   │ Sort by best match                         │
   └────────────────────────────────────────────┘

4. Recommendations Returned:
   ┌─────────────────────────────────────────────────┐
   │ 1. Jan 14, 6-7PM EST (100% - all available)    │
   │ 2. Jan 12, 3-4PM EST (66% - A, C available)    │
   │ 3. Jan 12, 2-3PM EST (66% - A, B available)    │
   └─────────────────────────────────────────────────┘
```

## 🎯 Algorithm Visualization

### Slot Matching Algorithm

```
Input:
  Proposed Window: ═══════════════════════════  (2:00 PM - 4:00 PM)
  Duration Needed:     ══════ (60 minutes)
  
Candidate Generation (15-min sliding window):
  Candidate 1:      ══════                       (2:00 PM - 3:00 PM)
  Candidate 2:         ══════                    (2:15 PM - 3:15 PM)
  Candidate 3:            ══════                 (2:30 PM - 3:30 PM)
  Candidate 4:               ══════              (2:45 PM - 3:45 PM)
  Candidate 5:                  ══════           (3:00 PM - 4:00 PM)

User Availability Check:
  User A:           ════════════                 (2:00 PM - 3:30 PM)
  User B:           ════════════════════         (2:00 PM - 4:00 PM)
  User C:                      ════════          (3:00 PM - 4:00 PM)

Overlap Analysis:
  Candidate 1:      ✓ ✓ ✗  (66% available)
  Candidate 2:      ✓ ✓ ✗  (66% available)
  Candidate 3:      ✓ ✓ ✗  (66% available)
  Candidate 4:      ✗ ✓ ✓  (66% available)
  Candidate 5:      ✗ ✓ ✓  (66% available)

Best Recommendation: Candidates 1-3 (earlier times preferred)
```

## 🚀 Deployment Architecture

### AWS Production Infrastructure (Implemented)

```
┌─────────────────────────────────────────────────────────────────┐
│                         AWS Cloud (VPC)                         │
│                                                                 │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │         Public Subnets (2 AZs)                           │  │
│  │                                                          │  │
│  │  ┌────────────────────────────────────────────────────┐ │  │
│  │  │   Application Load Balancer (ALB)                  │ │  │
│  │  │   • HTTP Listener (Port 80)                        │ │  │
│  │  │   • HTTPS Listener (Port 443 - optional)           │ │  │
│  │  │   • Health Checks: /health (30s interval)          │ │  │
│  │  │   • Target Group: EC2 instances on port 8080       │ │  │
│  │  └───────────────────┬────────────────────────────────┘ │  │
│  │                      │                                   │  │
│  │        ┌─────────────┼─────────────┐                    │  │
│  │        │             │             │                    │  │
│  │        ▼             ▼             ▼                    │  │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐             │  │
│  │  │ EC2 #1   │  │ EC2 #2   │  │ EC2 #3   │             │  │
│  │  │ Go App   │  │ Go App   │  │ Go App   │ ...         │  │
│  │  └──────────┘  └──────────┘  └──────────┘             │  │
│  │        │             │             │                    │  │
│  │        └─────────────┼─────────────┘                    │  │
│  │                      │                                   │  │
│  │         Auto Scaling Group                              │  │
│  │         • Min: 1 instance                               │  │
│  │         • Max: 4 instances                              │  │
│  │         • Desired: 2 instances                          │  │
│  │         • Launch Template with user_data                │  │
│  │         • Instance refresh for zero-downtime deploys    │  │
│  └──────────────────────┬────────────────────────────────────┘  │
│                         │                                        │
│  ┌──────────────────────▼────────────────────────────────────┐  │
│  │         Private Subnets (2 AZs)                           │  │
│  │                                                           │  │
│  │  ┌───────────────────────────────────────────────────┐   │  │
│  │  │  AWS RDS MySQL 8.0                                │   │  │
│  │  │  • Multi-AZ Deployment                            │   │  │
│  │  │  • Automated backups                              │   │  │
│  │  │  • Storage auto-scaling (20-100 GB)               │   │  │
│  │  │  • Encryption at rest                             │   │  │
│  │  │  • Performance Insights (prod)                    │   │  │
│  │  └───────────────────────────────────────────────────┘   │  │
│  └───────────────────────────────────────────────────────────┘  │
│                                                                 │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  Supporting Services                                      │  │
│  │  • AWS Secrets Manager (DB credentials)                  │  │
│  │  • CloudWatch Logs (app, error, system, access)          │  │
│  │  • CloudWatch Metrics & Dashboard                        │  │
│  │  • CloudWatch Alarms (CPU high/low, error rate)          │  │
│  │  • S3 Bucket (ALB access logs)                           │  │
│  │  • NAT Gateway (for private subnet internet access)      │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

### Auto-scaling Policies

```
CloudWatch Metrics → Auto Scaling Decisions:

┌─────────────────────────────────────────────────────────┐
│ CPU Utilization                                         │
│ 100% │███████████████████░░░░░░░░░░░░░░░░░░░░░░░░░░░░│ │
│  80% │███████████████████████████░░░░░░░░░░░ Threshold│ │
│  70% │███████████████████████████████░░░░░░ (Scale Up)│ │
│  50% │███████████████████████████████████████░░░░░░░░│ │ ← Target
│  20% │███████████████████████████████████████████████│ │ ← Scale Down
│   0% └───────────────────────────────────────────────┘ │
│       Time →                                            │
└─────────────────────────────────────────────────────────┘

Scaling Policies:
  1. Simple Scaling:
     • CPU > 70% for 2 min  → Add 1 instance
     • CPU < 20% for 2 min  → Remove 1 instance
     • Cooldown: 5 minutes

  2. Target Tracking:
     • Maintain 50% average CPU utilization
     • ALB automatically adjusts instance count

Current State Example:
  ┌────────────────────────────────────┐
  │ Instances: 2 (Desired)             │
  │ Min: 1, Max: 4                     │
  │ Average CPU: 45%                   │
  │ Health Status: 2/2 Healthy         │
  │ Action: Stable (no scaling needed) │
  └────────────────────────────────────┘
```

### CloudWatch Monitoring

```
Log Groups:
  /aws/ec2/{env}/application  → Application output
  /aws/ec2/{env}/error        → Error tracking
  /aws/ec2/{env}/system       → System/OS logs
  /aws/ec2/{env}/access       → Access logs

Metrics Collected:
  • ALB: Request count, latency, HTTP codes
  • EC2: CPU, memory, disk, network
  • RDS: Connections, CPU, storage, replication lag
  • Custom: Error rate, API response times

Alarms Configured:
  ┌─────────────────────────────────────────┐
  │ ⚠️  High CPU (>70%)     → Scale up     │
  │ ℹ️  Low CPU (<20%)      → Scale down   │
  │ 🚨 High Error Rate      → Alert team   │
  │ 📊 Dashboard Available  → Real-time    │
  └─────────────────────────────────────────┘
```

## 🔐 Security Layers (Optional/Future)

```
┌─────────────────────────────────────────────────────┐
│                   1. API Gateway                    │
│            (Rate Limiting, DDoS Protection)         │
└──────────────────────┬──────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────┐
│                2. Authentication                    │
│          (JWT, OAuth 2.0, API Keys)                 │
└──────────────────────┬──────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────┐
│                3. Authorization                     │
│         (RBAC, Resource-based Permissions)          │
└──────────────────────┬──────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────┐
│                4. Data Validation                   │
│          (Input Sanitization, Type Checking)        │
└──────────────────────┬──────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────┐
│                5. Business Logic                    │
│              (Service Layer)                        │
└──────────────────────┬──────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────┐
│                6. Data Access                       │
│         (Prepared Statements, Encryption)           │
└─────────────────────────────────────────────────────┘
```

## 📈 Scalability Strategy

### Vertical Scaling (Single Instance)
```
┌─────────────────┐
│   2 CPU Cores   │  →  ┌─────────────────┐
│   4 GB RAM      │      │   4 CPU Cores   │
│   100 GB Disk   │      │   8 GB RAM      │
└─────────────────┘      │   200 GB Disk   │
                         └─────────────────┘
  Good for: Early stage, simple deployments
  Limits: Hardware constraints, single point of failure
```

### Horizontal Scaling (Multiple Instances) - IMPLEMENTED
```
┌──────────┐
│ EC2 #1   │ ────┐
│ Go App   │     │
└──────────┘     │
                 │     ┌──────────────────────┐
┌──────────┐     ├────►│ Application Load     │◄──── Internet
│ EC2 #2   │ ────┤     │ Balancer (ALB)       │
│ Go App   │     │     └──────────────────────┘
└──────────┘     │
                 │        Auto Scaling Group
┌──────────┐     │        • Min: 1, Max: 4
│ EC2 #3   │ ────┤        • Health Checks
│ Go App   │     │        • Rolling Updates
└──────────┘     │
                 │
┌──────────┐     │
│ EC2 #4   │ ────┘
│ Go App   │ (scales based on demand)
└──────────┘

  ✅ Implemented: Production, high availability
  ✅ Benefits: No single point of failure, handles load spikes
  ✅ Features: Auto-scaling, health checks, zero-downtime deploys
```

### Database Scaling
```
┌──────────────┐
│   Primary    │ ◄──── Writes
│  (Read/Write)│
└───────┬──────┘
        │ Replication
        ├───────────────────┐
        │                   │
        ▼                   ▼
┌──────────────┐    ┌──────────────┐
│  Replica 1   │    │  Replica 2   │
│  (Read Only) │    │  (Read Only) │
└──────────────┘    └──────────────┘
        ▲                   ▲
        └───────────────────┴───── Read Queries
```

## 🔄 CI/CD Pipeline

```
Developer                GitLab              CI/CD                 AWS
    │                      │                   │                     │
    │ git push            │                   │                     │
    ├────────────────────►│                   │                     │
    │                      │ Trigger Pipeline │                     │
    │                      ├──────────────────►│                     │
    │                      │                   │                     │
    │                      │              ┌────▼─────┐              │
    │                      │              │  Build   │              │
    │                      │              │  Test    │              │
    │                      │              │  Lint    │              │
    │                      │              └────┬─────┘              │
    │                      │                   │                     │
    │                      │              ┌────▼─────┐              │
    │                      │              │  Build   │              │
    │                      │              │  Docker  │              │
    │                      │              │  Image   │              │
    │                      │              └────┬─────┘              │
    │                      │                   │                     │
    │                      │                   │ Push Image         │
    │                      │              ┌────▼─────┐              │
    │                      │              │ Registry │              │
    │                      │              └────┬─────┘              │
    │                      │                   │                     │
    │                      │                   │ Deploy              │
    │                      │                   ├────────────────────►│
    │                      │                   │                     │
    │                      │              ┌────▼─────┐         ┌────▼────┐
    │                      │              │ Health   │         │ Running │
    │                      │              │ Check    │         │  Pods   │
    │                      │              └──────────┘         └─────────┘
```

## 📊 Monitoring & Observability

```
Application                Metrics                Visualization
    │                         │                         │
    │ Prometheus metrics      │                         │
    ├────────────────────────►│                         │
    │ /metrics endpoint       │                         │
    │                         │ Scrape                  │
    │                     ┌───▼────┐                   │
    │                     │Prometh-│                   │
    │                     │  eus   │                   │
    │                     └───┬────┘                   │
    │                         │                         │
    │                         │ Query                   │
    │                         ├────────────────────────►│
    │                         │                    ┌────▼────┐
    │ Logs                    │                    │ Grafana │
    ├─────────────────────────┼───────────────────►│Dashboard│
    │ JSON structured         │                    └─────────┘
    │                         │
    │                    ┌────▼────┐
    │ Errors/Traces      │  ELK/   │
    └───────────────────►│Datadog  │
                         └─────────┘
```

## 🎯 Technology Decisions

| Component | Options Considered | Choice | Rationale |
|-----------|-------------------|--------|-----------|
| Web Framework | net/http, Gin, Mux, Echo | **Gorilla Mux** | Standard, lightweight, excellent routing, no bloat |
| Database | PostgreSQL, MySQL, DynamoDB | **AWS RDS MySQL 8.0** | Proven reliability, good timezone support, managed service, cost-effective |
| Database Driver | GORM, sqlx, raw SQL | **database/sql + AWS SDK** | Direct control, no ORM overhead, native AWS integration |
| Load Balancer | Nginx, HAProxy, ALB | **AWS Application Load Balancer** | Managed service, health checks, auto-scaling integration |
| Compute | ECS, EKS, EC2 | **EC2 with Auto Scaling Group** | Simpler than K8s, cost-effective, suitable for monolith |
| Monitoring | Prometheus, Datadog, CloudWatch | **AWS CloudWatch** | Native AWS integration, logs + metrics unified, cost-effective |
| IaC | Terraform, CloudFormation, Pulumi | **Terraform** | Multi-cloud capable, declarative, mature ecosystem |
| CI/CD | GitHub Actions, GitLab CI, Jenkins | **GitLab CI (Planned)** | Integrated, powerful, YAML-based |

## 📚 Next Steps

1. Review the [Implementation Plan](./IMPLEMENTATION_PLAN.md) for detailed phases
2. Check the [Quick Start Guide](./QUICKSTART.md) to begin implementation
3. Use the [Checklist](./CHECKLIST.md) to track progress
4. Start with Phase 1: Project Setup

**Understanding the architecture is key to successful implementation!** 🎯
