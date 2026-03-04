# Build stage
FROM golang:1.21-alpine AS builder

# Installa dipendenze di build
RUN apk add --no-cache git

# Imposta directory di lavoro
WORKDIR /app

# Copia go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copia tutto il codice sorgente
COPY . .

# Build l'applicazione
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o qr-menu .

# Runtime stage
FROM alpine:latest

# Installa ca-certificates per HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copia il binary dalla build stage
COPY --from=builder /app/qr-menu .

# Copia la directory templates (IMPORTANTE!)
COPY --from=builder /app/templates ./templates

# Crea le directory necessarie
RUN mkdir -p storage static static/qrcodes static/images static/images/dishes

# Esponi la porta
EXPOSE 8080

# Avvia l'applicazione
CMD ["./qr-menu"]
