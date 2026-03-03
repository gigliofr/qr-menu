package handlers

import (
	"context"
	"log"
	"github.com/gigliofr/qr-menu/db"
	"time"
)

// RecordAuditLog registra un evento di audit nel database
// action: es "MENU_CREATED", "MENU_UPDATED", "MENU_DELETED"
// resourceType: es "menu", "restaurant", "item"
// resourceID: ID della risorsa modificata
// restaurantID: ID del ristorante proprietario
// clientIP: indirizzo IP del client
// userAgent: user agent del client
// status: "success", "failure", o "warning"
func RecordAuditLog(ctx context.Context, action, resourceType, resourceID, restaurantID, clientIP, userAgent, status string) {
	// Crea context con timeout per non bloccare la request
	auditCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Prepara il documento di audit usando la struttura AuditLog
	auditLog := &db.AuditLog{
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		RestaurantID: restaurantID,
		IPAddress:    clientIP,
		UserAgent:    userAgent,
		Status:       status,
		Timestamp:    time.Now(),
	}

	// Registra nel database
	if db.MongoInstance != nil {
		err := db.MongoInstance.CreateAuditLog(auditCtx, auditLog)
		if err != nil {
			// Non blocchiamo la response se l'audit fail, solo log
			log.Printf("⚠️  Errore registrazione audit log: %v", err)
		}
	}
}

// RecordAuditLogAsync registra un evento di audit in background senza bloccare la response
// Utile per operazioni non-critical
func RecordAuditLogAsync(action, resourceType, resourceID, restaurantID, clientIP, userAgent, status string) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Panic in audit logging: %v", r)
			}
		}()
		RecordAuditLog(context.Background(), action, resourceType, resourceID, restaurantID, clientIP, userAgent, status)
	}()
}
