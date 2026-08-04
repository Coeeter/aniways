FROM golang:1.24.5-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /out/api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/worker ./cmd/worker

# --- API ---
FROM alpine:3.20 AS api
WORKDIR /app
COPY --from=builder /out/api /app/api
COPY --chmod=755 docker/entrypoint.sh /app/entrypoint.sh
RUN adduser -D aniways - \
    && mkdir -p /app/tmp \
    && chown -R aniways:aniways /app/tmp
USER aniways
CMD ["/app/entrypoint.sh", "/app/api"]

# --- WORKER ---
FROM alpine:3.20 AS worker
WORKDIR /app
COPY --from=builder /out/worker /app/worker
COPY --chmod=755 docker/entrypoint.sh /app/entrypoint.sh
RUN adduser -D aniways
USER aniways
CMD ["/app/entrypoint.sh", "/app/worker"]
