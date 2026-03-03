# Build stage
FROM golang:1.24 AS builder

WORKDIR /app

COPY . .
RUN GOSUMDB=off go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o qr-menu ./

# Runtime stage
FROM gcr.io/distroless/static:nonroot

WORKDIR /app
COPY --from=builder /app/qr-menu /app/qr-menu

EXPOSE 8080
USER nonroot:nonroot

ENTRYPOINT ["/app/qr-menu"]
