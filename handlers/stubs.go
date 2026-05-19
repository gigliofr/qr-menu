package handlers

import (
    "context"
    "encoding/json"
    "net/http"
    "time"

    "github.com/gorilla/mux"
)

// HomeHandler pagina principale
func HomeHandler(w http.ResponseWriter, r *http.Request) {
    setSecurityHeaders(w)
    renderTemplate(w, "guest_access", nil)
}

// ShareMenuHandler mostra la pagina di condivisione per un menu
func ShareMenuHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    menuID := vars["id"]
    data := struct{ MenuID string }{MenuID: menuID}
    renderTemplate(w, "share_menu", data)
}

// HealthHandler healthcheck
func HealthHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// Add/Edit/Delete/Upload item handlers - simple stubs that redirect back
func AddItemHandler(w http.ResponseWriter, r *http.Request) {
    if !requireValidCSRF(w, r) {
        return
    }
    http.Redirect(w, r, "/admin", http.StatusSeeOther)
}
func EditItemHandler(w http.ResponseWriter, r *http.Request) {
    if !requireValidCSRF(w, r) {
        return
    }
    http.Redirect(w, r, "/admin", http.StatusSeeOther)
}
func DeleteItemHandler(w http.ResponseWriter, r *http.Request) {
    if !requireValidCSRF(w, r) {
        return
    }
    http.Redirect(w, r, "/admin", http.StatusSeeOther)
}
func UploadItemImageHandler(w http.ResponseWriter, r *http.Request) {
    if !requireValidCSRF(w, r) {
        return
    }
    // In real implementation handle file upload; here stub
    http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

// helper: context with timeout
func contextWithTimeout(r *http.Request) (context.Context, context.CancelFunc) {
    return context.WithTimeout(r.Context(), 5*time.Second)
}

