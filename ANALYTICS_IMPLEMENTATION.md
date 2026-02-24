# 📊 Analytics Dashboard - Documentazione Completa

## 🎯 Panoramica

Il sistema di Analytics Dashboard fornisce monitoraggio in tempo reale e analisi dettagliate del comportamento degli utenti per ogni ristorante. Il sistema traccia:

- **Visualizzazioni** di menu e piatti
- **Scansioni QR code** con geolocalizzazione
- **Condivisioni** per piattaforma (WhatsApp, Telegram, Facebook, Twitter)
- **Dati device** (tipo dispositivo, browser, sistema operativo, paese)
- **Trend temporali** (orari di picco, trend giornalieri)

## 🏗️ Architettura del Sistema

### Componenti Principali

```
analytics/
├── analytics.go          # Core system - tracking e aggregazione dati
├── api.go               # (se necessario) Endpoint API per analytics
└── persistence.go       # (se necessario) Salvataggio dati su disco

handlers/
├── handlers.go          # Integrazione tracking negli handler
└── auth.go             # Autenticazione per dashboard

models/
└── menu.go             # Strutture dati per eventi

templates/
└── analytics_dashboard.html  # Dashboard UI con Chart.js
```

### Flusso dei Dati

```
User Action (view/share/QR scan)
    ↓
Handler HTTP (PublicMenuHandler, ShareMenuHandler, GetActiveMenuHandler)
    ↓
Track Event (ViewEvent, ShareEvent, QRScanEvent)
    ↓
Analytics.GetAnalytics().Track*()
    ↓
RestaurantStats (aggregazione in memoria)
    ↓
GetDashboardData() → Template Rendering
    ↓
HTML + Chart.js Visualization
```

## 🔄 Flusso di Tracking

### 1. Visualizzazione Menu (`PublicMenuHandler`)

```go
// Triggered quando l'utente accede a /menu/{id}
event := analytics.ViewEvent{
    RestaurantID: menu.RestaurantID,
    MenuID:       menuID,
    Timestamp:    time.Now(),
    UserIP:       clientIP,
    UserAgent:    userAgent,
    Referrer:     r.Header.Get("Referer"),
}
analytics.GetAnalytics().TrackView(event)
```

**Dati Raccolti:**
- Menu ID visualizzato
- Device Type (Mobile/Desktop/Tablet)
- Browser e Sistema Operativo
- Paese (da IP)
- Referrer (provenienza)

### 2. Scansione QR Code (`GetActiveMenuHandler`)

```go
// Triggered quando l'utente scansiona il QR code via /r/{username}
event := analytics.QRScanEvent{
    RestaurantID: restaurant.ID,
    MenuID:       restaurant.ActiveMenuID,
    Timestamp:    time.Now(),
    UserIP:       clientIP,
    UserAgent:    userAgent,
}
analytics.GetAnalytics().TrackQRScan(event)
```

**Dati Raccolti:**
- Menu attivo scansionato
- Localizzazione (derivabile dall'IP)
- Device dell'utente
- Timestamp scansione

### 3. Condivisione Menu (`ShareMenuHandler` e `TrackShareHandler`)

```go
// Opzione 1: Semplice registrazione accesso share page
event := analytics.ShareEvent{
    RestaurantID: menu.RestaurantID,
    MenuID:       menuID,
    Platform:     "share_page",
    Timestamp:    time.Now(),
    UserIP:       clientIP,
    UserAgent:    userAgent,
}

// Opzione 2: Tracking specifico per piattaforma (POST a /api/track/share)
// {
//   "menu_id": "...",
//   "platform": "whatsapp|telegram|facebook|twitter|copy_link"
// }
```

**Piattaforme Supportate:**
- WhatsApp
- Telegram
- Facebook
- Twitter
- Copia Link

## 📊 API Endpoints

### GET `/admin/analytics`
Mostra la dashboard analytics interattiva (autenticato).

**Query Parameters:**
- `days`: 7, 30 o 90 (default: 7)

**Esempio:**
```
GET /admin/analytics?days=30
```

### GET `/api/analytics`
Restituisce i dati analytics in JSON (autenticato).

**Query Parameters:**
- `days`: 7, 30 o 90 (default: 7)

**Response:**
```json
{
    "total_views": 1234,
    "unique_views": 567,
    "total_shares": 89,
    "qr_scans": 45,
    "daily_trend": [
        {"date": "2024-01-01", "views": 100, "qr_scans": 10},
        ...
    ],
    "device_stats": {
        "Mobile": 800,
        "Desktop": 300,
        "Tablet": 134
    },
    "hourly_stats": {
        "12": 150,
        "13": 200,
        ...
    },
    "popular_items": [
        {
            "item_id": "...",
            "item_name": "Pizza Margherita",
            "views": 120,
            "price": 12.50
        }
    ],
    "share_breakdown": {
        "whatsapp": 45,
        "telegram": 20,
        "facebook": 15,
        "twitter": 5,
        "copy_link": 4
    }
}
```

### POST `/api/track/share`
Traccia una condivisione specifica per piattaforma (pubblico).

**Request Body:**
```json
{
    "menu_id": "abc123",
    "platform": "whatsapp|telegram|facebook|twitter|copy_link"
}
```

**Response:**
```json
{
    "status": "success"
}
```

## 🎨 Dashboard Features

### Stats Cards
4 card principali che mostrano:
1. **Visualizzazioni Totali** - Numero totale access ai menu
2. **Condivisioni** - Numero totale di share
3. **Scansioni QR** - Numero di scansioni QR code
4. **Visitatori Unici** - Device/IP unici

### Grafici
- **Trend Visualizzazioni** (Line Chart): Trend temporale di views e QR scans
- **Dispositivi** (Doughnut Chart): Distribuzione mobile/desktop/tablet
- **Orari di Picco** (Bar Chart): Traffico per ora del giorno

### Insights Cards
- **Top Paesi** - Provenienze geografiche dei visitatori
- **Browser** - Browser più utilizzati
- **Condivisioni per Piattaforma** - Breakdown WhatsApp/Telegram/etc
- **Trend Orari** - Quando i visitatori sono più attivi

## 🔐 Autenticazione

### Requisiti
- Admin deve essere loggato per accedere a `/admin/analytics`
- Solo dati del proprio ristorante sono visibili
- API endpoints richiedono autenticazione sessione

### Verifica Sessione
```go
session, err := getSessionFromRequest(r)
if err != nil || session.RestaurantID == "" {
    http.Redirect(w, r, "/login", http.StatusFound)
    return
}
// Usa session.RestaurantID per filtrare i dati
```

## 💾 Persistenza Dati

Attualmente i dati analytics sono memorizzati **in memoria** (tipo `map[string]*RestaurantStats`).

### Per Implementare Persistenza:
```go
// Opzione 1: JSON su disco
analytics.SaveToFile("analytics/data.json")
analytics.LoadFromFile("analytics/data.json")

// Opzione 2: Database (PostgreSQL/MongoDB)
// Salvare ogni evento in tabella events
// Query aggregati al momento della dashboard

// Opzione 3: Cloud Analytics
// Google Analytics, Mixpanel, Segment, etc.
```

## 🛠️ Implementazione Dettagliata

### Step 1: Inizializzazione (main.go)
```go
import "qr-menu/analytics"

func main() {
    // Il singleton globale si inizializza automaticamente
    _ = analytics.GetAnalytics()
}
```

### Step 2: Tracking nei Handlers

**Aggiunta tracking nel PublicMenuHandler:**
```go
go func() {
    event := analytics.ViewEvent{
        RestaurantID: menu.RestaurantID,
        MenuID:       menuID,
        Timestamp:    time.Now(),
        UserIP:       getClientIP(r),
        UserAgent:    r.Header.Get("User-Agent"),
    }
    analytics.GetAnalytics().TrackView(event)
}()
```

### Step 3: Esposizione Dashboard

**Route aggiunta in main.go:**
```go
r.HandleFunc("/admin/analytics", handlers.RequireAuth(handlers.AnalyticsDashboardHandler)).Methods("GET")
r.HandleFunc("/api/analytics", handlers.RequireAuth(handlers.AnalyticsAPIHandler)).Methods("GET")
r.HandleFunc("/api/track/share", handlers.TrackShareHandler).Methods("POST")
```

### Step 4: Rendering Template

**AnalyticsDashboardHandler:**
```go
dashboardData := analytics.GetAnalytics().GetDashboardData(session.RestaurantID, days)
renderTemplate(w, "analytics_dashboard", struct {
    Restaurant *models.Restaurant
    Analytics  map[string]interface{}
}{
    Restaurant: restaurant,
    Analytics:  dashboardData,
})
```

## 📈 Expected Outputs

### RestaurantStats Structure
```go
type RestaurantStats struct {
    RestaurantID     string            // ID ristorante
    TotalViews       int               // Visite totali
    UniqueViews      int               // IP/Device unici
    DailyViews       map[string]int    // Views per giorno (YYYY-MM-DD)
    HourlyViews      map[int]int       // Views per ora (0-23)
    DeviceTypes      map[string]int    // Mobile/Desktop/Tablet
    OperatingSystems map[string]int    // iOS/Android/Windows/Mac
    Browsers         map[string]int    // Chrome/Safari/Firefox
    Countries        map[string]int    // IT/FR/DE/etc
    MenuViews        map[string]int    // Views per menu
    PopularItems     []PopularItem     // Top piatti per views
    ShareStats       ShareStats        // Breakdown condivisioni
    QRCodeScans      map[string]int    // Scansioni per menu
    LastUpdated      time.Time         // Ultimo aggiornamento
}
```

## 🚀 Performance Considerations

### Memory Usage
- Circa **5KB per visitatore univoco**
- 1000 visitatori al mese = ~5MB
- Pulire dati vecchi ogni X giorni per mantenere performance

### Optimization Tips
1. **Cache Dashboard**: Memorizzare HTML renderizzato per 5 minuti
2. **Batch Updates**: Aggregare eventi in batch instad of real-time
3. **Archivio Storico**: Spostare dati vecchi > 90 giorni
4. **Limit Query**: Filtrare per max 90 giorni negli aggregati

## 📱 Tracking Specifici per Device

### Mobile Detection (da User-Agent)
```go
deviceType := "Mobile"
if strings.Contains(userAgent, "iPad") {
    deviceType = "Tablet"
} else if strings.Contains(userAgent, "Android|iOS") {
    deviceType = "Mobile"
} else {
    deviceType = "Desktop"
}
```

### Browser Detection
```
Chrome, Firefox, Safari, Edge, Opera
(Implementato in analytics.go)
```

### OS Detection
```
iOS, Android, Windows, macOS, Linux
(Implementato in analytics.go)
```

## 🔄 Aggiornamenti Futuri

### Phase 2: Dashboard Avanzata
- [ ] Export PDF report
- [ ] Export CSV listing
- [ ] Grafici comparativi (week-over-week, month-over-month)
- [ ] Heatmap del traffico orario
- [ ] Mappa geografica con visitatori per paese
- [ ] Email report settimanale/mensile

### Phase 3: Funzionalità Avanzate
- [ ] Machine Learning anomaly detection
- [ ] Predictive analytics (che giorni/orari saranno più traffici)
- [ ] Conversion funnel (view → share → sale)
- [ ] A/B testing integration
- [ ] Custom event tracking

### Phase 4: Integrazione Esterna
- [ ] Google Analytics integration
- [ ] Segment.com integration
- [ ] Metabase dashboard
- [ ] Grafana monitoring
- [ ] Webhook alerts (traffic spikes, anomalies)

## 📝 Testing

### Test Tracking manuale
```bash
# Apri il menu
curl "http://localhost:8080/menu/abc123"

# Scansiona QR (simula)
curl "http://localhost:8080/r/username"

# Track share
curl -X POST "http://localhost:8080/api/track/share" \
  -H "Content-Type: application/json" \
  -d '{"menu_id":"abc123","platform":"whatsapp"}'

# Visualizza dashboard
curl "http://localhost:8080/admin/analytics?days=7"

# Ottieni JSON
curl "http://localhost:8080/api/analytics?days=30"
```

## 🐛 Debugging

### Log Analytics Events
```go
// Aggiungi logging in analytics.go
logger.Info("Track View", map[string]interface{}{
    "restaurant": event.RestaurantID,
    "menu": event.MenuID,
    "device": event.DeviceType,
})
```

### Check In-Memory Stats
```go
stats := analytics.GetAnalytics().GetRestaurantStats("restaurant-123")
fmt.Printf("%+v\n", stats)
```

## ✅ Checklist Implementazione

- [x] Analytics package con singleton pattern
- [x] Event tracking (ViewEvent, ShareEvent, QRScanEvent)
- [x] Integrazione in handlers (PublicMenuHandler, GetActiveMenuHandler, ShareMenuHandler)
- [x] Dashboard HTML template con Chart.js
- [x] API endpoints GET `/api/analytics` e POST `/api/track/share`
- [x] Device detection (Mobile/Desktop/Tablet)
- [x] Aggregazione dati e statistiche
- [x] Autenticazione dashboard
- [x] Routes e middleware integrazione
- [ ] Unit tests per analytics engine
- [ ] Load testing per performance
- [ ] Persistenza dati (file o DB)
- [ ] Backup e recovery
- [ ] GDPR compliance (privacy dati utenti)

## 📞 Support

Per domande o problemi con il sistema analytics, consulta:
1. Commenti inline nel codice (`analytics/analytics.go`)
2. Template HTML (`templates/analytics_dashboard.html`)
3. Handler integration (`handlers/handlers.go`)
