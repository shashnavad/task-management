# Next Steps

## Comprehensive Testing Strategy
To improve confidence across features, add tests in this order:

1. **Core Unit Tests (fast feedback)**
   - Task service edge cases:
     - creating tasks with optional fields (`assignee_id`, `due_date`) missing
     - invalid status transitions (if business rules are added)
     - update/delete on non-existent task IDs
   - Auth service:
     - password hashing and verification
     - JWT generation/parsing with expiration and malformed tokens
   - Middleware:
     - missing/invalid auth header
     - role/permission enforcement scenarios

2. **Repository + DB Integration Tests**
   - CRUD lifecycle per service repository (create/read/update/delete)
   - transaction and rollback behavior for failures
   - constraints and data integrity checks (foreign key, nullability, uniqueness)
   - pagination/filter queries where applicable

3. **API/Handler Tests**
   - request validation and status codes (`400`, `401`, `403`, `404`, `409`, `500`)
   - happy path and error path for each endpoint in `auth`, `task`, `project`, `file`, and `notification`
   - JSON schema/shape assertions for responses

4. **Event-Driven and Saga Reliability Tests**
   - verify event payload structure and required metadata fields
   - test idempotency for duplicate event deliveries
   - saga failure injection to ensure compensation is always triggered correctly
   - out-of-order event handling for consumers

5. **End-to-End and Performance Tests**
   - API Gateway -> service routing tests for key user journeys
   - WebSocket notification flow tests after task/project events
   - load tests around task creation + event publication latency
   - regression benchmarks for p95 response time and saga completion duration

## Useful Commands for CI and Local Validation
```sh
# Unit/integration tests
go test -v ./...

# Coverage report
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out

# Race detector (high value for concurrent/event code)
go test -race ./...

# Repeat tests to catch flakiness
go test -count=10 ./shared/... ./services/task/...
```
