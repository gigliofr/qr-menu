# QR Menu System 🍽️

Sistema completo per la creazione e gestione di menu digitali accessibili tramite QR code.

## 📋 Caratteristiche Principali

- **🎨 Interfaccia Web Intuitiva**: Crea e gestisci menu attraverso un'interfaccia web facile da utilizzare
- **📱 QR Code Automatico**: Genera automaticamente QR code per ogni menu completato
- **🌐 Menu Pubblici Responsive**: Menu ottimizzati per la visualizzazione su dispositivi mobili
- **💾 Persistenza Dati**: Salvataggio automatico dei menu in file JSON
- **🔗 API REST**: Endpoints API per integrazione con altri sistemi
- **⚡ Real-time**: Aggiornamenti immediati e visualizzazione istantanea

## 🚀 Avvio Rapido

### Prerequisiti
- Go 1.21 o superiore
- Connessione internet per il download delle dipendenze

### Installazione e Avvio

1. **Naviga nella directory del progetto**:
   ```bash
   cd qr-menu
   ```

2. **Scarica le dipendenze**:
   ```bash
   go mod download
   ```

3. **Avvia il server**:
   ```bash
   go run main.go
   ```

4. **Accedi all'applicazione**:
   - Interfaccia Admin: http://localhost:8080/admin
   - API: http://localhost:8080/api/

### Configurazione Porta

Per cambiare la porta del server, imposta la variabile d'ambiente `PORT`:

```bash
# Windows
set PORT=3000
go run main.go

# Linux/Mac
PORT=3000 go run main.go
```

## 📖 Guida all'Uso

### 1. Creare un Nuovo Menu

1. Accedi all'interfaccia admin: http://localhost:8080/admin
2. Clicca su "➕ Crea Nuovo Menu"
3. Compila i dettagli del ristorante e del menu
4. Aggiungi categorie e piatti
5. Salva il menu

### 2. Completare un Menu

1. Dalla dashboard admin, clicca "✏️ Modifica" sul menu desiderato
2. Verifica che tutte le informazioni siano corrette
3. Clicca "🎯 Completa Menu e Genera QR Code"
4. Il QR code verrà generato automaticamente

### 3. Condividere il Menu

- **QR Code**: Stampa il QR code generato per permettere ai clienti di scansionarlo
- **URL Diretto**: Condividi l'URL del menu pubblico
- **Visualizzazione**: Il menu è ottimizzato per dispositivi mobili

## 🔌 API Endpoints

### Menu Management
- `GET /api/menus` - Lista tutti i menu
- `GET /api/menu/{id}` - Ottieni un menu specifico
- `POST /api/menu` - Crea un nuovo menu (JSON)
- `POST /api/menu/{id}/generate-qr` - Genera QR code per un menu

### Esempio Creazione Menu via API

```bash
curl -X POST http://localhost:8080/api/menu \
  -H "Content-Type: application/json" \
  -d '{
    "restaurant_id": "Ristorante esempio",
    "name": "Menu della Casa",
    "description": "I nostri piatti tradizionali",
    "categories": [
      {
        "id": "cat1",
        "name": "Antipasti",
        "description": "Per iniziare",
        "items": [
          {
            "id": "item1",
            "name": "Bruschetta",
            "description": "Pane tostato con pomodoro fresco",
            "price": 6.50,
            "category": "Antipasti",
            "available": true
          }
        ]
      }
    ]
  }'
```

## 📁 Struttura del Progetto

```
qr-menu/
├── main.go                 # Server principale
├── go.mod                  # Dipendenze Go
├── models/
│   └── menu.go            # Strutture dati
├── handlers/
│   └── handlers.go        # Gestori HTTP
├── templates/
│   ├── admin.html         # Dashboard amministrazione
│   ├── create_menu.html   # Form creazione menu
│   ├── edit_menu.html     # Form modifica menu
│   └── public_menu.html   # Visualizzazione pubblica
├── static/
│   ├── css/
│   │   └── style.css      # Stili CSS
│   ├── js/
│   │   └── script.js      # JavaScript
│   └── qrcodes/          # QR code generati
└── storage/              # File JSON dei menu
```

## 🛠️ Personalizzazione

### Modificare i Template

I template HTML si trovano in `templates/` e possono essere personalizzati:
- `admin.html` - Dashboard amministrazione
- `create_menu.html` - Form creazione menu
- `edit_menu.html` - Form modifica menu
- `public_menu.html` - Visualizzazione pubblica del menu

### Aggiungere Stili Personalizzati

Modifica il file `static/css/style.css` per personalizzare l'aspetto dell'applicazione.

### Estendere l'API

Aggiungi nuovi endpoints modificando `handlers/handlers.go` e registrandoli in `main.go`.

## 🔧 Dipendenze

- **gorilla/mux**: Router HTTP per Go
- **skip2/go-qrcode**: Libreria per la generazione di QR code
- **google/uuid**: Generazione di UUID univoci

## 📄 Formato Dati Menu

Struttura JSON di un menu completo:

```json
{
  "id": "uuid-del-menu",
  "restaurant_id": "Nome Ristorante",
  "name": "Nome Menu",
  "description": "Descrizione del menu",
  "categories": [
    {
      "id": "uuid-categoria",
      "name": "Nome Categoria",
      "description": "Descrizione categoria",
      "items": [
        {
          "id": "uuid-piatto",
          "name": "Nome Piatto",
          "description": "Descrizione piatto",
          "price": 12.50,
          "category": "Nome Categoria",
          "available": true,
          "image_url": "url-immagine-opzionale"
        }
      ]
    }
  ],
  "created_at": "2024-02-24T10:00:00Z",
  "updated_at": "2024-02-24T10:00:00Z",
  "is_completed": true,
  "qr_code_path": "static/qrcodes/menu_uuid.png",
  "public_url": "http://localhost:8080/menu/uuid"
}
```

## 🚀 Deployment

### Compilazione per Produzione

```bash
# Compila per Linux
GOOS=linux GOARCH=amd64 go build -o qr-menu-linux main.go

# Compila per Windows
GOOS=windows GOARCH=amd64 go build -o qr-menu.exe main.go

# Compila per macOS
GOOS=darwin GOARCH=amd64 go build -o qr-menu-mac main.go
```

### Variabili d'Ambiente

- `PORT`: Porta del server (default: 8080)

### Docker (Opzionale)

Crea un `Dockerfile`:

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/static ./static
EXPOSE 8080
CMD ["./main"]
```

## 🤝 Contributi

Contributi sono benvenuti! Per contribuire:

1. Fork del repository
2. Crea un branch per la tua feature
3. Commit delle modifiche
4. Push al branch
5. Apri una Pull Request

## 📝 License

Questo progetto è rilasciato sotto licenza MIT. Vedi il file LICENSE per maggiori dettagli.

## 🆘 Supporto

Per problemi o domande:
1. Controlla la documentazione sopra
2. Apri un Issue su GitHub
3. Verifica i log del server per errori

## 🔄 Aggiornamenti Futuri

Funzionalità pianificate:
- [ ] Upload immagini per i piatti
- [ ] Gestione multi-ristorante
- [ ] Traduzione multilingua
- [ ] Integrazione con sistemi di pagamento
- [ ] Analytics e statistiche
- [ ] Notifiche push per aggiornamenti menu
- [ ] Gestione ordini online

## 📞 Contatti

Sviluppato per QR Menu System - Sistema di gestione menu digitali