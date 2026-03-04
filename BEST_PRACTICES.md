# 🛡️ Best Practices: Sicurezza, Performance e Robustezza

## 🔐 Sicurezza

### Autenticazione Multi-Ristorante

#### ✅ Implementato
- **Separazione User/Restaurant**: Un account può gestire più ristoranti
- **Session con UserID + RestaurantID**: Traccia sia l'utente che il ristorante attivo
- **Ownership Verification**: `getCurrentRestaurant()` verifica che il ristorante appartenga all'utente
- **GDPR Compliance**: Consensi privacy e marketing salvati nel database con timestamp

#### 🚀 Raccomandazioni Aggiuntive

**1. Implementare JWT per API**
```go
// Per API REST, usa JWT invece di session cookie
type JWTClaims struct {
    UserID       string   `json:"user_id"`
    RestaurantID string   `json:"restaurant_id,omitempty"`
    Roles        []string `json:"roles"`
    jwt.StandardClaims
}
```

**2. Rate Limiting Avanzato**
```go
// Già implementato in middleware/security.go
// Attiva in routes.go:
r.Use(middleware.RateLimitByUser(100, time.Minute))
```

**3. Password Policy**
```go
// In handlers/auth.go RegisterHandler
func validatePassword(password string) []string {
    var errors []string
    if len(password) < 12 { // ⭐ Aumenta da 8 a 12
        errors = append(errors, "Password minimo 12 caratteri")
    }
    if !regexp.MustCompile(`[A-Z]`).MatchString(password) {
        errors = append(errors, "Deve contenere maiuscola")
    }
    if !regexp.MustCompile(`[a-z]`).MatchString(password) {
        errors = append(errors, "Deve contenere minuscola")
    }
    if !regexp.MustCompile(`[0-9]`).MatchString(password) {
        errors = append(errors, "Deve contenere numero")
    }
    if !regexp.MustCompile(`[!@#$%^&*]`).MatchString(password) {
        errors = append(errors, "Deve contenere simbolo speciale")
    }
    return errors
}
```

**4. Audit Log Completo**
```go
// Logga TUTTE le operazioni sensibili
type AuditEvent struct {
    Timestamp    time.Time
    UserID       string
    RestaurantID string
    Action       string // "create_menu", "delete_menu", "switch_restaurant"
    IPAddress    string
    UserAgent    string
    Resource     string // ID della risorsa modificata
    Success      bool
}

// Salva in MongoDB per analisi post-mortem
db.audit_events.insertOne(event)
```

**5. 2FA (Two-Factor Authentication)**
```go
// Per utenti business, implementa TOTP
import "github.com/pquerna/otp/totp"

func enableTOTP(userID string) (string, error) {
    key, err := totp.Generate(totp.GenerateOpts{
        Issuer:      "QR Menu",
        AccountName: userID,
    })
    // Salva key.Secret() in User.TOTPSecret (encrypted)
    return key.URL(), nil
}
```

### Protezione CSRF

#### ✅ Implementato
- **Middleware CSRF**: `middleware/security.go` - CSRFProtectionMiddleware
- **Token Generation**: In handlers/auth.go

#### 🚀 Da Completare
```go
// 1. Genera token CSRF in ogni form
// templates/admin.html
<input type="hidden" name="csrf_token" value="{{.CSRFToken}}">

// 2. Valida nel middleware
func validateCSRFToken(token string, sessionID string) bool {
    // Verifica che il token sia valido per questa sessione
    // Usa HMAC-SHA256 con secret key
    mac := hmac.New(sha256.New, []byte(os.Getenv("CSRF_SECRET")))
    mac.Write([]byte(sessionID))
    expected := base64.URLEncoding.EncodeToString(mac.Sum(nil))
    return hmac.Equal([]byte(token), []byte(expected))
}
```

### SQL Injection Protection

#### ✅ Implementato
- **MongoDB Prepared Statements**: Tutti i query usano BSON che previene injection
- **Input Sanitization**: `sanitizeInput()` in handlers/handlers.go

#### 🚀 Verifica
```go
// ✅ CORRETTO (usa BSON)
db.restaurants.FindOne(ctx, bson.M{"_id": restaurantID})

// ❌ SBAGLIATO (mai fare così)
// query := fmt.Sprintf("SELECT * FROM users WHERE id='%s'", userID)
```

### XSS Protection

#### ✅ Implementato
- **Template Auto-Escape**: Go templates escapano automaticamente HTML
- **Content-Security-Policy**: Header impostato in `setSecurityHeaders()`

#### 🚀 Migliora CSP
```go
func setSecurityHeaders(w http.ResponseWriter) {
    w.Header().Set("Content-Security-Policy", 
        "default-src 'self'; "+
        "script-src 'self' 'unsafe-inline' fonts.googleapis.com; "+
        "style-src 'self' 'unsafe-inline' fonts.googleapis.com; "+
        "img-src 'self' data: https:; "+
        "font-src 'self' fonts.gstatic.com; "+
        "frame-ancestors 'none'; "+
        "base-uri 'self'")
}
```

---

## ⚡ Performance

### Database Indexing

#### ✅ Implementato
Script migrazione crea automaticamente:

```javascript
// users
{ username: 1 } UNIQUE
{ email: 1 } UNIQUE
{ is_active: 1, last_login: -1 }

// restaurants
{ owner_id: 1, is_active: 1 }
{ owner_id: 1, created_at: -1 }

// sessions
{ user_id: 1, last_accessed: -1 }
{ last_accessed: 1 } TTL (30 giorni)

// menus
{ restaurant_id: 1, is_active: 1 }
```

#### 🚀 Monitora Performance
```javascript
// Abilita MongoDB profiler
db.setProfilingLevel(1, { slowms: 100 }); // Logga query >100ms

// Analizza query lente
db.system.profile.find({ millis: { $gt: 100 }})
    .sort({ millis: -1 })
    .limit(10);

// Verifica uso indici
db.restaurants.find({ owner_id: "user_123" }).explain("executionStats");
// Deve mostrare: "stage": "IXSCAN" (Index Scan, non COLLSCAN)
```

### Caching

#### 🚀 Implementa Redis per Session Cache
```go
import "github.com/go-redis/redis/v8"

var redisClient *redis.Client

func init() {
    redisClient = redis.NewClient(&redis.Options{
        Addr: os.Getenv("REDIS_URL"),
        DB:   0,
    })
}

func getSessionFromCache(sessionID string) (*models.Session, error) {
    ctx := context.Background()
    data, err := redisClient.Get(ctx, "session:"+sessionID).Result()
    if err == redis.Nil {
        // Cache miss, leggi da MongoDB
        session, err := db.MongoInstance.GetSessionByID(ctx, sessionID)
        if err == nil {
            // Salva in cache (TTL 1 ora)
            cacheSession(sessionID, session, time.Hour)
        }
        return session, err
    }
    // Cache hit
    var session models.Session
    json.Unmarshal([]byte(data), &session)
    return &session, nil
}
```

#### 🚀 Cache Menu Pubblici
```go
// Menu pubblici non cambiano spesso, cachali per 5 minuti
func PublicMenuHandler(w http.ResponseWriter, r *http.Request) {
    menuID := mux.Vars(r)["id"]
    
    // Try cache
    cacheKey := "public_menu:" + menuID
    if cached, err := redisClient.Get(ctx, cacheKey).Result(); err == nil {
        w.Header().Set("Content-Type", "text/html; charset=utf-8")
        w.Header().Set("X-Cache", "HIT")
        w.Write([]byte(cached))
        return
    }
    
    // Cache miss, genera e salva
    menu, _ := db.MongoInstance.GetMenuByID(ctx, menuID)
    html := renderMenuHTML(menu)
    redisClient.Set(ctx, cacheKey, html, 5*time.Minute)
    
    w.Header().Set("X-Cache", "MISS")
    w.Write([]byte(html))
}
```

### Connection Pooling

#### ✅ Implementato
MongoDB driver Go gestisce automaticamente connection pool

#### 🚀 Ottimizza
```go
// In db/mongo.go Connect()
clientOptions := options.Client().
    ApplyURI(mongoURI).
    SetMaxPoolSize(50).        // ⭐ Max 50 connessioni
    SetMinPoolSize(10).        // ⭐ Min 10 connessioni always open
    SetMaxConnIdleTime(60 * time.Second). // Chiudi idle dopo 60s
    SetServerSelectionTimeout(5 * time.Second)
```

### Image Optimization

#### 🚀 Implementa Resize Automatico
```go
import "github.com/disintegration/imaging"

func optimizeImage(imgPath string) error {
    img, err := imaging.Open(imgPath)
    if err != nil {
        return err
    }
    
    // Resize mantenendo aspect ratio
    resized := imaging.Fit(img, 800, 600, imaging.Lanczos)
    
    // Salva come JPEG con qualità 85
    return imaging.Save(resized, imgPath, imaging.JPEGQuality(85))
}
```

### Lazy Loading

```html
<!-- In templates, usa loading="lazy" per immagini -->
<img src="/static/menu_images/{{.Image}}" 
     loading="lazy" 
     alt="{{.Name}}">
```

---

## 🔧 Robustezza

### Error Handling

#### ✅ Buone Pratiche Già Implementate
```go
// Context timeout per evitare query infinite
ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
defer cancel()

// Gestione errori con log strutturato
if err != nil {
    log.Printf("Errore recupero menu: %v", err)
    http.Error(w, "Errore interno", http.StatusInternalServerError)
    return
}
```

#### 🚀 Implementa Error Types
```go
// errors/errors.go
type AppError struct {
    Code    string // "RESTAURANT_NOT_FOUND", "UNAUTHORIZED"
    Message string
    Err     error
    Context map[string]interface{}
}

func (e *AppError) Error() string {
    return fmt.Sprintf("%s: %s (context: %v)", e.Code, e.Message, e.Context)
}

// Uso
func GetRestaurantByID(id string) (*Restaurant, error) {
    restaurant, err := db.FindOne(...)
    if err == mongo.ErrNoDocuments {
        return nil, &AppError{
            Code:    "RESTAURANT_NOT_FOUND",
            Message: "Ristorante non trovato",
            Context: map[string]interface{}{"id": id},
        }
    }
    return restaurant, nil
}
```

### Circuit Breaker

#### 🚀 Proteggi Servizi Esterni
```go
import "github.com/sony/gobreaker"

var cb *gobreaker.CircuitBreaker

func init() {
    cb = gobreaker.NewCircuitBreaker(gobreaker.Settings{
        Name:        "MongoDB",
        MaxRequests: 3,
        Interval:    time.Minute,
        Timeout:     10 * time.Second,
    })
}

func GetMenuWithCircuitBreaker(menuID string) (*Menu, error) {
    result, err := cb.Execute(func() (interface{}, error) {
        return db.MongoInstance.GetMenuByID(ctx, menuID)
    })
    
    if err != nil {
        // Circuit aperto, usa fallback
        return getCachedMenu(menuID), nil
    }
    
    return result.(*Menu), nil
}
```

### Health Checks

#### 🚀 Endpoint Health Completo
```go
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()
    
    health := map[string]interface{}{
        "status": "healthy",
        "timestamp": time.Now().Unix(),
        "version": "2.0.0",
    }
    
    // Check MongoDB
    if err := db.MongoInstance.Ping(ctx); err != nil {
        health["status"] = "unhealthy"
        health["mongodb"] = "disconnected"
    } else {
        health["mongodb"] = "connected"
    }
    
    // Check Redis (se implementato)
    if redisClient != nil {
        if err := redisClient.Ping(ctx).Err(); err != nil {
            health["redis"] = "disconnected"
        } else {
            health["redis"] = "connected"
        }
    }
    
    statusCode := http.StatusOK
    if health["status"] == "unhealthy" {
        statusCode = http.StatusServiceUnavailable
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    json.NewEncoder(w).Encode(health)
}

// Route: GET /health
// Railway/Cloud Run usano questo per healthchecks
```

### Graceful Shutdown

#### 🚀 Chiudi Connessioni Pulitamente
```go
// main.go
func main() {
    // ... setup ...
    
    server := &http.Server{
        Addr:    ":" + port,
        Handler: router,
    }
    
    // Channel per segnali OS
    stop := make(chan os.Signal, 1)
    signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
    
    // Avvia server in goroutine
    go func() {
        log.Printf("Server listening on :%s", port)
        if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Server error: %v", err)
        }
    }()
    
    // Attendi segnale shutdown
    <-stop
    log.Println("Shutting down server...")
    
    // Graceful shutdown (max 30s)
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    if err := server.Shutdown(ctx); err != nil {
        log.Printf("Server forced to shutdown: %v", err)
    }
    
    // Chiudi MongoDB
    if db.MongoInstance != nil {
        db.MongoInstance.Disconnect()
    }
    
    log.Println("Server exited")
}
```

### Request Timeout

```go
// middleware/timeout.go
func TimeoutMiddleware(timeout time.Duration) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            ctx, cancel := context.WithTimeout(r.Context(), timeout)
            defer cancel()
            
            done := make(chan struct{})
            go func() {
                next.ServeHTTP(w, r.WithContext(ctx))
                close(done)
            }()
            
            select {
            case <-done:
                return
            case <-ctx.Done():
                http.Error(w, "Request timeout", http.StatusGatewayTimeout)
            }
        })
    }
}

// Usa in routes.go
r.Use(middleware.TimeoutMiddleware(30 * time.Second))
```

---

## 📊 Monitoring

### Metriche da Tracciare

#### Application Metrics
```go
import "github.com/prometheus/client_golang/prometheus"

var (
    httpRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total HTTP requests",
        },
        []string{"method", "endpoint", "status"},
    )
    
    httpRequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "HTTP request duration",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "endpoint"},
    )
)

func init() {
    prometheus.MustRegister(httpRequestsTotal)
    prometheus.MustRegister(httpRequestDuration)
}

// Prometheus endpoint: /metrics
```

#### Business Metrics
```go
// Traccia operazioni business
var (
    restaurantsCreated = prometheus.NewCounter(...)
    menusCreated = prometheus.NewCounter(...)
    activeUsers = prometheus.NewGauge(...)
    menuViews = prometheus.NewCounterVec(...)
)
```

### Log Aggregation

#### Structured Logging
```go
import "go.uber.org/zap"

logger, _ := zap.NewProduction()
defer logger.Sync()

logger.Info("User login",
    zap.String("user_id", userID),
    zap.String("ip", ip),
    zap.Int("restaurant_count", len(restaurants)),
)
```

### Alerting

#### Setup Alerts (esempio con Prometheus + Alertmanager)
```yaml
# alerting_rules.yml
groups:
  - name: qr_menu_alerts
    rules:
      - alert: HighErrorRate
        expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.05
        for: 5m
        annotations:
          summary: "High error rate detected"
          
      - alert: SlowQueries
        expr: mongodb_query_duration_seconds > 1
        for: 2m
        annotations:
          summary: "MongoDB queries slow"
```

---

## 🎯 Checklist Produzione

### Pre-Deploy
- [ ] Tutte le variabili d'ambiente configurate
- [ ] Certificati SSL/TLS validi
- [ ] Backup database recente
- [ ] Indici MongoDB creati
- [ ] Health check endpoint funzionante
- [ ] Log strutturati abilitati
- [ ] Rate limiting configurato
- [ ] CORS configurato correttamente

### Post-Deploy
- [ ] Health check ritorna 200 OK
- [ ] Login funziona
- [ ] Registrazione con GDPR funziona
- [ ] Multi-restaurant selection funziona
- [ ] Menu creation/edit funziona
- [ ] Analytics tracciati
- [ ] Nessun errore 500 nei log
- [ ] Performance <500ms per pagina

### Monitoring Continuo
- [ ] Monitoraggio errori (Sentry/Rollbar)
- [ ] Monitoraggio performance (New Relic/Datadog)
- [ ] Monitoraggio uptime (Pingdom/UptimeRobot)
- [ ] Alert email/Slack configurati
- [ ] Backup automatici giornalieri

---

## 📚 Risorse

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [Go Security Best Practices](https://github.com/OWASP/Go-SCP)
- [MongoDB Performance Best Practices](https://www.mongodb.com/docs/manual/administration/analyzing-mongodb-performance/)
- [Cloud Run Best Practices](https://cloud.google.com/run/docs/best-practices)
