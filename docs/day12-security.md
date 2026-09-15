# Day 12 - Authorization and API Security

## Practice

This sprint adds security enforcement for project resources and hardens API inputs.

### Implemented rules

- Enforce deny-by-default access checks for protected project resources.
- Support role-based access using admin, manager, member, and owner roles.
- Require membership validation before granting project access.
- Restrict project and task mutation by ownership and role.
- Reject malformed, oversized, and abusive requests.
- Add security headers and rate limiting.
- Audit security-sensitive actions.

## Authorization matrix

| Endpoint | Authenticated | Required role | Project membership required | Notes |
|---|---:|---|---:|---|
| GET /health | No | None | No | Public health check |
| POST /api/v1/auth/login | No | None | No | Login endpoint |
| POST /api/v1/auth/register | No | None | No | Register endpoint |
| GET /api/v1/me | Yes | Any user | No | Returns current identity |
| GET /api/v1/projects | Yes | member/admin/manager | Yes | Project access required |
| POST /api/v1/projects | Yes | admin/manager | No | Must still be authenticated |
| PUT /api/v1/projects/{id} | Yes | owner/admin/manager | Yes | Owner or manager authorization |
| DELETE /api/v1/projects/{id} | Yes | owner/admin | Yes | Higher privilege required |
| GET /api/v1/tasks | Yes | member/admin/manager | Yes | Must be in the project |
| POST /api/v1/tasks | Yes | member/admin/manager | Yes | Must be in the project |
| PUT /api/v1/tasks/{id} | Yes | owner/admin/manager | Yes | Task ownership or manager role |
| DELETE /api/v1/tasks/{id} | Yes | owner/admin/manager | Yes | Must satisfy mutation rules |

## Negative security tests

- User without membership cannot access project resources.
- User with member role cannot modify another project without permission.
- Non-admin cannot delete a project owned by another user.
- Oversized request bodies are rejected with 413.
- Invalid or malformed JSON is rejected with 400.
- High-frequency requests are rejected with 429.
- Unauthenticated requests fail with 401 and unauthorized requests fail with 403.

## Threat model

### Assets

- User identity and employee data
- Project metadata and permissions
- Task records and ownership
- JWT secrets and refresh tokens
- Database integrity and service availability

### Threats

- Unauthorized access to another user’s project
- Horizontal privilege escalation by changing ownership or project membership
- Mass request abuse and brute-force credential attacks
- Malformed or oversized payload injection
- Disclosure through missing security headers and CORS misconfiguration

### Mitigations

- JWT authentication plus project-level membership checks
- Deny-by-default authorization model
- Role and ownership validation before mutation
- Input validation and request-size limits
- Rate limiting for abuse prevention
- Security headers and conservative CORS configuration
- Audit logging for sensitive actions

### Residual risk

- Authorization logic still depends on correct membership and role data being maintained.
- External clients can still abuse the application if they bypass API access controls through compromised credentials.
- A large-scale deployment requires additional WAF, monitoring, and centralized audit collection.

## Security hardening in code

The project now includes:

- `internal/security/authorization.go` for role and access rules.
- `internal/delivery/http/middleware/security.go` for headers, request-size limiting, rate limiting, and CORS.
- Negative tests under `internal/delivery/http/middleware/security_test.go`.
- Authorization checks under `internal/security/authorization_test.go`.
