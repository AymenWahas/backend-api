# WorkHub Backend API

A Go-based backend for managing employees, projects, tasks, and user authentication. The service exposes a REST API over HTTPS and uses PostgreSQL for persistence, Redis for cache/event streaming, and JWT-based authentication for protected endpoints.

## Overview

This project implements a company/work-management API with the following core features:

- Employee management
- Project management
- Task management
- User registration and login
- JWT access token authentication
- Refresh-token session handling
- Redis Streams event publishing for task events
- Notification worker processing for async task notifications
- OpenAPI documentation

## Tech Stack

- Go
- PostgreSQL with GORM
- Redis
- JWT authentication
- HTTPS server with TLS certificates
- Redis Streams for event-driven task notifications

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
│   ├── openapi.json
│   └── openapi.yaml
├── internal/
│   ├── auth/
│   ├── cache/
│   ├── config/
│   ├── database/
│   ├── delivery/
│   ├── domain/
│   ├── event/
│   ├── messaging/
│   ├── repository/
│   ├── usecase/
│   └── worker/
├── migrations/
├── go.mod
├── README.md
└── ...
```

## Core Domain Models

- Employee: employee identity, email, department, timestamps
- Project: project owner, name, description, version, associated tasks
- Task: project association, title, status, timestamps
- User: employee-linked user account with password hash
- AuthSession: refresh-token session tracking and revocation

## Authentication

The API supports register/login/refresh/logout flows using JWT and refresh tokens.

### Auth endpoints

- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/refresh`
- `POST /api/v1/auth/logout`
- `GET /api/v1/me` (requires JWT auth)

Protected routes use an auth middleware that validates JWT tokens and attaches the employee/user context.

## Main API Routes

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

### Tasks

- `GET /api/v1/tasks`
- `POST /api/v1/tasks`
- `GET /api/v1/tasks/{id}`
- `PUT /api/v1/tasks/{id}`
- `DELETE /api/v1/tasks/{id}`

### Health

- `GET /health`

### Documentation

- `GET /swagger/`
- `GET /openapi.json`

## Configuration

The app reads configuration from environment variables. The default values are defined in `internal/config/config.go`.

Required/important variables:

```bash
PORT=8443
DB_HOST=localhost
DB_PORT=5434
DB_USER=postgres
DB_PASSWORD=postgrespassword
DB_NAME=employee_db
REQUEST_TIMEOUT=5s
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=10
DB_CONN_MAX_LIFETIME=30m
DB_CONN_MAX_IDLE_TIME=5m
JWT_SECRET=your-secret-key
ACCESS_TOKEN_TTL=15m
```

The application validates that `JWT_SECRET` is present before starting.

## Running the Project

1. Start PostgreSQL and Redis.
2. Set the required environment variables.
3. Run database migrations if needed.
4. Start the API:

```bash
export JWT_SECRET="your-secret-key"
go run ./cmd/api
```

The server listens on HTTPS using:

- `certs/cert.pem`
- `certs/key.pem`

and defaults to port `8443` unless overridden.

## Redis and Event Flow

The service publishes task events to a Redis Stream named `task-events` when tasks are created or modified. A worker consumes those events and stores notifications in the database.

This gives the application an async notification pattern without blocking the main API flow.

## Database Notes

The project includes migration files under the `migrations/` directory. PostgreSQL is used as the main transactional database, and the app uses GORM models mapped to tables such as:

- employees
- projects
- tasks
- users
- auth_sessions
- notifications

## Notes on API Style

- Uses JSON request/response payloads
- Validates request payloads before processing
- Returns consistent JSON error responses
- Uses middleware for logging, request IDs, timeouts, recovery, and JSON content type enforcement

## Future Ideas

Possible improvements include:

- role-based access control
- user profile management
- pagination/filtering improvements
- queue-based notification reliability
- more complete Swagger examples and security schemes

## License

This project is for learning and backend API practice. Check the repository status and local project instructions for any specific licensing or assignment constraints.
