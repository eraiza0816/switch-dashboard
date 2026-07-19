FROM golang:1.22-alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o switch-dashboard ./cmd/switch-dashboard/

FROM scratch
COPY --from=builder /build/switch-dashboard /switch-dashboard
COPY --from=builder /build/templates/ /templates/
COPY --from=builder /build/static/ /static/
EXPOSE 8080
ENTRYPOINT ["/switch-dashboard"]
