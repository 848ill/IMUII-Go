# ============================================================
# AURA Core — Multi-stage Dockerfile (Pure Go, <25 MB final)
# ============================================================

# Stage 1: Build
FROM golang:1.23-alpine AS builder

RUN apk add --no-cache ca-certificates

WORKDIR /build
COPY go.mod ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /build/bin/aura-core ./cmd/server

# Stage 2: Runtime (scratch — minimal image)
FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /build/bin/aura-core /aura-core
COPY --from=builder /build/web /web

EXPOSE 8090

ENTRYPOINT ["/aura-core"]
