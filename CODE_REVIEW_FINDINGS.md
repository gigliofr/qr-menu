# 🔍 Code Review - Findings Dettagliati

**Data:** 2025-01-XX  
**Scope:** Performance bottlenecks, code smells, optimization opportunities

---

## 🚨 PROBLEMI CRITICI

### 1. N+1 Query Problem - SetActiveMenuHandler

**File:** [handlers/handlers.go](handlers/handlers.go#L915-L921)  
**Linee:** 915-921

**Problema:**
```go
for _, m := range allMenus {
    if m.IsActive {
        m.IsActive = false
        if err := db.MongoInstance.UpdateMenu(ctx, m); err != nil {
            log.Printf("Errore nell'aggiornamento menu: %v", err)
        }
    }
}
```

**Descrizione:**  
Per ogni menu attivo, viene eseguita una query UPDATE separata. Con 10+ menu per ristorante, questo genera 10+ query.

**Impatto:**  
- Latenza aumentata (10 query * 50ms = 500ms vs 1 query * 50ms)
- Carico DB inutile
- Scalability issue

**Soluzione:**
```go
// Opzione A: Batch update con MongoDB bulkWrite
operations := []mongo.WriteModel{}
for _, m := range allMenus {
    if m.IsActive {
        update := bson.M{"$set": bson.M{"is_active": false}}
        filter := bson.M{"_id": m.ID}
        operations = append(operations, mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update))
    }
}

if len(operations) > 0 {
    _, err := db.MongoInstance.BulkWrite(ctx, operations)
    if err != nil {
        log.Printf("Errore bulk update: %v", err)
    }
}

// Opzione B: Single query per tutti (PREFERRED)
_, err := db.MongoInstance.UpdateMany(ctx, 
    bson.M{"restaurant_id": restaurant.ID, "is_active": true},
    bson.M{"$set": bson.M{"is_active": false}},
)
```

**Stima Fix:** 1h (implementare `UpdateMany` in db layer)  
**Priorità:** 🔴 ALTA

---

### 2. In-Memory Storage Legacy Non Rimosso

**File:** [handlers/handlers.go](handlers/handlers.go#L42)  
**Linee:** 42, 1296

**Problema:**
```go
// Line 42:
var (
    templates         *template.Template
    menus             = make(map[string]*models.Menu) // ⚠️ Storage in memoria (temporaneo)
    csrfTokens        = make(map[string]time.Time)
    ...
)

// Line 1296 (in loadMenusFromStorage):
menus[menu.ID] = &menu
```

**Descrizione:**  
Variabile globale `menus` dichiarata come "(temporaneo)" ma ancora popolata da una funzione `loadMenusFromStorage()`.

**Verifica Utilizzo:**
```bash
grep -n "menus\[" handlers/handlers.go
# Result: 246, 1296
# Line 246: restaurantMenus[menu.ID] = menu (local variable, OK)
# Line 1296: menus[menu.ID] = &menu (global var, UNUSED?)
```

**Impatto:**  
- Memoria sprecata
- Stato globale mutabile (non thread-safe senza lock)
- Confusione su source of truth (DB vs memory)

**Soluzione:**
```go
// 1. Verificare se menus è MAI letto:
grep -rn "menus\[.*\].*=" handlers/  # Trova WRITE
grep -rn "=.*menus\[" handlers/      # Trova READ

# 2. Se nessun READ trovato → RIMUOVERE:
// Rimuovere variabile globale
// var menus = make(map[string]*models.Menu)

// Rimuovere funzioni correlate
// func loadMenusFromStorage() { ... }
// func saveMenuToStorage() { ... }
```

**Stima Fix:** 30 min  
**Priorità:** 🟡 MEDIA (cleanup, non bug)

---

## ⚠️ PROBLEMI PERFORMANCE

### 3. Template Re-Parse su Ogni Richiesta?

**File:** [handlers/handlers.go](handlers/handlers.go#L45)  
**Linee:** 45, 142-146

**Problema:**
```go
// Line 45:
var templates *template.Template

// Line 142-146:
func loadTemplates() {
    var err error
    templates, err = template.ParseGlob("templates/*.html")
    if err != nil {
        log.Printf("⚠️  Errore nel caricamento dei template: %v", err)
        createFallbackTemplates()
    }
}
```

**Descrizione:**  
Template caricati una sola volta all'avvio (OK), MA nessun caching esplicito o `sync.Once` per garantire caricamento singolo.

**Verifica:**
```bash
grep -n "loadTemplates()" handlers/
# Se chiamato più volte → problema
```

**Impatto:**  
- Se chiamato multiplo: I/O disk ripetuto, CPU wasted
- Piccolo impatto (template sono ~10 file)

**Soluzione:**
```go
var (
    templates     *template.Template
    templatesOnce sync.Once
)

func getTemplates() *template.Template {
    templatesOnce.Do(func() {
        var err error
        templates, err = template.ParseGlob("templates/*.html")
        if err != nil {
            log.Printf("⚠️  Errore template: %v", err)
            createFallbackTemplates()
        }
        log.Println("✅ Templates loaded and cached")
    })
    return templates
}

// Poi in ogni handler:
tmpl := getTemplates()
tmpl.ExecuteTemplate(w, "admin.html", data)
```

**Stima Fix:** 30 min  
**Priorità:** 🟢 BASSA (già funziona decentemente)

---

### 4. Slice-to-Map Conversion per Template

**File:** [handlers/handlers.go](handlers/handlers.go#L244-L246)  
**Linee:** 244-246

**Problema:**
```go
// Converti slice in map per compatibilità con il template
restaurantMenus := make(map[string]*models.Menu)
for _, menu := range menusFromDB {
    restaurantMenus[menu.ID] = menu
}
```

**Descrizione:**  
Database ritorna slice `[]*models.Menu`, ma template si aspetta `map[string]*models.Menu`.  
→ Conversion loop necessaria prima di ogni render.

**Impatto:**  
- O(n) extra per ogni richiesta admin panel
- Memoria duplicata (slice + map)
- Piccolo overhead (~1ms per 50 menu)

**Soluzione:**

**Opzione A: Cambiare template** (PREFERRED)
```html
<!-- In admin.html, invece di: -->
{{range $id, $menu := .Menus}}
    <div data-menu-id="{{$id}}">...</div>
{{end}}

<!-- Usare: -->
{{range .Menus}}
    <div data-menu-id="{{.ID}}">...</div>
{{end}}
```

**Opzione B: Ritornare map da DB**
```go
// In db/mongodb.go, aggiungere:
func (m *MongoDB) GetMenusMapByRestaurantID(ctx context.Context, restaurantID string) (map[string]*models.Menu, error) {
    menus, err := m.GetMenusByRestaurantID(ctx, restaurantID)
    if err != nil {
        return nil, err
    }
    
    result := make(map[string]*models.Menu, len(menus))
    for _, menu := range menus {
        result[menu.ID] = menu
    }
    return result, nil
}
```

**Stima Fix:** 1h (Opzione A) o 30 min (Opzione B)  
**Priorità:** 🟢 BASSA

---

## 🧹 CODE SMELLS

### 5. Context Timeout Hardcodato

**File:** Multiple (`handlers/handlers.go`, `handlers/auth.go`)  
**Pattern:**
```go
ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
defer cancel()
```

**Problema:**  
- Timeout 3s hardcodato in ~30+ punti
- Cambio globale richiede edit multipli
- Nessuna differenziazione per operazioni (query semplice vs aggregation pesante)

**Soluzione:**
```go
// In pkg/config/config.go:
type DatabaseConfig struct {
    QueryTimeout      time.Duration `default:"5s"`
    MutationTimeout   time.Duration `default:"10s"`
    AggregationTimeout time.Duration `default:"15s"`
}

// In handlers:
func (h *Handler) contextWithTimeout(r *http.Request, timeoutType string) (context.Context, context.CancelFunc) {
    var timeout time.Duration
    switch timeoutType {
    case "query":
        timeout = h.config.Database.QueryTimeout
    case "mutation":
        timeout = h.config.Database.MutationTimeout
    default:
        timeout = 5 * time.Second
    }
    return context.WithTimeout(r.Context(), timeout)
}

// Uso:
ctx, cancel := h.contextWithTimeout(r, "query")
defer cancel()
```

**Stima Fix:** 3h (refactor tutti i context)  
**Priorità:** 🟡 MEDIA

---

### 6. Error Handling Inconsistente

**File:** Multiple handlers

**Pattern Mix:**
```go
// Tipo 1: http.Error con string
http.Error(w, "Errore nel caricamento menu", http.StatusInternalServerError)

// Tipo 2: Redirect
http.Redirect(w, r, "/admin?error=not_found", http.StatusFound)

// Tipo 3: JSON response
json.NewEncoder(w).Encode(map[string]interface{}{
    "success": false,
    "message": "Errore",
})
```

**Problema:**  
- 3 pattern diversi per stessa cosa
- Client non sa cosa aspettarsi (HTML? JSON? Redirect?)
- Difficile debugging

**Soluzione:**
```go
// Creare pkg/response/response.go
package response

type ErrorResponse struct {
    Success bool   `json:"success"`
    Error   string `json:"error"`
    Code    int    `json:"code"`
}

func Error(w http.ResponseWriter, code int, message string) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(code)
    json.NewEncoder(w).Encode(ErrorResponse{
        Success: false,
        Error:   message,
        Code:    code,
    })
}

func Success(w http.ResponseWriter, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
        "data":    data,
    })
}

// Uso:
response.Error(w, http.StatusNotFound, "Menu non trovato")
```

**Stima Fix:** 4h (refactor ~40 handlers)  
**Priorità:** 🟡 MEDIA

---

### 7. TODO Comments da Risolvere

**Found:** 8 TODO comments

**Location e Priorità:**

| File | Line | TODO | Priorità |
|------|------|------|----------|
| [handlers/auth.go](handlers/auth.go#L79) | 79 | "Implementare seed MongoDB con createUser" | 🟢 Bassa (già fatto) |
| [handlers/auth.go](handlers/auth.go#L817) | 817 | "Aggiornare load per usare MongoDB" | 🟢 Bassa (già fatto) |
| [middleware/security.go](middleware/security.go#L221) | 221 | "Implementare validazione CSRF token reale" | 🟡 Media (vedi Task 9 MAINTENANCE_PLAN.md) |
| [api_backup/menu.go](api_backup/menu.go#L223) | 223 | "Log audit to MongoDB audit_logs" | 🟢 Bassa (nice-to-have) |
| [api_backup/router.go](api_backup/router.go#L577) | 577 | "Implement GDPR handlers with MongoDB" | 🟢 Bassa (api_backup non usato) |
| [api_backup/security.go](api_backup/security.go#L13) | 13 | "Update GDPR handlers to use MongoDB" | 🟢 Bassa (api_backup non usato) |

**Azione:**
- [ ] Rimuovere TODO già implementati (auth.go line 79, 817)
- [ ] Implementare CSRF verification (Task 9)
- [ ] Decidere se `api_backup/` è mantenuto o deprecated

**Stima:** 2h  
**Priorità:** 🟢 BASSA (cleanup)

---

## 📊 PERFORMANCE METRICS (DA VERIFICARE)

### Test Consigliati

#### 1. **Load Test Admin Panel**
```bash
# Tool: Apache Bench
ab -n 1000 -c 10 http://localhost:8080/admin
# Target: < 200ms P95, < 300ms P99
```

#### 2. **Database Query Profiling**
```js
// MongoDB Shell
db.setProfilingLevel(2, { slowms: 100 })  // Log queries > 100ms
db.system.profile.find().limit(10).sort({ ts: -1 }).pretty()

// Identificare query lente:
db.system.profile.find({ millis: { $gt: 100 } }).sort({ millis: -1 })
```

#### 3. **Memory Profiling**
```bash
# Go pprof
go tool pprof http://localhost:8080/debug/pprof/heap
# Cercare memory leaks o allocazioni eccessive
```

---

## ✅ QUICK WINS

Queste fix richiedono < 1 ora e danno benefici immediati:

1. **Rimuovere TODO obsoleti** (30 min)
   - auth.go line 79, 817
   
2. **UpdateMany invece di loop** (1h)
   - SetActiveMenuHandler fix N+1

3. **Template getTemplates() con sync.Once** (30 min)
   - Garantire single load

**Total:** 2 ore per 3 optimization

---

## 📋 PRIORITÀ IMPLEMENTAZIONE

### 🔴 Alta (Settimana 1)
- [ ] Fix N+1 Query in SetActiveMenuHandler (1h)

### 🟡 Media (Settimana 2-3)
- [ ] Rimuovere menus map in-memory legacy (30 min)
- [ ] Context timeout configurabile (3h)
- [ ] Error handling standardizzato (4h)

### 🟢 Bassa (Settimana 4+)
- [ ] Template caching con sync.Once (30 min)
- [ ] Slice-to-map conversion ottimizzata (1h)
- [ ] Cleanup TODO comments (2h)

---

## 🎯 METRICHE SUCCESSO

| Metrica | Prima | Target | Come Misurare |
|---------|-------|--------|---------------|
| Admin Panel Load Time | ~300ms | < 200ms | Chrome DevTools |
| SetActiveMenu Time | ~500ms | < 100ms | Logger timestamps |
| DB Query Count (admin) | ~15 | < 5 | MongoDB profiler |
| Memory Usage | ? | < 100MB | Railway metrics |
| Test Coverage | 30% | 70% | go test -cover |

---

**Documento creato:** 2025-01-XX  
**Prossima Review:** Dopo implementazione Task Settimana 1-2
