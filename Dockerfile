FROM golang:1.25.0-alpine AS builder

# Build Stage
RUN apk update && apk add --no-cache git ca-certificates tzdata
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o openhouse-2025-api ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/
COPY --from=builder /app/openhouse-2025-api .
COPY --from=builder /app/internal/templates ./internal/templates
COPY --from=builder /app/db ./db
RUN mkdir -p /app/uploads
EXPOSE 8080
CMD ["./openhouse-2025-api"]