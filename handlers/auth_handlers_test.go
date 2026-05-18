package handlers

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"qr-menu/db"
	"qr-menu/models"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

// TestLoginHandlerValidCredentials tests login with valid credentials
func TestLoginHandlerValidCredentials(t *testing.T) {
	if db.MongoInstance == nil {
		t.Skip("MongoDB not available")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create test user
	testUsername := "test_user_" + uuid.New().String()[:8]
	testPassword := "TestPassword123!"
	testEmail := testUsername + "@test.com"

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(testPassword), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	// Create user in DB
	testUser := &models.User{
		ID:           uuid.New().String(),
		Username:     testUsername,
		PasswordHash: string(hashedPassword),
		Email:        testEmail,
		CreatedAt:    time.Now(),
		IsActive:     true,
	}

	if err := db.MongoInstance.CreateUser(ctx, testUser); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Prepare login request as form data (LoginHandler expects form values)
	form := "username=" + testUsername + "&password=" + testPassword
	req, err := http.NewRequest("POST", "/login", bytes.NewReader([]byte(form)))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Execute request via LoginHandler
	w := httptest.NewRecorder()
	LoginHandler(w, req)

	// Expect a redirect (user has no restaurants in this test) to /add-restaurant
	if w.Result().StatusCode != http.StatusFound {
		t.Fatalf("Expected redirect after login, got status %d", w.Result().StatusCode)
	}
	loc, _ := w.Result().Location()
	if loc.Path != "/add-restaurant" {
		t.Fatalf("Expected redirect to /add-restaurant, got %s", loc.Path)
	}

	// Cleanup
	db.MongoInstance.DeleteUser(ctx, testUser.ID)
}

// TestMenuCreationWithOwnershipValidation tests menu creation with restaurant ownership check
func TestMenuCreationWithOwnershipValidation(t *testing.T) {
	if db.MongoInstance == nil {
		t.Skip("MongoDB not available")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create test restaurant
	restaurantID := uuid.New().String()
	ownerID := uuid.New().String()

	restaurant := &models.Restaurant{
		ID:        restaurantID,
		OwnerID:   ownerID,
		Name:      "Test Restaurant",
		Email:     "test@restaurant.com",
		Username:  "test_rest_" + uuid.New().String()[:8],
		Address:   "Test Address",
		Phone:     "+39 123 456 7890",
		CreatedAt: time.Now(),
		IsActive:  true,
	}

	if err := db.MongoInstance.CreateRestaurant(ctx, restaurant); err != nil {
		t.Fatalf("Failed to create test restaurant: %v", err)
	}

	// Create test menu
	menu := &models.Menu{
		ID:           uuid.New().String(),
		RestaurantID: restaurantID,
		Name:         "Test Menu",
		Description:  "Test menu for ownership validation",
		IsActive:     false,
		CreatedAt:    time.Now(),
	}

	if err := db.MongoInstance.CreateMenu(ctx, menu); err != nil {
		t.Fatalf("Failed to create test menu: %v", err)
	}

	// Verify menu belongs to correct restaurant
	retrievedMenu, err := db.MongoInstance.GetMenuByID(ctx, menu.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve menu: %v", err)
	}

	if retrievedMenu.RestaurantID != restaurantID {
		t.Errorf("Menu ownership validation failed: expected restaurant %s, got %s", restaurantID, retrievedMenu.RestaurantID)
	}

	// Cleanup
	db.MongoInstance.DeleteMenu(ctx, menu.ID)
	db.MongoInstance.DeleteRestaurant(ctx, restaurantID)
}

// TestSetActiveMenuTransactionConsistency tests that setting active menu maintains consistency
func TestSetActiveMenuTransactionConsistency(t *testing.T) {
	if db.MongoInstance == nil {
		t.Skip("MongoDB not available")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create test restaurant
	restaurantID := uuid.New().String()
	
	restaurant := &models.Restaurant{
		ID:        restaurantID,
		OwnerID:   uuid.New().String(),
		Name:      "Consistency Test Restaurant",
		Email:     "consistency@test.com",
		Username:  "consistency_" + uuid.New().String()[:8],
		CreatedAt: time.Now(),
		IsActive:  true,
	}

	if err := db.MongoInstance.CreateRestaurant(ctx, restaurant); err != nil {
		t.Fatalf("Failed to create test restaurant: %v", err)
	}

	// Create multiple menus
	menuA := &models.Menu{
		ID:           uuid.New().String(),
		RestaurantID: restaurantID,
		Name:         "Menu A",
		IsActive:     true,
		CreatedAt:    time.Now(),
	}

	menuB := &models.Menu{
		ID:           uuid.New().String(),
		RestaurantID: restaurantID,
		Name:         "Menu B",
		IsActive:     false,
		CreatedAt:    time.Now(),
	}

	if err := db.MongoInstance.CreateMenu(ctx, menuA); err != nil {
		t.Fatalf("Failed to create menu A: %v", err)
	}
	if err := db.MongoInstance.CreateMenu(ctx, menuB); err != nil {
		t.Fatalf("Failed to create menu B: %v", err)
	}

	// Deactivate all, then activate menu B
	// First: deactivate all active menus
	if err := db.MongoInstance.UpdateManyMenus(ctx,
		bson.M{"restaurant_id": restaurantID, "is_active": true},
		bson.M{"is_active": false},
	); err != nil {
		t.Fatalf("Failed to deactivate all menus: %v", err)
	}

	// Activate menu B
	menuB.IsActive = true
	if err := db.MongoInstance.UpdateMenu(ctx, menuB); err != nil {
		t.Fatalf("Failed to activate menu B: %v", err)
	}

	// Verify only menu B is active
	checkMenuA, _ := db.MongoInstance.GetMenuByID(ctx, menuA.ID)
	checkMenuB, _ := db.MongoInstance.GetMenuByID(ctx, menuB.ID)

	if checkMenuA.IsActive {
		t.Errorf("Menu A should be inactive but got IsActive=true")
	}
	if !checkMenuB.IsActive {
		t.Errorf("Menu B should be active but got IsActive=false")
	}

	// Cleanup
	db.MongoInstance.DeleteMenu(ctx, menuA.ID)
	db.MongoInstance.DeleteMenu(ctx, menuB.ID)
	db.MongoInstance.DeleteRestaurant(ctx, restaurantID)
}
