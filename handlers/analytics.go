package handlers

import (
	"net/http"
)

// TODO: Move analytics and tracking handlers here: AnalyticsHandler, TrackEventHandler, etc.

func AnalyticsModulePlaceholder(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("Analytics handlers not yet migrated"))
}
