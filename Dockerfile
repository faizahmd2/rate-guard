# ---------------------------------------------------------
# Build frontend
# ---------------------------------------------------------

FROM node:22-alpine AS frontend-builder

WORKDIR /build/ui

COPY ui/package*.json ./
RUN npm ci

COPY ui/ ./

RUN npm run build


# ---------------------------------------------------------
# Build backend
# ---------------------------------------------------------

FROM golang:1.26-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Copy the built React application into the Go embed location.
COPY --from=frontend-builder \
    /build/internal/web/dist \
    /build/internal/web/dist

RUN CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    go build \
    -ldflags="-s -w" \
    -o /build/rateguard \
    ./cmd/rateguard


# ---------------------------------------------------------
# Final image
# ---------------------------------------------------------

FROM alpine:3.22

WORKDIR /app

RUN addgroup -S rateguard \
    && adduser -S rateguard -G rateguard \
    && mkdir -p /app/data \
    && chown -R rateguard:rateguard /app

COPY --from=builder \
    --chown=rateguard:rateguard \
    /build/rateguard \
    /app/rateguard

USER rateguard

EXPOSE 4215

CMD ["/app/rateguard"]
