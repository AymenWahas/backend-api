# Architecture Decision Records

## ADR-001: Use PostgreSQL as the system of record

Status: Accepted

Context:
- The application models structured domain entities such as employees, projects, tasks, memberships, and auth sessions.
- These entities require relational constraints, transactional updates, and reporting-friendly queries.

Decision:
- PostgreSQL is the primary persistence layer and source of truth.
- GORM is used to access the database, with explicit migration management stored under the `migrations/` directory.

Consequences:
- Strong consistency and ACID semantics are available for critical workflows.
- Operational complexity increases because database backup, restore, and schema change plans are now needed.

## ADR-002: Use Redis Streams for task event publishing

Status: Accepted

Context:
- The API needs asynchronous notifications without blocking the primary request path.
- A reliable event stream is preferred over direct synchronous side effects.

Decision:
- Task creation events are published to a Redis Stream.
- A worker reads the stream, performs idempotent notification persistence, and acknowledges events after processing.

Consequences:
- API latency is reduced for the main request flow.
- Eventual consistency is introduced, so operational monitoring and recovery procedures are required.

## ADR-003: Protect secrets via environment variables and reject missing values

Status: Accepted

Context:
- JWT signing keys and database credentials are sensitive and cannot be hard-coded in source control.

Decision:
- Secrets are provided through environment variables and `docker-compose.yml` now fails fast with `${VAR:?message}` when required variables are absent.
- Repository samples use `.env.example` without shipping real credentials.

Consequences:
- Deployment becomes safer and more reproducible.
- Operators must provide a real `.env` or secret manager values before startup.

## ADR-004: Publish pprof endpoints only for local debugging

Status: Accepted

Context:
- CPU, heap, goroutine, and mutex profiling are necessary for performance diagnosis.
- Public exposure of profiling data can leak operational details and increase attack surface.

Decision:
- `pprof` routes are exposed behind the same server and protected by the application middleware stack, but they are intended for local diagnosis and controlled access only.

Consequences:
- Operators can collect diagnostics without adding external profiling tooling.
- Access control and network segmentation remain important in production deployments.
