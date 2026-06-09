# Todo List API

Production-grade Go REST API — chi router, Clean Architecture, PostgreSQL, JWT auth, Swagger UI.

## Quick start

```bash
cp .env.example .env          # edit DATABASE_URL, JWT_SECRET, etc.
make docker-up                # start postgres + app
make migrate-up               # apply migrations
# app is live at http://localhost:8080
# swagger UI at http://localhost:8080/swagger/index.html
```

## Development

```bash
# postgres must be running
make run          # go run ./cmd/api
make test         # unit tests (no DB required)
make lint         # golangci-lint
make swag         # regenerate Swagger docs from annotations
```

## Migrations

Requires [golang-migrate](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate):

```bash
make migrate-up
make migrate-down
```

## Docker

```bash
make docker-up    # build + start
make docker-down  # stop + remove containers
```

## API

All protected routes require: `Authorization: Bearer <token>`

### Auth

```bash
# Register
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"alice@example.com","password":"password123","name":"Alice"}'

# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"alice@example.com","password":"password123"}'
```

Both return:
```json
{"data": {"access_token": "...", "user": {...}}}
```

### Users

```bash
curl http://localhost:8080/api/v1/me \
  -H 'Authorization: Bearer <token>'
```

### Todos

```bash
# Create
curl -X POST http://localhost:8080/api/v1/todos \
  -H 'Authorization: Bearer <token>' \
  -H 'Content-Type: application/json' \
  -d '{"title":"Buy milk","description":"Whole milk","due_date":"2026-06-15T00:00:00Z"}'

# List (paginated, optional completed filter)
curl 'http://localhost:8080/api/v1/todos?page=1&limit=20&completed=false' \
  -H 'Authorization: Bearer <token>'

# Get by ID
curl http://localhost:8080/api/v1/todos/<id> \
  -H 'Authorization: Bearer <token>'

# Update
curl -X PATCH http://localhost:8080/api/v1/todos/<id> \
  -H 'Authorization: Bearer <token>' \
  -H 'Content-Type: application/json' \
  -d '{"completed":true}'

# Delete
curl -X DELETE http://localhost:8080/api/v1/todos/<id> \
  -H 'Authorization: Bearer <token>'
```

### Health

```bash
curl http://localhost:8080/healthz
curl http://localhost:8080/readyz
```

## Architecture

```
HTTP request
  → chi router
  → middleware (request_id, logger, recover, cors, auth)
  → Handler (decode, validate, call service)
  → Service (business logic)
  → Repository interface
  → PostgreSQL via pgx
```

Feature packages (`user`, `auth`, `todo`) each contain model, DTO, repository interface, postgres implementation, service, and handler. Dependencies flow inward; no package imports a sibling package except `auth` which needs `user.Repository`.

## Swagger

Annotations live in handler files. After changing handlers or DTOs:

```bash
make swag    # regenerates docs/
```

Swagger UI: `http://localhost:8080/swagger/index.html`
