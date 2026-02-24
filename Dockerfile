# Build stage
FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o qr-menu ./

# Runtime stage
FROM gcr.io/distroless/static:nonroot

WORKDIR /app
COPY --from=builder /app/qr-menu /app/qr-menu

EXPOSE 8080
USER nonroot:nonroot

ENTRYPOINT ["/app/qr-menu"]
