# Operations Runbook

## Service overview
- API: Go application exposing authenticated endpoints over HTTPS.
- Database: PostgreSQL 15.
- Cache/Event stream: Redis 7.
- Worker: background notification consumer reading the Redis stream.

## Startup and health checks

### Local development
```bash
set -a
source .env
set +a

docker compose up -d postgres redis migrate

go run ./cmd/api
```

### Health verification
```bash
curl -k https://localhost:8443/health
curl -k https://localhost:8443/metrics
```

## Common operational checks
- Confirm Postgres is healthy: `docker compose ps postgres`
- Confirm Redis is healthy: `docker compose ps redis`
- Confirm migration status: `docker compose logs migrate --tail 50`
- Confirm API logs: `docker compose logs api --tail 50`

## Incident response

### Dependency outage: PostgreSQL unavailable
1. Check `docker compose ps` and service health.
2. Validate credentials in `.env` and `docker-compose.yml` are correct.
3. Inspect `docker compose logs postgres`.
4. If needed, restore from the latest verified backup.
5. Rerun migration service to apply missing schema changes.

### Dependency outage: Redis unavailable
1. Check Redis health and process state.
2. Review `docker compose logs redis` for connection or persistence issues.
3. Reconnect the app after Redis returns to normal.
4. Verify the worker resumes consuming pending events from the stream.

### Application elevated latency or error spikes
1. Collect pprof profiles from `/debug/pprof/profile?seconds=30` or `/debug/pprof/heap`.
2. Inspect `/metrics` for request latency, errors, and DB saturation.
3. Compare with historical throughput and capacity assumptions.
4. Use graceful degradation by reducing batch or queue activity if needed.

## Backup and restore drill

### Backup
```bash
docker exec backend-api-postgres-1 pg_dump -U postgres -d employee_db > backup.sql
```

### Restore verification
```bash
docker exec -i backend-api-postgres-1 psql -U postgres -d employee_db < backup.sql
```

### RPO / RTO targets
- RPO: target 15 minutes for routine backups in a production deployment.
- RTO: target less than 1 hour for restore validation and service recovery.

## Alerting recommendations
- API 5xx error rate > threshold
- DB connection saturation or query latency increase
- Redis stream lag or worker backlog growth
- JWT validation failures at abnormal rate

## Known limitations
- The repo is currently a learning/development project rather than a hardened production multi-zone deployment.
- TLS certs are local files and should be replaced with proper certificate management in a real environment.
- Public exposure should be behind a reverse proxy or load balancer with WAF and rate limiting controls.
