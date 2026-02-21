# Ava Sales Ticketing API (Go + Gin + PostgreSQL)

Production-ready REST API with Gin, pgx, PostgreSQL, SQL migrations (golang-migrate), UUID PKs, ticket timeline events, and consistent error responses.

## Architecture

```
cmd/server/main.go
internal/config
internal/db
internal/models
internal/repositories
internal/services
internal/handlers
internal/router
migrations
```

## Business Rules

- `updated_at` is auto-maintained via DB trigger.
- Closing statuses (`RESOLVED`, `REJECTED`) require `close_reason`; `closed_at` set on close.
- Transition away from terminal statuses is **forbidden**.
- Feedback is one-per-ticket (duplicate insert returns `409 conflict`).
- Ticket update creates event rows:
  - status change => `STATUS_CHANGED`
  - assignment update => `ASSIGN`
  - close action => additional `CLOSE`

## Requirements

- Go 1.22+
- PostgreSQL 15+
- golang-migrate CLI (`migrate`)

## Environment Variables

- `DATABASE_URL` (required)
- `PORT` (optional, default `8080`)

## Run locally

1. Start Postgres:

```bash
docker compose up -d postgres
```

2. Export env:

```bash
export DATABASE_URL='postgres://app:app@localhost:5432/ava_sales?sslmode=disable'
export PORT=8080
```

3. Run migrations:

```bash
migrate -path migrations -database "$DATABASE_URL" up
```

4. Start API:

```bash
go run ./cmd/server
```

## Run via Docker Compose (API + DB)

```bash
docker compose up --build
```

## API Base Path

`/api`

## Error format

```json
{
  "error": {
    "code": "validation_error",
    "message": "..."
  }
}
```

## Example cURL

### Agencies

```bash
curl -s "http://localhost:8080/api/agencies?city=Bandung&province=West%20Java&q=service&limit=20&offset=0"
```

```bash
curl -s "http://localhost:8080/api/agencies/<agency_id>"
```

### Tickets (Customer)

```bash
curl -s "http://localhost:8080/api/tickets?customer_id=<customer_uuid>"
```

```bash
curl -s "http://localhost:8080/api/tickets/<ticket_id>"
```

```bash
curl -s -X POST "http://localhost:8080/api/tickets" \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id":"11111111-1111-1111-1111-111111111111",
    "device_serial":"SN-ABC-123",
    "subject":"Device won\"t boot",
    "description":"The device fails to start after update",
    "category":"hardware"
  }'
```

```bash
curl -s -X POST "http://localhost:8080/api/tickets/<ticket_id>/attachments" \
  -H "Content-Type: application/json" \
  -d '{
    "file_type":"IMAGE",
    "file_url":"https://cdn.example.com/images/1.jpg",
    "uploaded_by":"11111111-1111-1111-1111-111111111111"
  }'
```

```bash
curl -s -X POST "http://localhost:8080/api/tickets/<ticket_id>/feedback" \
  -H "Content-Type: application/json" \
  -d '{"score":5,"solved":true,"comment":"Fast response"}'
```

```bash
curl -s "http://localhost:8080/api/tickets/<ticket_id>/events"
```

### Tickets (Technician)

```bash
curl -s "http://localhost:8080/api/tech/tickets?status=IN_PROGRESS&assigned_technician_id=<tech_uuid>"
```

```bash
curl -s -X PATCH "http://localhost:8080/api/tech/tickets/<ticket_id>?actor_id=<tech_uuid>" \
  -H "Content-Type: application/json" \
  -d '{
    "status":"RESOLVED",
    "assigned_technician_id":"22222222-2222-2222-2222-222222222222",
    "close_reason":"Replaced failing battery"
  }'
```

```bash
curl -s -X POST "http://localhost:8080/api/tech/tickets/<ticket_id>/comment" \
  -H "Content-Type: application/json" \
  -d '{
    "actor_id":"22222222-2222-2222-2222-222222222222",
    "message":"Please confirm after restarting your device"
  }'
```
