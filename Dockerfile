FROM oven/bun:1 AS frontend
WORKDIR /app/frontend
COPY frontend/package.json ./
RUN bun install
COPY frontend/ ./
RUN bun run build

FROM golang:1.26-bookworm AS builder
RUN apt-get update && apt-get install -y --no-install-recommends gcc g++ libc6-dev && rm -rf /var/lib/apt/lists/*
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /app/static/dist/ ./static/dist/
RUN CGO_ENABLED=1 go build -o switch-dashboard ./cmd/switch-dashboard/

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && rm -rf /var/lib/apt/lists/*
COPY --from=builder /build/switch-dashboard /switch-dashboard
COPY --from=builder /build/templates/ /templates/
COPY --from=builder /build/static/ /static/
EXPOSE 8081
VOLUME ["/data"]
ENTRYPOINT ["/switch-dashboard"]
CMD ["-d", "/data"]
