# Build stage
FROM golang:1.24 AS builder

WORKDIR /app

# Copy only go.mod and go.sum first for better layer caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o qr-menu

# Runtime stage
FROM gcr.io/distroless/base-debian12

WORKDIR /app
COPY --from=builder /app/qr-menu /app/qr-menu

EXPOSE 8080

ENTRYPOINT ["/app/qr-menu"]
