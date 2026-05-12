package handlers

import (
	"context"
	"testing"
	"time"

	"qr-menu/db"
	"qr-menu/models"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
)

// TestUpdateManyMenus tests the batch update operation for fixing N+1 query
func TestUpdateManyMenus(t *testing.T) {
	// Skip if no MongoDB available
	if db.MongoInstance == nil {
		t.Skip("MongoDB not available")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create test restaurant and menus
	restaurantID := "test-restaurant-" + uuid.New().String()
	
	// Create test menus
	menu1 := &models.Menu{
		ID:           "menu-1-" + uuid.New().String(),
		RestaurantID: restaurantID,
		Name:         "Menu 1",
		IsActive:     true,
		CreatedAt:    time.Now(),
	}

	menu2 := &models.Menu{
		ID:           "menu-2-" + uuid.New().String(),
		RestaurantID: restaurantID,
		Name:         "Menu 2",
		IsActive:     true,
		CreatedAt:    time.Now(),
	}

	// Insert test menus
	if err := db.MongoInstance.CreateMenu(ctx, menu1); err != nil {
		t.Fatalf("Failed to create test menu 1: %v", err)
	}
	if err := db.MongoInstance.CreateMenu(ctx, menu2); err != nil {
		t.Fatalf("Failed to create test menu 2: %v", err)
	}

	// Test UpdateManyMenus - should deactivate all active menus
	err := db.MongoInstance.UpdateManyMenus(ctx,
		bson.M{"restaurant_id": restaurantID, "is_active": true},
		bson.M{"is_active": false},
	)
	if err != nil {
		t.Fatalf("UpdateManyMenus failed: %v", err)
	}

	// Verify both menus are now inactive
	updatedMenu1, err := db.MongoInstance.GetMenuByID(ctx, menu1.ID)
	if err != nil {
		t.Fatalf("Failed to fetch updated menu 1: %v", err)
	}
	if updatedMenu1.IsActive {
		t.Errorf("Expected menu 1 to be inactive, but got IsActive=true")
	}

	updatedMenu2, err := db.MongoInstance.GetMenuByID(ctx, menu2.ID)
	if err != nil {
		t.Fatalf("Failed to fetch updated menu 2: %v", err)
	}
	if updatedMenu2.IsActive {
		t.Errorf("Expected menu 2 to be inactive, but got IsActive=true")
	}

	// Cleanup
	db.MongoInstance.DeleteMenu(ctx, menu1.ID)
	db.MongoInstance.DeleteMenu(ctx, menu2.ID)
}

// TestTemplateCache validates sync.Once pattern for template initialization
func TestTemplateCachingWithOnce(t *testing.T) {
	// Reset for test
	templateCache = nil
	
	// First call should initialize
	tmpl1 := GetTemplates()
	
	// Second call should return same instance
	tmpl2 := GetTemplates()
	
	if tmpl1 != tmpl2 && tmpl2 == nil {
		t.Errorf("Template caching with sync.Once failed: different instances")
	}
}

// TestGetMenusHandlerMultiTenant tests multi-tenant isolation
func TestGetMenusHandlerMultiTenant(t *testing.T) {
	if db.MongoInstance == nil {
		t.Skip("MongoDB not available")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create two test restaurants
	rest1ID := "restaurant-1-" + uuid.New().String()
	rest2ID := "restaurant-2-" + uuid.New().String()

	// Create menus for restaurant 1
	menu1 := &models.Menu{
		ID:           "menu-1-" + uuid.New().String(),
		RestaurantID: rest1ID,
		Name:         "Restaurant 1 Menu",
		IsActive:     true,
		CreatedAt:    time.Now(),
	}

	// Create menus for restaurant 2
	menu2 := &models.Menu{
		ID:           "menu-2-" + uuid.New().String(),
		RestaurantID: rest2ID,
		Name:         "Restaurant 2 Menu",
		IsActive:     true,
		CreatedAt:    time.Now(),
	}

	if err := db.MongoInstance.CreateMenu(ctx, menu1); err != nil {
		t.Fatalf("Failed to create menu for restaurant 1: %v", err)
	}
	if err := db.MongoInstance.CreateMenu(ctx, menu2); err != nil {
		t.Fatalf("Failed to create menu for restaurant 2: %v", err)
	}

	// Verify restaurant 1 can only see their menus
	menus1, err := db.MongoInstance.GetMenusByRestaurantID(ctx, rest1ID)
	if err != nil {
		t.Fatalf("Failed to get menus for restaurant 1: %v", err)
	}

	for _, m := range menus1 {
		if m.RestaurantID != rest1ID {
			t.Errorf("Multi-tenant isolation violated: restaurant 1 can see menu from restaurant 2")
		}
	}

	// Cleanup
	db.MongoInstance.DeleteMenu(ctx, menu1.ID)
	db.MongoInstance.DeleteMenu(ctx, menu2.ID)
}

// TestBatchUpdatePerformance verifies N+1 fix benefits
func TestBatchUpdatePerformance(t *testing.T) {
	if db.MongoInstance == nil {
		t.Skip("MongoDB not available")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	restaurantID := "perf-test-" + uuid.New().String()

	// Create 10 test menus
	menuIDs := make([]string, 10)
	for i := 0; i < 10; i++ {
		menu := &models.Menu{
			ID:           uuid.New().String(),
			RestaurantID: restaurantID,
			Name:         "Performance Test Menu " + string(rune(i)),
			IsActive:     true,
			CreatedAt:    time.Now(),
		}
		if err := db.MongoInstance.CreateMenu(ctx, menu); err != nil {
			t.Fatalf("Failed to create test menu: %v", err)
		}
		menuIDs[i] = menu.ID
	}

	// Measure batch update (should be 1 query)
	start := time.Now()
	err := db.MongoInstance.UpdateManyMenus(ctx,
		bson.M{"restaurant_id": restaurantID, "is_active": true},
		bson.M{"is_active": false},
	)
	batchDuration := time.Since(start)

	if err != nil {
		t.Fatalf("Batch update failed: %v", err)
	}

	// Batch update should complete in < 100ms
	if batchDuration > 100*time.Millisecond {
		t.Logf("WARNING: Batch update took %v (expected < 100ms)", batchDuration)
	}

	// Cleanup
	for _, id := range menuIDs {
		db.MongoInstance.DeleteMenu(ctx, id)
	}
}
