package handlers

import (
	"net/http"
)

// TODO: Move restaurant-related handlers into this file: CreateRestaurantHandler,
// SelectRestaurantHandler, UpdateRestaurantHandler, DeleteRestaurantHandler, etc.

func RestaurantModulePlaceholder(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("Restaurant handlers not yet migrated"))
}
