# Use the official Go image 
FROM golang:1.23-alpine AS builder

# Set the working directory
WORKDIR /app

# Copy all source code including go.mod and go.sum
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
