# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk --no-cache add ca-certificates git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o qr-menu .

# Runtime stage
FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/qr-menu ./qr-menu
COPY --from=builder /app/static ./static
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/web ./web

EXPOSE 8080

ENV PORT=8080

ENTRYPOINT ["./qr-menu"]