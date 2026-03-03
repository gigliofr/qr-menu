package api

import (
	"context"
	"net/http"
	"qr-menu/db"
	"qr-menu/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

// DebugMenusHandler - endpoint temporaneo per debuggare il problema dei menu
func DebugMenusHandler(w http.ResponseWriter, r *http.Request) {
	restaurantID := GetRestaurantIDFromRequest(r)
	
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	
	// 1. Verifica connessione MongoDB
	if db.MongoInstance == nil {
		ErrorResponse(w, 500, "NO_DB", "Database non connesso", "")
		return
	}
	
	// 2. Query diretta per vedere TUTTI i menu nel DB
	coll := db.MongoInstance.GetDB().Collection("menus")
	
	cursor, err := coll.Find(ctx, bson.M{})
	if err != nil {
		ErrorResponse(w, 500, "QUERY_ERROR", "Errore query", err.Error())
		return
	}
	defer cursor.Close(ctx)
	
	var allMenus []bson.M
	if err = cursor.All(ctx, &allMenus); err != nil {
		ErrorResponse(w, 500, "DECODE_ERROR", "Errore decodifica", err.Error())
		return
	}
	
	// 3. Query filtrata per restaurant_id
	cursor2, err := coll.Find(ctx, bson.M{"restaurant_id": restaurantID})
	if err != nil {
		ErrorResponse(w, 500, "FILTERED_QUERY_ERROR", "Errore query filtrata", err.Error())
		return
	}
	defer cursor2.Close(ctx)
	
	var filteredMenus []*models.Menu
	if err = cursor2.All(ctx, &filteredMenus); err != nil {
		ErrorResponse(w, 500, "FILTERED_DECODE_ERROR", "Errore decodifica filtrata", err.Error())
		return
	}
	
	// 4. Risposta debug
	debugInfo := map[string]interface{}{
		"restaurant_id":          restaurantID,
		"total_menus_in_db":      len(allMenus),
		"filtered_menus_count":   len(filteredMenus),
		"all_menus_raw":          allMenus,
		"filtered_menus_decoded": filteredMenus,
	}
	
	SuccessResponse(w, debugInfo, nil)
}
