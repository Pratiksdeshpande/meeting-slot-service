.PHONY: help init deps test test-coverage run build lint migrate-up migrate-down clean \
       docker-mysql docker-mysql-stop docker-mysql-logs run-local

# Docker MySQL container settings
MYSQL_CONTAINER_NAME := mysql-meeting
MYSQL_ROOT_PASSWORD := password
MYSQL_DATABASE := meetingslots
MYSQL_USER := appuser
MYSQL_PASSWORD := password
MYSQL_PORT := 3306

help:
	@echo "Available targets:"
	@echo "  init              - Initialize project"
	@echo "  deps              - Install dependencies"
	@echo "  test              - Run tests"
	@echo "  test-coverage     - Run tests with coverage"
	@echo "  run               - Run application"
	@echo "  run-local         - Load env and run application (Linux/Mac)"
	@echo "  build             - Build binary"
	@echo "  lint              - Run linter"
	@echo "  migrate-up        - Run database migrations"
	@echo "  migrate-down      - Rollback database migrations"
	@echo "  clean             - Clean build artifacts"
	@echo ""
	@echo "Docker targets:"
	@echo "  docker-mysql      - Start MySQL in Docker"
	@echo "  docker-mysql-stop - Stop and remove MySQL container"
	@echo "  docker-mysql-logs - View MySQL container logs"

init:
	go mod download

deps:
	go mod tidy
	go mod download

test:
	go test -v -race ./...

test-coverage:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

run:
	go run cmd/server/main.go

build:
	CGO_ENABLED=0 go build -o bin/server cmd/server/main.go

lint:
	golangci-lint run

migrate-up:
	@echo "Running database migrations..."
	go run cmd/migrate/main.go up

migrate-down:
	@echo "Rolling back database migrations..."
	go run cmd/migrate/main.go down

clean:
	rm -rf bin/
	rm -f coverage.out coverage.html
	go clean

# =============================================================================
# Docker MySQL Targets
# =============================================================================

docker-mysql:
	@echo "Starting MySQL container..."
	@docker run --name $(MYSQL_CONTAINER_NAME) \
		-e MYSQL_ROOT_PASSWORD=$(MYSQL_ROOT_PASSWORD) \
		-e MYSQL_DATABASE=$(MYSQL_DATABASE) \
		-e MYSQL_USER=$(MYSQL_USER) \
		-e MYSQL_PASSWORD=$(MYSQL_PASSWORD) \
		-p $(MYSQL_PORT):3306 \
		-d mysql:8.0
	@echo "MySQL container started. Waiting for initialization..."
	@sleep 10
	@echo "MySQL is ready at localhost:$(MYSQL_PORT)"

docker-mysql-stop:
	@echo "Stopping MySQL container..."
	@docker stop $(MYSQL_CONTAINER_NAME) 2>/dev/null || true
	@docker rm $(MYSQL_CONTAINER_NAME) 2>/dev/null || true
	@echo "MySQL container stopped and removed"

docker-mysql-logs:
	@docker logs -f $(MYSQL_CONTAINER_NAME)

# =============================================================================
# Local Development
# =============================================================================

run-local:
	@echo "Loading environment and starting server..."
	@. ./env.local.sh && go run cmd/server/main.go
