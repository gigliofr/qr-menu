package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"qr-menu/localization"
)

// LocaleResponse è la risposta delle richieste locale
type LocaleResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// GetTranslationsHandler recupera tutte le traduzioni per una locale
func GetTranslationsHandler(w http.ResponseWriter, r *http.Request) {
	// Leggi la locale da query params (default: accetta header Accept-Language)
	locale := r.URL.Query().Get("locale")
	if locale == "" {
		locale = "it" // Default locale
	}

	lm := localization.GetLocalizationManager()
	translations := lm.GetAll(locale)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(LocaleResponse{
		Status: "success",
		Data: map[string]interface{}{
			"locale":        locale,
			"translations": translations,
		},
	})
}

// GetSupportedLocalesHandler recupera la lista delle locale supportate
func GetSupportedLocalesHandler(w http.ResponseWriter, r *http.Request) {
	lm := localization.GetLocalizationManager()
	locales := lm.GetSupportedLocales()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(LocaleResponse{
		Status: "success",
		Data: map[string]interface{}{
			"locales": locales,
			"count":   len(locales),
		},
	})
}

// SetUserLocaleHandler imposta la locale preferita dell'utente
func SetUserLocaleHandler(w http.ResponseWriter, r *http.Request) {
	// Verifica autenticazione
	session, err := getSessionFromRequest(r)
	if err != nil || session.RestaurantID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		Locale string `json:"locale"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.Locale == "" {
		http.Error(w, "Locale richiesta", http.StatusBadRequest)
		return
	}

	lm := localization.GetLocalizationManager()
	err = lm.SetUserLocale(session.RestaurantID, req.Locale)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(LocaleResponse{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(LocaleResponse{
		Status:  "success",
		Message: "Locale impostata",
		Data: map[string]interface{}{
			"locale": req.Locale,
		},
	})
}

// GetUserLocaleHandler recupera la locale preferita dell'utente
func GetUserLocaleHandler(w http.ResponseWriter, r *http.Request) {
	// Verifica autenticazione
	session, err := getSessionFromRequest(r)
	if err != nil || session.RestaurantID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	lm := localization.GetLocalizationManager()
	userLocale := lm.GetUserLocale(session.RestaurantID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(LocaleResponse{
		Status: "success",
		Data: map[string]interface{}{
			"locale": userLocale,
		},
	})
}

// GetTranslationHandler recupera una singola traduzione
func GetTranslationHandler(w http.ResponseWriter, r *http.Request) {
	// Verifica autenticazione
	session, err := getSessionFromRequest(r)
	if err != nil || session.RestaurantID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Leggi i parametri
	key := r.URL.Query().Get("key")
	locale := r.URL.Query().Get("locale")

	if key == "" {
		http.Error(w, "Key è richiesta", http.StatusBadRequest)
		return
	}

	if locale == "" {
		lm := localization.GetLocalizationManager()
		locale = lm.GetUserLocale(session.RestaurantID)
	}

	lm := localization.GetLocalizationManager()
	translation := lm.Get(locale, key)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(LocaleResponse{
		Status: "success",
		Data: map[string]interface{}{
			"key":         key,
			"locale":      locale,
			"translation": translation,
		},
	})
}

// FormatCurrencyHandler formatta un numero come valuta
func FormatCurrencyHandler(w http.ResponseWriter, r *http.Request) {
	// Verifica autenticazione
	session, err := getSessionFromRequest(r)
	if err != nil || session.RestaurantID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Leggi i parametri
	locale := r.URL.Query().Get("locale")
	amountStr := r.URL.Query().Get("amount")

	if amountStr == "" {
		http.Error(w, "Amount è richiesta", http.StatusBadRequest)
		return
	}

	if locale == "" {
		lm := localization.GetLocalizationManager()
		locale = lm.GetUserLocale(session.RestaurantID)
	}

	// Converte l'amount a float
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		http.Error(w, "Invalid amount", http.StatusBadRequest)
		return
	}

	lm := localization.GetLocalizationManager()
	formatted := lm.FormatCurrency(locale, amount)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(LocaleResponse{
		Status: "success",
		Data: map[string]interface{}{
			"amount":    amount,
			"locale":    locale,
			"formatted": formatted,
		},
	})
}
