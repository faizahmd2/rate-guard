# RateGuard

Distributed API Rate Limiting Service.

RateGuard is a lightweight centralized rate-limiting service that applications can call before processing an operation.

```text
Application
    │
    │ POST /v1/check
    ▼
RateGuard
    │
    ├── Rule Cache
    │
    └── Redis
          │
          └── Atomic rate limiter
```

The application decides what to do with the result — continue, retry, queue, or reject the request.

## Run with Docker

The easiest way to run RateGuard is with Docker.

RateGuard is distributed as a single container image. Redis is the only external runtime dependency.

### Pull the image

```bash
docker pull ghcr.io/faizahmd2/rate-guard:latest
```

### Generate a cookie secret

```bash
openssl rand -hex 32
```

Copy the generated value for cookie secret.

### Run RateGuard

Once Redis is running:

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

If Redis is running somewhere else, replace `REDIS_HOST` with your Redis address.
For redis host can be localhost or with docker host.docker.internal or ip where redis hosted.

RateGuard will be available at:

```text
http://localhost:4215
```

Open the URL in your browser to access the admin UI.

### Check the container

```bash
docker ps
```

View logs:

```bash
docker logs -f rateguard
```

Stop RateGuard:

```bash
docker stop rateguard
```

Start it again:

```bash
docker start rateguard
```

Remove the container:

```bash
docker rm -f rateguard
```

The SQLite database is persisted through the Docker volume:

```text
rateguard_data
```

## Features

* Centralized API rate limiting
* Redis-backed distributed limiting
* Token Bucket
* Fixed Window
* Sliding Window
* SQLite by default
* PostgreSQL support
* In-memory + Redis rule cache
* Singleflight cache-miss protection
* API token authentication
* Admin UI
* Docker support
* Single binary deployment
* Embedded frontend

## Supported Algorithms

### Token Bucket

Useful for allowing controlled bursts while maintaining an average request rate.

```json
{
  "capacity": 120,
  "refill_rate": 2,
  "key_strategy": "account"
}
```

### Fixed Window

Limits requests within a fixed time window.

```json
{
  "limit": 5,
  "window_seconds": 10,
  "key_strategy": "account"
}
```

### Sliding Window

Tracks requests over a rolling time window.

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

---

# Requirements

For local development:

* Go 1.26+
* Node.js 22+
* Redis
* Git

Production environment/Docker users only need:

* Docker
* Docker Compose

---

# Local Development

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

## Backend

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

## Frontend

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

# Docker Development

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

# Environment Variables

Create `.env` from the example file:

```bash
cp .env.example .env
```

---

# Architecture

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

# Storage

SQLite is the default storage backend and is suitable for a single-instance deployment.

PostgreSQL can be used when a separate database server is preferred.

Redis is required for distributed rate-limiting state.

---

# Security

RateGuard uses API tokens for application access.

Tokens are passed using:

```http
Authorization: Bearer <API_TOKEN>
```

Only token hashes are persisted.

Admin authentication is separate from application API tokens.

---

# SDKs

Official client SDKs are maintained separately:

**RateGuard SDKs**

https://github.com/faizahmd2/rate-guard-sdk

Currently available:

* Node.js
* Python

---

# License

MIT
