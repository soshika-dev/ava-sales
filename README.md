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
- JWT auth (`Authorization: Bearer <token>`)

## Security change

`X-Customer-Id` and `X-Technician-Id` are **not trusted**.
Identity is taken from JWT `sub`, then server resolves user role/customer/technician IDs from `app_users`.

## Environment Variables

- `PORT` (default `8080`)
- `DATABASE_URL` (required)
- `JWT_SECRET` (required)

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

## Run Server

```bash
go run ./cmd/server
```

## Dev JWTs

`004_create_app_users.sql` seeds these users:

- customer user id: `11111111-1111-1111-1111-111111111111`
- technician user id: `22222222-2222-2222-2222-222222222222`
- admin user id: `33333333-3333-3333-3333-333333333333`

Create token (example using [jwt.io](https://jwt.io)):

- Header: `{ "alg": "HS256", "typ": "JWT" }`
- Payload: `{ "sub": "11111111-1111-1111-1111-111111111111", "exp": 1924992000 }`
- Sign with `JWT_SECRET`

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

### Agencies (public)

```bash
curl -X GET "http://localhost:8080/api/agencies?city=Bandung&province=West%20Java&q=service&page=1&page_size=20"
```

### Customer tickets (JWT required)

```bash
curl -X POST "http://localhost:8080/api/tickets" \
  -H "Authorization: Bearer <customer_jwt>" \
  -H "Content-Type: application/json" \
  -d '{
    "device_serial":"SN-001-XYZ",
    "subject":"Device not booting",
    "description":"The device shows a black screen after startup.",
    "category":"HARDWARE"
  }'
```

```bash
curl -X GET "http://localhost:8080/api/tickets?page=1&page_size=20" \
  -H "Authorization: Bearer <customer_jwt>"
```

### Technician tickets (JWT required)

```bash
curl -X PATCH "http://localhost:8080/api/tech/tickets/<ticket-id>" \
  -H "Authorization: Bearer <tech_jwt>" \
  -H "Content-Type: application/json" \
  -d '{
    "status":"RESOLVED",
    "close_reason":"Replaced faulty cable"
  }'
```
