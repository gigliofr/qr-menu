# Use the official Go image 
FROM golang:1.24-alpine AS builder

# Set the working directory
WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod go.sum ./

# Download dependencies - generates go.sum entries
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN go build -o qr-menu .

# Runtime stage - use alpine runtime
FROM alpine:latest

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/qr-menu /app/qr-menu

EXPOSE 8080

# Run the application
CMD ["./qr-menu"]
