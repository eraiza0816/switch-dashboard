FROM oven/bun:1 AS frontend
WORKDIR /app/frontend
COPY frontend/package.json ./
RUN bun install
COPY frontend/ ./
RUN bun run build

FROM golang:1.26.5-alpine AS builder
RUN apk add --no-cache gcc musl-dev
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /app/static/dist/ ./static/dist/
RUN CGO_ENABLED=1 go build -ldflags="-linkmode external -extldflags -static" -o switch-dashboard ./cmd/switch-dashboard/

FROM scratch
COPY --from=builder /build/switch-dashboard /switch-dashboard
COPY --from=builder /build/templates/ /templates/
COPY --from=builder /build/static/ /static/
EXPOSE 8081
ENTRYPOINT ["/switch-dashboard"]
