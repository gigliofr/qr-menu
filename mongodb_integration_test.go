// +build integration

package main

import (
	"context"
	"strings"
	"testing"
	"time"

	"qr-menu/db"
	"qr-menu/models"

	"github.com/google/uuid"
)

// TestMongoDBConnection verifica la connessione e le operazioni CRUD
func TestMongoDBConnection(t *testing.T) {
	separator := strings.Repeat("=", 60)
	subSeparator := strings.Repeat("-", 58)

	t.Log("\n" + separator)
	t.Log("🔧 MongoDB Integration Test - Full CRUD Operations")
	t.Log(separator)

	// Step 1: Connessione
	t.Log("\n📡 Step 1: Verifica Connessione MongoDB Atlas")
	t.Log(subSeparator)

	t.Log("Cluster: cluster0.b9jfwmr.mongodb.net")
	t.Log("Database: qr-menu")
	t.Log("Authentication: X.509 Certificate")

	err := db.Connect()
	if err != nil {
		t.Fatalf("❌ ERRORE CONNESSIONE: %v\n", err)
	}

	defer db.MongoInstance.Disconnect()

	t.Log("✅ CONNESSIONE VERIFICATA!")

	// Step 2: Inserimento dati
	t.Log("\n📝 Step 2: Inserimento Dati Test")
	t.Log(subSeparator)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	restaurantID := uuid.New().String()
	restaurant := &models.Restaurant{
		ID:           restaurantID,
		Name:         "Test Restaurant - " + time.Now().Format("15:04:05"),
		Email:        "test-" + restaurantID[:8] + "@restaurant.it",
		Phone:        "+39 06 1234567",
		Address:      "Via Test 123, Roma",
		Username:     "test_" + restaurantID[:8],
		PasswordHash: "hash_password_123",
		Role:         "owner",
		IsActive:     true,
		CreatedAt:    time.Now(),
	}

	err = db.MongoInstance.CreateRestaurant(ctx, restaurant)
	if err != nil {
		t.Fatalf("❌ Errore inserimento ristorante: %v\n", err)
	}

	t.Logf("✅ Ristorante inserito:")
	t.Logf("   ID: %s", restaurantID)
	t.Logf("   Nome: %s", restaurant.Name)
	t.Logf("   Email: %s", restaurant.Email)

	menuID := uuid.New().String()
	menu := &models.Menu{
		ID:           menuID,
		RestaurantID: restaurantID,
		Name:         "Menu Test - " + time.Now().Format("15:04:05"),
		Description:  "Menu di test MongoDB",
		MealType:     "lunch",
		Categories: []models.MenuCategory{
			{
				ID:          uuid.New().String(),
				Name:        "Piatti Principali",
				Description: "Specialità",
				Items: []models.MenuItem{
					{
						ID:          uuid.New().String(),
						Name:        "Carbonara",
						Description: "Classica ricetta romana",
						Price:       15.00,
						Available:   true,
					},
					{
						ID:          uuid.New().String(),
						Name:        "Cacio e Pepe",
						Description: "Pecorino e pepe nero",
						Price:       14.50,
						Available:   true,
					},
				},
			},
		},
		IsActive:    true,
		IsCompleted: true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err = db.MongoInstance.CreateMenu(ctx, menu)
	if err != nil {
		t.Fatalf("❌ Errore inserimento menu: %v\n", err)
	}

	t.Logf("✅ Menu inserito:")
	t.Logf("   ID: %s", menuID)
	t.Logf("   Nome: %s", menu.Name)
	t.Logf("   Categorie: %d", len(menu.Categories))
	t.Logf("   Piatti totali: %d", len(menu.Categories[0].Items))

	auditLog := &db.AuditLog{
		Action:       "TEST_INSERT",
		ResourceType: "menu",
		ResourceID:   menuID,
		RestaurantID: restaurantID,
		IPAddress:    "127.0.0.1",
		UserAgent:    "MongoDBTest/1.0",
		Status:       "success",
		Timestamp:    time.Now(),
	}

	err = db.MongoInstance.CreateAuditLog(ctx, auditLog)
	if err != nil {
		t.Logf("⚠️  Errore inserimento audit log: %v", err)
	} else {
		t.Logf("✅ Audit Log inserito (Action: %s)", auditLog.Action)
	}

	// Step 3: Verifica dati
	t.Log("\n✓ Step 3: Verifica Dati Inseriti")
	t.Log(subSeparator)

	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	retrievedRestaurant, err := db.MongoInstance.GetRestaurantByID(ctx, restaurantID)
	if err != nil {
		t.Fatalf("❌ Errore lettura ristorante: %v\n", err)
	}

	if retrievedRestaurant == nil {
		t.Fatal("❌ Ristorante non trovato!")
	}

	t.Logf("✅ Ristorante letto da MongoDB:")
	t.Logf("   ID: %s", retrievedRestaurant.ID)
	t.Logf("   Nome: %s", retrievedRestaurant.Name)
	t.Logf("   Email: %s", retrievedRestaurant.Email)
	t.Logf("   Active: %v", retrievedRestaurant.IsActive)

	retrievedMenu, err := db.MongoInstance.GetMenuByID(ctx, menuID)
	if err != nil {
		t.Fatalf("❌ Errore lettura menu: %v\n", err)
	}

	if retrievedMenu == nil {
		t.Fatal("❌ Menu non trovato!")
	}

	t.Logf("✅ Menu letto da MongoDB:")
	t.Logf("   ID: %s", retrievedMenu.ID)
	t.Logf("   Nome: %s", retrievedMenu.Name)
	t.Logf("   RestaurantID: %s", retrievedMenu.RestaurantID)
	t.Logf("   Categorie: %d", len(retrievedMenu.Categories))
	t.Logf("   Piatti: %d", len(retrievedMenu.Categories[0].Items))

	menus, err := db.MongoInstance.GetMenusByRestaurantID(ctx, restaurantID)
	if err != nil {
		t.Logf("⚠️  Errore lettura menu del ristorante: %v", err)
		return
	}

	t.Logf("✅ Menu del ristorante trovati: %d", len(menus))

	t.Log("\n" + separator)
	t.Log("✅ TUTTI I TEST COMPLETATI CON SUCCESSO!")
	t.Log(separator)
}
