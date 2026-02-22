# Customer Help Application MVP Backend

Go + Gin + PostgreSQL backend for customer support tickets with clean layering:

- `handlers` (HTTP)
- `service` (business rules)
- `repository` (PostgreSQL)

## Tech Stack

- Go (1.23+)
- Gin
- pgx/pgxpool
- PostgreSQL
- Goose migrations

## Project Structure

- `cmd/server/main.go`
- `internal/config`
- `internal/db`
- `internal/models`
- `internal/middleware`
- `internal/repository`
- `internal/service`
- `internal/handlers`
- `migrations`

## Environment Variables

- `PORT` (default `8080`)
- `DATABASE_URL` (required)

Example:

```bash
cp .env.example .env
```

## Install Dependencies

```bash
go mod tidy
```

## Migrations

Install goose CLI:

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
```

Run migrations:

```bash
goose -dir migrations postgres "$DATABASE_URL" up
```

Rollback last migration:

```bash
goose -dir migrations postgres "$DATABASE_URL" down
```

## Run Server

```bash
go run ./cmd/server
```

## API Error Format

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "...",
    "details": "..."
  }
}
```

## Example cURL

### Agencies

```bash
curl -X GET "http://localhost:8080/api/agencies?city=Bandung&province=West%20Java&q=service&page=1&page_size=20"
```

### Customer tickets

List customer tickets:

```bash
curl -X GET "http://localhost:8080/api/tickets?status=NEW&page=1&page_size=20" \
  -H "X-Customer-Id: aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
```

Get ticket detail:

```bash
curl -X GET "http://localhost:8080/api/tickets/<ticket-id>" \
  -H "X-Customer-Id: aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
```

Create ticket:

```bash
curl -X POST "http://localhost:8080/api/tickets" \
  -H "Content-Type: application/json" \
  -H "X-Customer-Id: aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa" \
  -d '{
    "device_serial":"SN-001-XYZ",
    "subject":"Device not booting",
    "description":"The device shows a black screen after startup.",
    "category":"HARDWARE"
  }'
```

Add attachment:

```bash
curl -X POST "http://localhost:8080/api/tickets/<ticket-id>/attachments" \
  -H "Content-Type: application/json" \
  -H "X-Customer-Id: aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa" \
  -d '{
    "file_type":"IMAGE",
    "file_url":"https://cdn.example.com/ticket/attachment-1.jpg"
  }'
```

Add feedback:

```bash
curl -X POST "http://localhost:8080/api/tickets/<ticket-id>/feedback" \
  -H "Content-Type: application/json" \
  -H "X-Customer-Id: aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa" \
  -d '{
    "score":5,
    "solved":true,
    "comment":"Issue fixed quickly"
  }'
```

### Technician tickets

List technician tickets:

```bash
curl -X GET "http://localhost:8080/api/tech/tickets?status=IN_REVIEW&assigned=true&page=1&page_size=20" \
  -H "X-Technician-Id: bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
```

Patch ticket:

```bash
curl -X PATCH "http://localhost:8080/api/tech/tickets/<ticket-id>" \
  -H "Content-Type: application/json" \
  -H "X-Technician-Id: bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb" \
  -H "X-Actor-Role: TECHNICIAN" \
  -H "X-Actor-Id: bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb" \
  -d '{
    "status":"RESOLVED",
    "close_reason":"Replaced faulty cable"
  }'
```

Add technician comment:

```bash
curl -X POST "http://localhost:8080/api/tech/tickets/<ticket-id>/comment" \
  -H "Content-Type: application/json" \
  -H "X-Technician-Id: bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb" \
  -d '{"message":"Waiting for vendor confirmation"}'
```

## Notes

- `ticket_number` is generated using `ticket_number_seq` and formatted as `TK-#####`.
- Feedback is allowed only when ticket status is `RESOLVED` or `REJECTED`.
- Ticket writes that involve event creation are wrapped in DB transactions.
