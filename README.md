# ShortURLs

A URL shortening service with support for private and single-use links, visit analytics, QR code generation, and DDoS protection.

## Features

- **Normal links** — shorten any URL with a custom or auto-generated alias
- **Private links** — access requires a unique Access Token (bcrypt-hashed), shown to the creator only once
- **Single-use links** — automatically deactivated after the first successful redirect
- **Expiration** — automatic deactivation after a configurable lifetime (in seconds)
- **Soft delete / deactivation** — `is_deleted`, `is_deactive` flags
- **Redirect** — HTTP 302 to the original URL
- **QR codes** — PNG generation with Redis caching
- **Analytics** — per-link and global (top 5, browser stats, counters)
- **Rate limiting** — 20 requests/minute per IP via Redis sliding window

## Tech Stack

| Component     | Technology       |
|---------------|------------------|
| Language      | Go 1.25          |
| HTTP Framework| Gin              |
| Database      | PostgreSQL 16    |
| Cache / Rate Limiter | Redis 7   |
| Token Hashing | bcrypt            |
| QR Codes      | go-qrcode         |

## Quick Start

```bash
# Start all services (PG + Redis + app)
docker compose up --build -d

# Or run databases in Docker, app locally
docker compose up -d postgres redis
go run ./cmd/server
```

The service will be available at `http://localhost:8080`.

## API

### Endpoints

| Method   | Route                    | Description                     |
|----------|--------------------------|---------------------------------|
| `POST`   | `/links`                 | Create a short link             |
| `GET`    | `/links/{alias}`         | Get link info                   |
| `PATCH`  | `/links/{alias}`         | Update a link                   |
| `DELETE` | `/links/{alias}`         | Soft delete a link              |
| `GET`    | `/{alias}`               | Redirect to the original URL   |
| `GET`    | `/links/{alias}/qr`      | Generate QR code (PNG)          |
| `GET`    | `/analytics`             | Global analytics + top 5        |
| `GET`    | `/analytics/{alias}`     | Per-link analytics              |
| `GET`    | `/health`                | Health check                    |

### Examples

#### Create a link

```bash
# Normal
curl -X POST http://localhost:8080/links \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com"}'

# With custom alias
curl -X POST http://localhost:8080/links \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com","alias":"myLink"}'

# Private (requires token for access)
curl -X POST http://localhost:8080/links \
  -H "Content-Type: application/json" \
  -d '{"url":"https://secret.com","is_private":true}'
# Response includes `access_token` — shown only once!

# Single-use (expires after first redirect)
curl -X POST http://localhost:8080/links \
  -H "Content-Type: application/json" \
  -d '{"url":"https://single.com","is_single":true}'

# With expiration time (seconds)
curl -X POST http://localhost:8080/links \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com","alias":"temp","lifetime":3600}'
```

#### Redirect

```bash
# Normal link
curl -v http://localhost:8080/myLink
# → 302 → redirect

# Private link (pass token as query param)
curl -v "http://localhost:8080/privateLink?token=5036c9c6cc..."
```

#### Analytics

```bash
# Per-link
curl http://localhost:8080/analytics/myLink

# Global (top 5 + totals + browser stats)
curl http://localhost:8080/analytics
```

### Response Codes

| Code | Meaning                                    |
|------|--------------------------------------------|
| 200  | Success                                    |
| 201  | Created                                    |
| 302  | Redirect                                   |
| 403  | Forbidden (private link without token)     |
| 404  | Not found                                  |
| 409  | Conflict (alias already taken)             |
| 410  | Gone (deleted / deactivated / expired)     |
| 429  | Too Many Requests (rate limit exceeded)    |

## Architecture

```
_shortURLs/
├── cmd/server/main.go           # Entry point
├── internal/
│   ├── config/                  # Environment config
│   ├── database/                # PostgreSQL / Redis connections + migrations
│   ├── models/                  # Data models (Link, Analytics)
│   ├── repository/              # Database access layer
│   ├── cache/                   # Redis caching + rate limiter
│   ├── handlers/                # HTTP request handlers
│   ├── middleware/              # Gin middleware (rate limiting)
│   └── utils/                   # Helpers (alias, token, QR)
├── migrations/                  # SQL migration files (reference)
├── docker-compose.yml           # PG + Redis + app
└── Dockerfile                   # Multi-stage Go build
```

## Data Models

### Link

| Field        | Type              | Description                     |
|--------------|-------------------|---------------------------------|
| id           | UUID              | Primary key                     |
| alias        | VARCHAR(255) UNIQUE| Short URL slug                 |
| url          | TEXT              | Original URL                    |
| lifetime     | INTEGER NULL      | Validity period in seconds      |
| is_deleted   | BOOLEAN           | Soft delete flag                |
| is_deactive  | BOOLEAN           | Deactivation flag               |
| is_private   | BOOLEAN           | Private link flag               |
| is_single    | BOOLEAN           | Single-use flag                 |
| access_token | TEXT NULL         | bcrypt hash of the access token |

### Analytics

| Field          | Type              | Description                  |
|----------------|-------------------|------------------------------|
| id             | UUID              | Primary key                  |
| link_id        | UUID (FK → Link)  | Reference to a link          |
| success_count  | INTEGER           | Successful redirects         |
| error_count    | INTEGER           | Failed redirect attempts     |
| first_visit_at | TIMESTAMPTZ       | First visit timestamp        |
| last_visit_at  | TIMESTAMPTZ       | Last visit timestamp         |
| browser_stats  | JSONB             | Per-browser visit counters   |
| qr_scan_count  | INTEGER           | QR code scan count           |

## Key Design Decisions

- **Access Token** is bcrypt-hashed and stored in the database; the raw token is shown to the creator exactly once in the creation response
- **Analytics counters** are written synchronously to PostgreSQL; an async Redis → PG flush worker can be added for high-load scenarios
- **Rate limiter** uses Redis sorted sets (sliding window algorithm) — 20 requests per minute per IP
- **Link caching** in Redis via key `link:{alias}` with a 10-minute TTL
- **QR codes** cached in Redis via key `qr:{alias}` with a 24-hour TTL
