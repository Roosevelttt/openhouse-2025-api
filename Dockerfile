FROM golang:1.25.0-alpine AS builder

RUN apk update && apk add --no-cache git ca-certificates tzdata make bash
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
RUN go install github.com/pressly/goose/v3/cmd/goose@latest
COPY . .
COPY Makefile .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o openhouse-2025-api ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -o division-seeder ./cmd/divisionseeder

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata make bash
WORKDIR /root/
COPY --from=builder /app/openhouse-2025-api .
COPY --from=builder /go/bin/goose /usr/local/bin/
COPY --from=builder /app/division-seeder /usr/local/bin/
COPY --from=builder /app/internal/templates ./internal/templates
COPY --from=builder /app/db ./db
COPY --from=builder /app/Makefile .
RUN mkdir -p /app/uploads
EXPOSE 8080
CMD ["./openhouse-2025-api"]