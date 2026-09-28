# WorkHub Backend API

A Go-based backend for managing employees, projects, tasks, memberships, authentication, notifications, and related business workflows. The service exposes a REST API over HTTPS, persists data in PostgreSQL, uses Redis for caching and streaming, and applies JWT-based access control for protected endpoints.

## Overview

This project is a work-management backend designed around a company/team environment. It includes:

- Employee management
- Project and membership management
- Task creation, updates, and deletion
- User authentication with registration/login/refresh/logout
- JWT-protected routes with current-user lookup
- Redis Streams for task event publication and background notifications
- Asynchronous notification worker processing
- Metrics and health endpoints
- Swagger/OpenAPI documentation through a UI and generated spec

## Tech Stack

- Go 1.26+
- PostgreSQL via GORM and pgx
- Redis for cache and streaming
- JWT authentication using golang-jwt
- TLS/HTTPS server configuration
- Prometheus metrics
- Swagger UI via swaggest/swgui
- Docker and Docker Compose for local orchestration

## Core Architecture

The app follows a clean layered design:

- `cmd/api` — application entry point
- `internal/config` — environment-driven configuration validation
- `internal/database` — PostgreSQL connection and transaction helpers
- `internal/repository` — database access layer
- `internal/usecase` — business logic
- `internal/delivery/http` — HTTP handlers, routing, middleware, and Swagger UI
- `internal/domain` — core domain models and error definitions
- `internal/auth` — password hashing, JWT utilities, refresh-token handling
- `internal/cache` — Redis client setup
- `internal/messaging` — Redis Streams publisher
- `internal/event` — event payload definitions
- `internal/worker` — background notification consumer
- `internal/observability` — Prometheus metrics reporting
- `migrations` — database migration files
- `docs` — generated Swagger/OpenAPI artifacts
- `certs` — TLS certificate and private key

## Project Structure

```text
backend-api/
├── cmd/
│   └── api/
│       └── main.go
├── certs/
│   ├── cert.pem
│   └── key.pem
├── docs/
│   ├── docs.go
│   ├── swagger.json
│   └── swagger.yaml
├── internal/
│   ├── auth/
│   ├── cache/
│   ├── config/
│   ├── database/
│   ├── delivery/
│   │   └── http/
│   │       ├── handler/
│   │       ├── middleware/
│   │       └── router.go
│   ├── domain/
│   ├── event/
│   ├── messaging/
│   ├── observability/
│   ├── repository/
│   ├── security/
│   ├── usecase/
│   └── worker/
├── migrations/
│   ├── 000001_...
│   └── ...
├── .gitignore
├── docker-compose.yml
├── Dockerfile
├── go.mod
├── README.md
└── ...
```

## Domain Model Summary

The application revolves around the following entities:

- Employee
- User
- AuthSession
- Project
- Membership
- Task
- Notification

These objects are backed by PostgreSQL and are used across the repository, usecase, and HTTP layers.

## Authentication and Authorization

Authentication is implemented through:

- Password hashing with bcrypt-like secure password utilities
- JWT creation/validation for access tokens
- Refresh tokens stored in the database for revocation or invalidation workflows
- Request middleware that validates bearer tokens

Protected routes use the auth middleware to check token validity and attach user context before continuing.

### Auth endpoints

- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/refresh`
- `POST /api/v1/auth/logout`
- `GET /api/v1/me`

## Main API Routes

### Health

- `GET /health`

### Employees

- `GET /api/v1/employees`
- `POST /api/v1/employees`
- `GET /api/v1/employees/{id}`
- `PUT /api/v1/employees/{id}`
- `DELETE /api/v1/employees/{id}`

### Projects

- `GET /api/v1/projects`
- `POST /api/v1/projects`
- `GET /api/v1/projects/{id}`
- `PUT /api/v1/projects/{id}`
- `DELETE /api/v1/projects/{id}`
- `GET /api/v1/projects/{id}/members`
- `POST /api/v1/projects/{id}/members`
- `PATCH /api/v1/projects/{id}/members/{employeeId}/role`
- `DELETE /api/v1/projects/{id}/members/{employeeId}`

### Tasks

- `POST /api/v1/tasks`
- `GET /api/v1/tasks`
- `GET /api/v1/tasks/{id}`
- `PUT /api/v1/tasks/{id}`
- `DELETE /api/v1/tasks/{id}`

### Observability

- `GET /metrics`

### Swagger UI

- `GET /swagger/`
- `GET /openapi.json`

## Configuration

Configuration is loaded from environment variables in `internal/config/config.go` and validated before startup.

### Required environment variables

```bash
PORT=8443
DB_HOST=localhost
DB_PORT=5434
DB_USER=postgres
DB_PASSWORD=postgrespassword
DB_NAME=employee_db
REDIS_ADDR=localhost:6379
JWT_SECRET=your-secret-key
REQUEST_TIMEOUT=5s
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=10
DB_CONN_MAX_LIFETIME=30m
DB_CONN_MAX_IDLE_TIME=5m
ACCESS_TOKEN_TTL=15m
ALLOWED_ORIGINS=*
```

Important validation rules:

- `PORT` must not be empty
- `REQUEST_TIMEOUT` must be greater than zero
- `DB_MAX_OPEN_CONNS` must be greater than zero
- `DB_MAX_IDLE_CONNS` cannot be negative or exceed `DB_MAX_OPEN_CONNS`
- `JWT_SECRET` is required
- `ACCESS_TOKEN_TTL` must be greater than zero

## Local Development Setup

### 1) Configure environment variables

Copy the sample environment file and update the secrets before starting the stack.

```bash
cp .env.example .env
```

Then set the values in `.env` for:

```bash
DB_PASSWORD=postgrespassword
JWT_SECRET=replace-with-a-long-random-secret
```

### 2) Install dependencies

```bash
go mod download
```

### 3) Start supporting services

This project expects PostgreSQL and Redis to be running locally or via Docker Compose.

```bash
export $(grep -v '^#' .env | xargs)
docker compose up -d postgres redis
```

### 4) Run database migrations

The Compose file includes a `migrate` service for applying the SQL migration files in the `migrations/` directory.

```bash
docker compose up -d migrate
```

### 5) Run the API

```bash
set -a
source .env
set +a
go run ./cmd/api
```

The application starts an HTTPS server on port `8443` by default using:

- `certs/cert.pem`
- `certs/key.pem`

## Docker Compose

The repository includes `docker-compose.yml` with the following services:

- `postgres` — PostgreSQL 15 database
- `migrate` — runs SQL migrations against PostgreSQL
- `redis` — Redis 7
- `api` — builds the Go backend and exposes it on port `8443`

Example launch:

```bash
docker compose up --build
```

The `api` service also sets:

- `DB_HOST=postgres`
- `DB_PORT=5432`
- `DB_USER=postgres`
- `DB_PASSWORD=${DB_PASSWORD}`
- `DB_NAME=employee_db`
- `REDIS_ADDR=redis:6379`

## Redis and Event Flow

The app publishes task-related events to a Redis stream and uses a background notification worker to consume them asynchronously.

Typical flow:

1. A task is created or updated
2. The use case publishes an event to Redis Streams
3. The worker listens for new stream messages
4. A notification record is created for downstream processing

This makes task notifications non-blocking for the main API request lifecycle.

## Metrics and Observability

The application exposes Prometheus metrics at:

- `GET /metrics`

It also emits JSON logs using Go's `slog` package and records database statistics in a background loop.

## Swagger / OpenAPI

The project generates Swagger artifacts in the `docs/` directory, and the docs are served through the server.

Generated files:

- `docs/docs.go`
- `docs/swagger.json`
- `docs/swagger.yaml`

Swagger UI route:

- `http://localhost:8443/swagger/`

OpenAPI raw JSON route:

- `http://localhost:8443/openapi.json`

To regenerate Swagger docs manually:

```bash
go install github.com/swaggo/swag/cmd/swag@v1.8.9
$(go env GOPATH)/bin/swag init -g cmd/api/main.go -o docs
```

## Database Migrations

The project contains SQL migrations under `migrations/` and uses a migrate container to apply them on startup.

Migration patterns include tables for:

- employees
- projects
- tasks
- memberships
- notifications
- users
- auth sessions

## Notes on Security and Production Readiness

This project includes important backend conventions but should still be reviewed before production use:

- use strong secrets and rotate JWT keys regularly
- configure TLS certificates properly in real deployments
- restrict allowed origins and network exposures
- add rate limiting and WAF protections if exposed publicly
- centralize secrets in environment management or a secret store
- add test coverage for authentication, authorization, and database edge cases

## Build and Verification

To verify the project builds successfully:

```bash
go build ./...
```

This project was validated with the repository build command and is expected to compile cleanly when dependencies are available.

## License

This repository is intended for learning, backend development practice, and project-based API work. If a formal license is required for your environment, add it according to your organization or assignment rules.
