# Customer Help Application MVP Backend

Go + Gin + PostgreSQL backend for support tickets with layered architecture:

- handlers (HTTP)
- service (business)
- repository (DB)

## Stack

- Go 1.23+
- Gin
- pgx/pgxpool
- PostgreSQL
- Goose migrations
- JWT (HS256) + bcrypt

## Auth Overview

- Login: `POST /api/auth/login`
- Refresh: `POST /api/auth/refresh`
- Me: `GET /api/auth/me`
- Protected APIs use `Authorization: Bearer <access_token>`
- Access token includes `sub=<app_users.id>` and short expiration
- Refresh token is opaque, stored hashed in DB, and rotated on refresh

Client-provided identity headers are not trusted for auth.

## Environment

- `PORT` (default `8080`)
- `DATABASE_URL` (required)
- `JWT_SECRET` (required)
- `ACCESS_TOKEN_TTL_MINUTES` (default `15`)
- `REFRESH_TOKEN_TTL_DAYS` (default `7`)

```bash
cp .env.example .env
```

## Install

```bash
go mod tidy
```

## Migrate

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
goose -dir migrations postgres "$DATABASE_URL" up
```

## Run

```bash
go run ./cmd/server
```

## Dev seeded users

Migration seeds:

- customer user id: `11111111-1111-1111-1111-111111111111`
- technician user id: `22222222-2222-2222-2222-222222222222`
- admin user id: `33333333-3333-3333-3333-333333333333`

Credentials (all seeded users):

- password: `password`
- emails: `customer@example.com`, `tech@example.com`, `admin@example.com`
- usernames: `customer1`, `tech1`, `admin1`

## Postman quick flow

1) Login

```http
POST /api/auth/login
Content-Type: application/json

{
  "email": "customer@example.com",
  "password": "password"
}
```

2) Call protected endpoint

```http
GET /api/tickets?page=1&page_size=20
Authorization: Bearer <access_token>
```

3) Refresh access token

```http
POST /api/auth/refresh
Content-Type: application/json

{
  "refresh_token": "<refresh_token>"
}
```

## Error format

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "...",
    "details": "..."
  }
}
```
