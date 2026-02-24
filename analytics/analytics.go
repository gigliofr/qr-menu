package analytics

import (
	"encoding/json"
	"os"
	"path/filepath"
	"qr-menu/logger"
	"strings"
	"sync"
	"time"
)

// Analytics rappresenta il sistema di analisi
type Analytics struct {
	mu    sync.RWMutex
	stats map[string]*RestaurantStats
}

// RestaurantStats contiene le statistiche di un ristorante
type RestaurantStats struct {
	RestaurantID     string         `json:"restaurant_id"`
	TotalViews       int            `json:"total_views"`
	UniqueViews      int            `json:"unique_views"`
	DailyViews       map[string]int `json:"daily_views"`
	HourlyViews      map[int]int    `json:"hourly_views"`
	DeviceTypes      map[string]int `json:"device_types"`
	OperatingSystems map[string]int `json:"operating_systems"`
	Browsers         map[string]int `json:"browsers"`
	Countries        map[string]int `json:"countries"`
	MenuViews        map[string]int `json:"menu_views"`
	PopularItems     []PopularItem  `json:"popular_items"`
	ShareStats       ShareStats     `json:"share_stats"`
	QRCodeScans      map[string]int `json:"qr_code_scans"`
	LastUpdated      time.Time      `json:"last_updated"`
}

// PopularItem rappresenta un piatto popolare
type PopularItem struct {
	ItemID     string  `json:"item_id"`
	ItemName   string  `json:"item_name"`
	Views      int     `json:"views"`
	CategoryID string  `json:"category_id"`
	Price      float64 `json:"price"`
}

// ShareStats contiene statistiche di condivisione
type ShareStats struct {
	WhatsApp int `json:"whatsapp"`
	Telegram int `json:"telegram"`
	Facebook int `json:"facebook"`
	Twitter  int `json:"twitter"`
	CopyLink int `json:"copy_link"`
	Total    int `json:"total"`
}

// ViewEvent rappresenta un evento di visualizzazione
type ViewEvent struct {
	RestaurantID string    `json:"restaurant_id"`
	MenuID       string    `json:"menu_id"`
	ItemID       string    `json:"item_id,omitempty"`
	Timestamp    time.Time `json:"timestamp"`
	UserIP       string    `json:"user_ip"`
	UserAgent    string    `json:"user_agent"`
	DeviceType   string    `json:"device_type"`
	Browser      string    `json:"browser"`
	OS           string    `json:"os"`
	Country      string    `json:"country"`
	Referrer     string    `json:"referrer"`
	SessionID    string    `json:"session_id"`
}

// ShareEvent rappresenta un evento di condivisione
type ShareEvent struct {
	RestaurantID string    `json:"restaurant_id"`
	MenuID       string    `json:"menu_id"`
	Platform     string    `json:"platform"` // whatsapp, telegram, facebook, twitter, copy
	Timestamp    time.Time `json:"timestamp"`
	UserIP       string    `json:"user_ip"`
	UserAgent    string    `json:"user_agent"`
}

// QRScanEvent rappresenta una scansione QR
type QRScanEvent struct {
	RestaurantID string    `json:"restaurant_id"`
	MenuID       string    `json:"menu_id"`
	Timestamp    time.Time `json:"timestamp"`
	UserIP       string    `json:"user_ip"`
	UserAgent    string    `json:"user_agent"`
	Location     string    `json:"location,omitempty"`
}

var (
	globalAnalytics *Analytics
	once            sync.Once
)

// GetAnalytics restituisce l'istanza singleton di Analytics
func GetAnalytics() *Analytics {
	once.Do(func() {
		globalAnalytics = &Analytics{
			stats: make(map[string]*RestaurantStats),
		}
		globalAnalytics.loadFromStorage()
	})
	return globalAnalytics
}

// TrackView registra una visualizzazione pagina
func (a *Analytics) TrackView(event ViewEvent) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Inizializza stats se non esistono
	if a.stats[event.RestaurantID] == nil {
		a.stats[event.RestaurantID] = &RestaurantStats{
			RestaurantID:     event.RestaurantID,
			DailyViews:       make(map[string]int),
			HourlyViews:      make(map[int]int),
			DeviceTypes:      make(map[string]int),
			OperatingSystems: make(map[string]int),
			Browsers:         make(map[string]int),
			Countries:        make(map[string]int),
			MenuViews:        make(map[string]int),
			QRCodeScans:      make(map[string]int),
		}
	}

	stats := a.stats[event.RestaurantID]

	// Aggiorna contatori
	stats.TotalViews++

	// Vista giornaliera
	dayKey := event.Timestamp.Format("2006-01-02")
	stats.DailyViews[dayKey]++

	// Vista oraria
	hour := event.Timestamp.Hour()
	stats.HourlyViews[hour]++

	// Device info
	stats.DeviceTypes[event.DeviceType]++
	stats.OperatingSystems[event.OS]++
	stats.Browsers[event.Browser]++
	stats.Countries[event.Country]++

	// Menu views
	if event.MenuID != "" {
		stats.MenuViews[event.MenuID]++
	}

	stats.LastUpdated = time.Now()

	// Log evento
	logger.Info("Analytics: View tracked", map[string]interface{}{
		"restaurant_id": event.RestaurantID,
		"menu_id":       event.MenuID,
		"device_type":   event.DeviceType,
		"country":       event.Country,
	})

	// Salva in background
	go a.saveToStorage()
}

// TrackShare registra una condivisione
func (a *Analytics) TrackShare(event ShareEvent) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.stats[event.RestaurantID] == nil {
		a.stats[event.RestaurantID] = &RestaurantStats{
			RestaurantID: event.RestaurantID,
			DailyViews:   make(map[string]int),
			HourlyViews:  make(map[int]int),
		}
	}

	stats := a.stats[event.RestaurantID]

	// Aggiorna statistiche condivisione
	switch event.Platform {
	case "whatsapp":
		stats.ShareStats.WhatsApp++
	case "telegram":
		stats.ShareStats.Telegram++
	case "facebook":
		stats.ShareStats.Facebook++
	case "twitter":
		stats.ShareStats.Twitter++
	case "copy":
		stats.ShareStats.CopyLink++
	}
	stats.ShareStats.Total++
	stats.LastUpdated = time.Now()

	logger.AuditLog("SHARE_TRACKED", "analytics",
		"Condivisione tracciata", event.RestaurantID, event.UserIP, event.UserAgent,
		map[string]interface{}{
			"platform": event.Platform,
			"menu_id":  event.MenuID,
		})

	go a.saveToStorage()
}

// TrackQRScan registra una scansione QR
func (a *Analytics) TrackQRScan(event QRScanEvent) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.stats[event.RestaurantID] == nil {
		a.stats[event.RestaurantID] = &RestaurantStats{
			RestaurantID: event.RestaurantID,
			DailyViews:   make(map[string]int),
			HourlyViews:  make(map[int]int),
			QRCodeScans:  make(map[string]int),
		}
	}

	stats := a.stats[event.RestaurantID]

	// Incrementa scansioni QR
	dayKey := event.Timestamp.Format("2006-01-02")
	stats.QRCodeScans[dayKey]++
	stats.LastUpdated = time.Now()

	logger.AuditLog("QR_SCAN_TRACKED", "analytics",
		"Scansione QR tracciata", event.RestaurantID, event.UserIP, event.UserAgent,
		map[string]interface{}{
			"menu_id":  event.MenuID,
			"location": event.Location,
		})

	go a.saveToStorage()
}

// GetRestaurantStats restituisce le statistiche di un ristorante
func (a *Analytics) GetRestaurantStats(restaurantID string) *RestaurantStats {
	a.mu.RLock()
	defer a.mu.RUnlock()

	stats, exists := a.stats[restaurantID]
	if !exists {
		return &RestaurantStats{
			RestaurantID: restaurantID,
			DailyViews:   make(map[string]int),
			HourlyViews:  make(map[int]int),
		}
	}

	// Restituisci copia per evitare race conditions
	statsCopy := *stats
	return &statsCopy
}

// GetDashboardData calcola dati aggregati per dashboard
func (a *Analytics) GetDashboardData(restaurantID string, days int) map[string]interface{} {
	a.mu.RLock()
	defer a.mu.RUnlock()

	stats := a.stats[restaurantID]
	if stats == nil {
		return map[string]interface{}{
			"total_views":   0,
			"unique_views":  0,
			"total_shares":  0,
			"qr_scans":      0,
			"daily_trend":   []interface{}{},
			"device_stats":  map[string]int{},
			"popular_items": []interface{}{},
		}
	}

	// Calcola trend degli ultimi N giorni
	dailyTrend := make([]map[string]interface{}, 0, days)
	now := time.Now()

	for i := days - 1; i >= 0; i-- {
		date := now.AddDate(0, 0, -i)
		dayKey := date.Format("2006-01-02")
		views := stats.DailyViews[dayKey]
		qrScans := stats.QRCodeScans[dayKey]

		dailyTrend = append(dailyTrend, map[string]interface{}{
			"date":     dayKey,
			"views":    views,
			"qr_scans": qrScans,
		})
	}

	// Calcola totale scansioni QR
	totalQRScans := 0
	for _, scans := range stats.QRCodeScans {
		totalQRScans += scans
	}

	return map[string]interface{}{
		"total_views":     stats.TotalViews,
		"unique_views":    stats.UniqueViews,
		"total_shares":    stats.ShareStats.Total,
		"qr_scans":        totalQRScans,
		"daily_trend":     dailyTrend,
		"device_stats":    stats.DeviceTypes,
		"os_stats":        stats.OperatingSystems,
		"browser_stats":   stats.Browsers,
		"country_stats":   stats.Countries,
		"popular_items":   stats.PopularItems,
		"share_breakdown": stats.ShareStats,
		"last_updated":    stats.LastUpdated,
	}
}

// Storage functions

func (a *Analytics) saveToStorage() {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// Crea directory se non esiste
	if err := os.MkdirAll("storage/analytics", 0755); err != nil {
		logger.Error("Errore creazione directory analytics", map[string]interface{}{"error": err.Error()})
		return
	}

	// Salva ogni ristorante in file separato
	for restaurantID, stats := range a.stats {
		filename := filepath.Join("storage/analytics", restaurantID+".json")

		data, err := json.MarshalIndent(stats, "", "  ")
		if err != nil {
			logger.Error("Errore serializzazione analytics", map[string]interface{}{
				"restaurant_id": restaurantID,
				"error":         err.Error(),
			})
			continue
		}

		if err := os.WriteFile(filename, data, 0644); err != nil {
			logger.Error("Errore salvataggio analytics", map[string]interface{}{
				"restaurant_id": restaurantID,
				"file":          filename,
				"error":         err.Error(),
			})
		}
	}
}

func (a *Analytics) loadFromStorage() {
	analyticsDir := "storage/analytics"

	// Controlla se directory esiste
	if _, err := os.Stat(analyticsDir); os.IsNotExist(err) {
		return
	}

	entries, err := os.ReadDir(analyticsDir)
	if err != nil {
		logger.Error("Errore lettura directory analytics", map[string]interface{}{"error": err.Error()})
		return
	}

	loadedCount := 0
	for _, entry := range entries {
		if entry.IsDir() || !validateFileExtension(entry.Name(), ".json") {
			continue
		}

		filename := filepath.Join(analyticsDir, entry.Name())
		data, err := os.ReadFile(filename)
		if err != nil {
			logger.Error("Errore lettura file analytics", map[string]interface{}{
				"file":  filename,
				"error": err.Error(),
			})
			continue
		}

		var stats RestaurantStats
		if err := json.Unmarshal(data, &stats); err != nil {
			logger.Error("Errore parsing analytics", map[string]interface{}{
				"file":  filename,
				"error": err.Error(),
			})
			continue
		}

		a.stats[stats.RestaurantID] = &stats
		loadedCount++
	}

	logger.Info("Analytics caricate", map[string]interface{}{"restaurants": loadedCount})
}

// Utility functions

func validateFileExtension(filename, ext string) bool {
	return filepath.Ext(filename) == ext
}

// ParseUserAgent estrae informazioni da User-Agent
func ParseUserAgent(userAgent string) (deviceType, browser, os string) {
	// Implementazione semplificata - in produzione usare libreria dedicata
	ua := strings.ToLower(userAgent)

	// Detect device type
	if strings.Contains(ua, "mobile") || strings.Contains(ua, "android") || strings.Contains(ua, "iphone") {
		deviceType = "mobile"
	} else if strings.Contains(ua, "tablet") || strings.Contains(ua, "ipad") {
		deviceType = "tablet"
	} else {
		deviceType = "desktop"
	}

	// Detect browser
	if strings.Contains(ua, "chrome") {
		browser = "chrome"
	} else if strings.Contains(ua, "firefox") {
		browser = "firefox"
	} else if strings.Contains(ua, "safari") {
		browser = "safari"
	} else if strings.Contains(ua, "edge") {
		browser = "edge"
	} else {
		browser = "other"
	}

	// Detect OS
	if strings.Contains(ua, "windows") {
		os = "windows"
	} else if strings.Contains(ua, "mac") || strings.Contains(ua, "darwin") {
		os = "macos"
	} else if strings.Contains(ua, "linux") {
		os = "linux"
	} else if strings.Contains(ua, "android") {
		os = "android"
	} else if strings.Contains(ua, "ios") || strings.Contains(ua, "iphone") || strings.Contains(ua, "ipad") {
		os = "ios"
	} else {
		os = "other"
	}

	return
}

// GetCountryFromIP ottiene il paese dall'IP (implementazione mock)
func GetCountryFromIP(ip string) string {
	// In produzione, usare servizio di geolocalizzazione come MaxMind
	// Per ora restituiamo "IT" come default
	return "IT"
}
