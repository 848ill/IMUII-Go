# ============================================================
# AURA Core — HF Spaces Dockerfile (root-level wrapper)
# ============================================================

FROM golang:1.23-alpine AS builder

RUN apk add --no-cache ca-certificates

WORKDIR /build
COPY aura-core/go.mod ./
RUN go mod download
COPY aura-core/ .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /build/bin/aura-core ./cmd/server

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /build/bin/aura-core /aura-core
COPY --from=builder /build/web /web

ENV PORT=7860
EXPOSE 7860

ENTRYPOINT ["/aura-core"]
