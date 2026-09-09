# RateGuard

Centralized API rate limiting for your services.

RateGuard provides token bucket, fixed window, and sliding window
rate limiting through a simple HTTP API.

```text
Docker → Quick Start → API → SDKs
```

## Run with Docker

### Pull the image

```bash
docker pull ghcr.io/faizahmd2/rate-guard:latest
```

Make sure redis is runnng.

```bash
docker run -d \
  --name rateguard \
  -p 4215:4215 \
  -e REDIS_HOST=localhost \
  -e REDIS_PORT=6379 \
  -e RATEGUARD_COOKIE_SECRET="your-generated-secret" \
  -e ENVIRONMENT=production \
  -v rateguard_data:/app/data \
  ghcr.io/faizahmd2/rate-guard:latest
```

RateGuard will be available at:

```text
http://localhost:4215
```

Open the URL in your browser to access the admin UI.

---

## Official RateGuard SDKs

Supported languages:

- [Node.js](https://www.npmjs.com/package/@faizahmd2/rateguard-sdk)
- [Python](https://pypi.org/project/rateguard-sdk)

---

## API

### Check Rate Limit

```http
POST /v1/check
Authorization: Bearer <API_TOKEN>
Content-Type: application/json
```

Request:

```json
{
  "service": "payments",
  "resource": "create-payment",
  "key": "account:123"
}
```

Allowed response:

```json
{
  "decision": "ALLOW",
  "limit": 5,
  "remaining": 4,
  "retry_after_ms": 0
}
```

Denied response:

```json
{
  "decision": "DENY",
  "limit": 5,
  "remaining": 0,
  "retry_after_ms": 3200
}
```

## Local Development

* Go 1.26+
* Node.js 22+
* Redis
* Git

Clone the repository:

```bash
git clone https://github.com/faizahmd2/rate-guard.git
cd rate-guard
```

Create the environment file:

```bash
cp .env.example .env
```

Edit `.env` with your local configuration.

Start Redis.

### Backend

Download Go dependencies:

```bash
go mod download
```

Run RateGuard:

```bash
go run ./cmd/rateguard
```

RateGuard will be available at:

```text
http://localhost:4215
```

### Frontend

Install frontend dependencies:

```bash
cd ui
npm install
```

Run the frontend development server:

```bash
npm run dev
```

---

## Docker Development

Build the image locally:

```bash
docker build -t rateguard:local .
```

Create the environment file:

```bash
cp .env.example .env
```

Start RateGuard with Docker Compose:

```bash
docker compose up -d
```

Check the container:

```bash
docker ps
```

View logs:

```bash
docker logs -f rateguard
```

RateGuard will be available at:

```text
http://localhost:4215
```

Stop the service:

```bash
docker compose down
```

---

## Environment Variables

Create `.env` from the example file:

```bash
cp .env.example .env
```

---

## Architecture

RateGuard separates configuration management from the runtime rate-limiting path.

```text
                    ┌─────────────────────┐
                    │    Admin UI / API   │
                    └──────────┬──────────┘
                               │
                               ▼
                    ┌─────────────────────┐
                    │    Rule Service     │
                    └──────────┬──────────┘
                               │
                               ▼
                    ┌─────────────────────┐
                    │ SQLite / PostgreSQL │
                    └─────────────────────┘


Application ──POST /v1/check──► RateGuard
                                  │
                                  ▼
                              Rule Cache
                                  │
                                  ▼
                                 Redis
                                  │
                                  ▼
                          Atomic Lua Limiter
```

The database is the source of truth for rate-limit rules.

Redis handles the runtime rate-limiting state.

The Go application maintains a local rule cache and uses Redis as the distributed cache layer.

---

## Storage

SQLite is the default storage backend and is suitable for a single-instance deployment.

PostgreSQL can be used when a separate database server is preferred.

Redis is required for distributed rate-limiting state.

---

## Security

RateGuard uses API tokens for application access.

Tokens are passed using:

```http
Authorization: Bearer <API_TOKEN>
```

Only token hashes are persisted.

Admin authentication is separate from application API tokens.

---

## License

MIT
