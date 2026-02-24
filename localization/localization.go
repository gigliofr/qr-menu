package localization

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"qr-menu/logger"
)

// LocalizationManager gestisce la localizzazione multi-lingua
type LocalizationManager struct {
	mu                sync.RWMutex
	translations      map[string]map[string]interface{} // locale -> key -> value
	defaultLocale     string
	supportedLocales  []string
	currencySymbols   map[string]string
	dateFormats       map[string]string
	timeFormats       map[string]string
	translationPath   string
	userLocalePrefs   map[string]string // restaurantID -> locale
}

// Translation contiene una singola traduzione
type Translation struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

// LocaleInfo contiene informazioni su una locale
type LocaleInfo struct {
	Code       string `json:"code"`
	Name       string `json:"name"`
	NativeName string `json:"native_name"`
	Flag       string `json:"flag"`
	Direction  string `json:"direction"` // ltr o rtl
}

// TranslationResponse è la risposta delle traduzioni
type TranslationResponse struct {
	Locale       string                 `json:"locale"`
	Translations map[string]interface{} `json:"translations"`
}

var (
	defaultManager *LocalizationManager
	once           sync.Once

	// Supported locales
	supportedLocalesMap = map[string]LocaleInfo{
		"it": {
			Code:       "it",
			Name:       "Italian",
			NativeName: "Italiano",
			Flag:       "🇮🇹",
			Direction:  "ltr",
		},
		"en": {
			Code:       "en",
			Name:       "English",
			NativeName: "English",
			Flag:       "🇬🇧",
			Direction:  "ltr",
		},
		"fr": {
			Code:       "fr",
			Name:       "French",
			NativeName: "Français",
			Flag:       "🇫🇷",
			Direction:  "ltr",
		},
		"es": {
			Code:       "es",
			Name:       "Spanish",
			NativeName: "Español",
			Flag:       "🇪🇸",
			Direction:  "ltr",
		},
		"de": {
			Code:       "de",
			Name:       "German",
			NativeName: "Deutsch",
			Flag:       "🇩🇪",
			Direction:  "ltr",
		},
		"pt": {
			Code:       "pt",
			Name:       "Portuguese",
			NativeName: "Português",
			Flag:       "🇵🇹",
			Direction:  "ltr",
		},
		"ja": {
			Code:       "ja",
			Name:       "Japanese",
			NativeName: "日本語",
			Flag:       "🇯🇵",
			Direction:  "ltr",
		},
		"zh": {
			Code:       "zh",
			Name:       "Chinese",
			NativeName: "中文",
			Flag:       "🇨🇳",
			Direction:  "ltr",
		},
		"ar": {
			Code:       "ar",
			Name:       "Arabic",
			NativeName: "العربية",
			Flag:       "🇸🇦",
			Direction:  "rtl",
		},
	}

	// Currency symbols per locale
	currencySymbols = map[string]string{
		"it": "€",
		"en": "£",
		"fr": "€",
		"es": "€",
		"de": "€",
		"pt": "€",
		"ja": "¥",
		"zh": "¥",
		"ar": "﷼",
	}

	// Date formats per locale (Go format)
	dateFormats = map[string]string{
		"it": "02/01/2006",
		"en": "01/02/2006",
		"fr": "02/01/2006",
		"es": "02/01/2006",
		"de": "02.01.2006",
		"pt": "02/01/2006",
		"ja": "2006-01-02",
		"zh": "2006-01-02",
		"ar": "02/01/2006",
	}

	// Time formats per locale
	timeFormats = map[string]string{
		"it": "15:04",
		"en": "03:04 PM",
		"fr": "15:04",
		"es": "15:04",
		"de": "15:04",
		"pt": "15:04",
		"ja": "15:04",
		"zh": "15:04",
		"ar": "15:04",
	}
)

// GetLocalizationManager restituisce il singleton LocalizationManager
func GetLocalizationManager() *LocalizationManager {
	once.Do(func() {
		defaultManager = &LocalizationManager{
			defaultLocale:    "it",
			supportedLocales: []string{"it", "en", "fr", "es", "de", "pt", "ja", "zh", "ar"},
			currencySymbols:  currencySymbols,
			dateFormats:      dateFormats,
			timeFormats:      timeFormats,
			translations:     make(map[string]map[string]interface{}),
			userLocalePrefs:  make(map[string]string),
		}
	})
	return defaultManager
}

// Init inizializza il localization manager
func (lm *LocalizationManager) Init(translationPath string) error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	lm.translationPath = translationPath

	// Carica i file di traduzione
	for _, locale := range lm.supportedLocales {
		filePath := filepath.Join(translationPath, locale+".json")
		err := lm.loadTranslationFile(locale, filePath)
		if err != nil {
			logger.Warn("Errore caricamento traduzione", map[string]interface{}{
				"locale": locale,
				"error":  err.Error(),
			})
			// Continua con altre lingue anche se una fallisce
		}
	}

	logger.Info("Localization manager inizializzato", map[string]interface{}{
		"default_locale":     lm.defaultLocale,
		"supported_locales": strings.Join(lm.supportedLocales, ", "),
		"translation_path":   translationPath,
	})

	return nil
}

// loadTranslationFile carica un file di traduzione JSON
func (lm *LocalizationManager) loadTranslationFile(locale string, filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("errore lettura file traduzione: %w", err)
	}

	var translations map[string]interface{}
	err = json.Unmarshal(data, &translations)
	if err != nil {
		return fmt.Errorf("errore parse JSON: %w", err)
	}

	lm.translations[locale] = translations
	return nil
}

// Get recupera una traduzione
func (lm *LocalizationManager) Get(locale string, key string) string {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	// Usa la locale dell'utente se disponibile
	currentLocale := locale
	if currentLocale == "" {
		currentLocale = lm.defaultLocale
	}

	// Verifica se la locale è supportata
	if _, exists := lm.translations[currentLocale]; !exists {
		currentLocale = lm.defaultLocale
	}

	// Ottieni la traduzione
	if trans, ok := lm.translations[currentLocale]; ok {
		if value, ok := trans[key]; ok {
			return fmt.Sprintf("%v", value)
		}
	}

	// Fallback alla lingua di default
	if currentLocale != lm.defaultLocale {
		if trans, ok := lm.translations[lm.defaultLocale]; ok {
			if value, ok := trans[key]; ok {
				return fmt.Sprintf("%v", value)
			}
		}
	}

	// Se nulla trovato, ritorna la chiave stessa
	return key
}

// GetAll recupera tutte le traduzioni per una locale
func (lm *LocalizationManager) GetAll(locale string) map[string]interface{} {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	// Usa la locale dell'utente se disponibile
	currentLocale := locale
	if currentLocale == "" {
		currentLocale = lm.defaultLocale
	}

	// Verifica se la locale è supportata
	if translations, exists := lm.translations[currentLocale]; exists {
		return translations
	}

	// Fallback alla lingua di default
	if translations, exists := lm.translations[lm.defaultLocale]; exists {
		return translations
	}

	return make(map[string]interface{})
}

// GetWithParams recupera una traduzione con parametri
func (lm *LocalizationManager) GetWithParams(locale string, key string, params map[string]string) string {
	translation := lm.Get(locale, key)

	// Sostituisci i parametri {{param}} con i valori
	for paramKey, paramValue := range params {
		translation = strings.ReplaceAll(translation, "{{"+paramKey+"}}", paramValue)
	}

	return translation
}

// GetSupportedLocales restituisce la lista di locale supportate
func (lm *LocalizationManager) GetSupportedLocales() []LocaleInfo {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	var locales []LocaleInfo
	for _, code := range lm.supportedLocales {
		if info, ok := supportedLocalesMap[code]; ok {
			locales = append(locales, info)
		}
	}

	return locales
}

// SetUserLocale imposta la locale preferita per un utente
func (lm *LocalizationManager) SetUserLocale(restaurantID string, locale string) error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	// Verifica che la locale sia supportata
	isSupported := false
	for _, supported := range lm.supportedLocales {
		if supported == locale {
			isSupported = true
			break
		}
	}

	if !isSupported {
		return fmt.Errorf("locale non supportata: %s", locale)
	}

	lm.userLocalePrefs[restaurantID] = locale
	logger.Info("Locale utente impostata", map[string]interface{}{
		"restaurant_id": restaurantID,
		"locale":        locale,
	})

	return nil
}

// GetUserLocale recupera la locale preferita dell'utente
func (lm *LocalizationManager) GetUserLocale(restaurantID string) string {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	if locale, exists := lm.userLocalePrefs[restaurantID]; exists {
		return locale
	}

	return lm.defaultLocale
}

// FormatCurrency formatta un numero come valuta
func (lm *LocalizationManager) FormatCurrency(locale string, amount float64) string {
	symbol := lm.getCurrencySymbol(locale)
	return fmt.Sprintf("%s %.2f", symbol, amount)
}

// FormatDate formatta una data secondo la locale
func (lm *LocalizationManager) FormatDate(locale string, t time.Time) string {
	format := lm.getDateFormat(locale)
	return t.Format(format)
}

// FormatTime formatta un orario secondo la locale
func (lm *LocalizationManager) FormatTime(locale string, t time.Time) string {
	format := lm.getTimeFormat(locale)
	return t.Format(format)
}

// FormatDateTime formatta una data e ora secondo la locale
func (lm *LocalizationManager) FormatDateTime(locale string, t time.Time) string {
	dateFormat := lm.getDateFormat(locale)
	timeFormat := lm.getTimeFormat(locale)
	date := t.Format(dateFormat)
	time := t.Format(timeFormat)
	return fmt.Sprintf("%s %s", date, time)
}

// getCurrencySymbol recupera il simbolo della valuta
func (lm *LocalizationManager) getCurrencySymbol(locale string) string {
	if symbol, ok := lm.currencySymbols[locale]; ok {
		return symbol
	}
	return lm.currencySymbols[lm.defaultLocale]
}

// getDateFormat recupera il formato della data
func (lm *LocalizationManager) getDateFormat(locale string) string {
	if format, ok := lm.dateFormats[locale]; ok {
		return format
	}
	return lm.dateFormats[lm.defaultLocale]
}

// getTimeFormat recupera il formato dell'orario
func (lm *LocalizationManager) getTimeFormat(locale string) string {
	if format, ok := lm.timeFormats[locale]; ok {
		return format
	}
	return lm.timeFormats[lm.defaultLocale]
}

// CreateDefaultTranslationFiles crea i file di traduzione di default
func (lm *LocalizationManager) CreateDefaultTranslationFiles(path string) error {
	// Crea la directory se non esiste
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("errore creazione directory: %w", err)
	}

	defaultTranslations := map[string]map[string]interface{}{
		"it": {
			"app_name":                    "QR Menu",
			"welcome":                     "Benvenuto",
			"login":                       "Accedi",
			"logout":                      "Esci",
			"register":                    "Registrati",
			"menu":                        "Menu",
			"order":                       "Ordine",
			"orders":                      "Ordini",
			"reservation":                 "Prenotazione",
			"reservations":                "Prenotazioni",
			"profile":                     "Profilo",
			"settings":                    "Impostazioni",
			"help":                        "Aiuto",
			"about":                       "Chi Siamo",
			"error":                       "Errore",
			"success":                     "Successo",
			"loading":                     "Caricamento...",
			"no_data":                     "Nessun dato disponibile",
			"order_created":               "Ordine creato con successo",
			"order_ready":                 "Ordine pronto",
			"reservation_confirmed":       "Prenotazione confermata",
			"order_status_pending":        "In attesa",
			"order_status_confirmed":      "Confermato",
			"order_status_preparing":      "In preparazione",
			"order_status_ready":          "Pronto",
			"order_status_completed":      "Completato",
			"order_status_cancelled":      "Annullato",
		},
		"en": {
			"app_name":                    "QR Menu",
			"welcome":                     "Welcome",
			"login":                       "Login",
			"logout":                      "Logout",
			"register":                    "Register",
			"menu":                        "Menu",
			"order":                       "Order",
			"orders":                      "Orders",
			"reservation":                 "Reservation",
			"reservations":                "Reservations",
			"profile":                     "Profile",
			"settings":                    "Settings",
			"help":                        "Help",
			"about":                       "About",
			"error":                       "Error",
			"success":                     "Success",
			"loading":                     "Loading...",
			"no_data":                     "No data available",
			"order_created":               "Order created successfully",
			"order_ready":                 "Order ready",
			"reservation_confirmed":       "Reservation confirmed",
			"order_status_pending":        "Pending",
			"order_status_confirmed":      "Confirmed",
			"order_status_preparing":      "Preparing",
			"order_status_ready":          "Ready",
			"order_status_completed":      "Completed",
			"order_status_cancelled":      "Cancelled",
		},
		"fr": {
			"app_name":                    "QR Menu",
			"welcome":                     "Bienvenue",
			"login":                       "Connexion",
			"logout":                      "Déconnexion",
			"register":                    "S'inscrire",
			"menu":                        "Menu",
			"order":                       "Commande",
			"orders":                      "Commandes",
			"reservation":                 "Réservation",
			"reservations":                "Réservations",
			"profile":                     "Profil",
			"settings":                    "Paramètres",
			"help":                        "Aide",
			"about":                       "À propos",
			"error":                       "Erreur",
			"success":                     "Succès",
			"loading":                     "Chargement...",
			"no_data":                     "Aucune donnée disponible",
			"order_created":               "Commande créée avec succès",
			"order_ready":                 "Commande prête",
			"reservation_confirmed":       "Réservation confirmée",
			"order_status_pending":        "En attente",
			"order_status_confirmed":      "Confirmé",
			"order_status_preparing":      "Préparation",
			"order_status_ready":          "Prêt",
			"order_status_completed":      "Complété",
			"order_status_cancelled":      "Annulé",
		},
	}

	// Crea i file di traduzione
	for locale, translations := range defaultTranslations {
		filePath := filepath.Join(path, locale+".json")

		// Controlla se il file esiste già
		if _, err := os.Stat(filePath); err == nil {
			continue // File esiste, salta
		}

		data, err := json.MarshalIndent(translations, "", "  ")
		if err != nil {
			logger.Error("Errore marshalling traduzioni", map[string]interface{}{
				"locale": locale,
				"error":  err.Error(),
			})
			continue
		}

		err = os.WriteFile(filePath, data, 0644)
		if err != nil {
			logger.Error("Errore scrittura file traduzione", map[string]interface{}{
				"locale": locale,
				"path":   filePath,
				"error":  err.Error(),
			})
			continue
		}

		logger.Info("File traduzione creato", map[string]interface{}{
			"locale": locale,
			"path":   filePath,
		})
	}

	return nil
}

// ReloadTranslations ricarica tutte le traduzioni
func (lm *LocalizationManager) ReloadTranslations() error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	lm.translations = make(map[string]map[string]interface{})

	for _, locale := range lm.supportedLocales {
		filePath := filepath.Join(lm.translationPath, locale+".json")
		err := lm.loadTranslationFile(locale, filePath)
		if err != nil {
			logger.Warn("Errore ricaricamento traduzione", map[string]interface{}{
				"locale": locale,
				"error":  err.Error(),
			})
		}
	}

	logger.Info("Traduzioni ricaricate", nil)
	return nil
}
