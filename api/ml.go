package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"qr-menu/ml"

	"github.com/gorilla/mux"
)

// ML & Analytics API Handlers

var (
	recommendationEngine *ml.RecommendationEngine
	predictiveAnalytics  *ml.PredictiveAnalytics
	abTestManager        *ml.ABTestManager
)

func init() {
	// Initialize ML components
	recommendationEngine = ml.NewRecommendationEngine(ml.RecommendationConfig{
		MinTrainingData:    20,
		SimilarityMetric:   "cosine",
		MaxRecommendations: 10,
	})
	
	predictiveAnalytics = ml.NewPredictiveAnalytics()
	abTestManager = ml.NewABTestManager()
}

// Recommendation API Handlers

// GetRecommendationsHandler returns personalized recommendations for a user
func GetRecommendationsHandler(w http.ResponseWriter, r *http.Request) {
	restaurantID := GetRestaurantIDFromRequest(r)
	
	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}
	
	excludeItems := r.URL.Query()["exclude"]
	
	recommendations := recommendationEngine.GetRecommendations(restaurantID, excludeItems, limit)
	
	SuccessResponse(w, recommendations, nil)
}

// GetSimilarItemsHandler returns items similar to a given item
func GetSimilarItemsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	itemID := vars["id"]
	
	limitStr := r.URL.Query().Get("limit")
	limit := 5
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}
	
	similar := recommendationEngine.GetSimilarItems(itemID, limit)
	
	SuccessResponse(w, similar, nil)
}

// GetTrendingItemsHandler returns trending items
func GetTrendingItemsHandler(w http.ResponseWriter, r *http.Request) {
	windowStr := r.URL.Query().Get("window")
	window := 24 * time.Hour
	if windowStr != "" {
		if hours, err := strconv.Atoi(windowStr); err == nil {
			window = time.Duration(hours) * time.Hour
		}
	}
	
	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}
	
	trending := recommendationEngine.GetTrendingItems(window, limit)
	
	SuccessResponse(w, trending, nil)
}

// TrackInteractionHandler tracks user interactions for recommendations
func TrackInteractionHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID          string  `json:"user_id"`
		ItemID          string  `json:"item_id"`
		InteractionType string  `json:"interaction_type"`
		Weight          float64 `json:"weight"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request", err.Error())
		return
	}
	
	if req.Weight == 0 {
		req.Weight = 1.0
	}
	
	recommendationEngine.RecordInteraction(req.UserID, req.ItemID, req.InteractionType, req.Weight)
	
	SuccessResponse(w, map[string]interface{}{
		"status": "recorded",
	}, nil)
}

// TrainRecommendationsHandler triggers recommendation model training
func TrainRecommendationsHandler(w http.ResponseWriter, r *http.Request) {
	err := recommendationEngine.Train()
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "TRAINING_ERROR", "Failed to train model", err.Error())
		return
	}
	
	stats := recommendationEngine.GetStats()
	SuccessResponse(w, map[string]interface{}{
		"status": "trained",
		"stats":  stats,
	}, nil)
}

// Predictive Analytics API Handlers

// ForecastDemandHandler forecasts future demand
func ForecastDemandHandler(w http.ResponseWriter, r *http.Request) {
	metric := r.URL.Query().Get("metric")
	if metric == "" {
		metric = "orders"
	}
	
	periodsStr := r.URL.Query().Get("periods")
	periods := 7 // Default 7 periods
	if periodsStr != "" {
		if p, err := strconv.Atoi(periodsStr); err == nil {
			periods = p
		}
	}
	
	forecasts, err := predictiveAnalytics.ForecastDemand(metric, periods)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "FORECAST_ERROR", "Failed to forecast demand", err.Error())
		return
	}
	
	SuccessResponse(w, forecasts, nil)
}

// DetectSeasonalityHandler detects seasonal patterns
func DetectSeasonalityHandler(w http.ResponseWriter, r *http.Request) {
	metric := r.URL.Query().Get("metric")
	if metric == "" {
		metric = "orders"
	}
	
	pattern := predictiveAnalytics.DetectSeasonality(metric)
	
	SuccessResponse(w, pattern, nil)
}

// AnalyzeTrendHandler analyzes trends in data
func AnalyzeTrendHandler(w http.ResponseWriter, r *http.Request) {
	metric := r.URL.Query().Get("metric")
	if metric == "" {
		metric = "revenue"
	}
	
	trend := predictiveAnalytics.AnalyzeTrend(metric)
	
	SuccessResponse(w, trend, nil)
}

// PredictPeakTimesHandler predicts peak demand times
func PredictPeakTimesHandler(w http.ResponseWriter, r *http.Request) {
	metric := r.URL.Query().Get("metric")
	if metric == "" {
		metric = "orders"
	}
	
	lookAheadStr := r.URL.Query().Get("lookahead_hours")
	lookAhead := 168 * time.Hour // Default 1 week
	if lookAheadStr != "" {
		if hours, err := strconv.Atoi(lookAheadStr); err == nil {
			lookAhead = time.Duration(hours) * time.Hour
		}
	}
	
	peakTimes := predictiveAnalytics.PredictPeakTimes(metric, lookAhead)
	
	SuccessResponse(w, map[string]interface{}{
		"peak_times": peakTimes,
		"metric":     metric,
		"lookahead":  lookAhead.String(),
	}, nil)
}

// OptimizeInventoryHandler suggests optimal inventory levels
func OptimizeInventoryHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	itemID := vars["item_id"]
	
	leadTimeStr := r.URL.Query().Get("lead_time_days")
	leadTime := 7 * 24 * time.Hour // Default 7 days
	if leadTimeStr != "" {
		if days, err := strconv.Atoi(leadTimeStr); err == nil {
			leadTime = time.Duration(days) * 24 * time.Hour
		}
	}
	
	optimization := predictiveAnalytics.OptimizeInventory(itemID, leadTime)
	
	SuccessResponse(w, optimization, nil)
}

// AddDataPointHandler adds a data point to time series
func AddDataPointHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Metric    string                 `json:"metric"`
		Value     float64                `json:"value"`
		Timestamp time.Time              `json:"timestamp"`
		Metadata  map[string]interface{} `json:"metadata"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request", err.Error())
		return
	}
	
	if req.Timestamp.IsZero() {
		req.Timestamp = time.Now()
	}
	
	point := ml.TimeSeriesPoint{
		Timestamp: req.Timestamp,
		Value:     req.Value,
		Metadata:  req.Metadata,
	}
	
	predictiveAnalytics.AddDataPoint(req.Metric, point)
	
	SuccessResponse(w, map[string]interface{}{
		"status": "recorded",
	}, nil)
}

// A/B Testing API Handlers

// CreateExperimentHandler creates a new A/B test
func CreateExperimentHandler(w http.ResponseWriter, r *http.Request) {
	var req ml.Experiment
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request", err.Error())
		return
	}
	
	experiment, err := abTestManager.CreateExperiment(req)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, "EXPERIMENT_ERROR", "Failed to create experiment", err.Error())
		return
	}
	
	SuccessResponse(w, experiment, nil)
}

// StartExperimentHandler starts an A/B test
func StartExperimentHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	experimentID := vars["id"]
	
	err := abTestManager.StartExperiment(experimentID)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, "START_ERROR", "Failed to start experiment", err.Error())
		return
	}
	
	experiment := abTestManager.GetExperiment(experimentID)
	SuccessResponse(w, experiment, nil)
}

// StopExperimentHandler stops an A/B test
func StopExperimentHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	experimentID := vars["id"]
	
	err := abTestManager.StopExperiment(experimentID)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, "STOP_ERROR", "Failed to stop experiment", err.Error())
		return
	}
	
	experiment := abTestManager.GetExperiment(experimentID)
	SuccessResponse(w, experiment, nil)
}

// GetExperimentResultsHandler returns A/B test results
func GetExperimentResultsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	experimentID := vars["id"]
	
	results := abTestManager.GetExperimentResults(experimentID)
	if results == nil {
		ErrorResponse(w, http.StatusNotFound, "NOT_FOUND", "Experiment not found", "")
		return
	}
	
	SuccessResponse(w, results, nil)
}

// AssignVariantHandler assigns a user to a variant
func AssignVariantHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	experimentID := vars["id"]
	
	var req struct {
		UserID string `json:"user_id"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request", err.Error())
		return
	}
	
	variantID, err := abTestManager.AssignVariant(experimentID, req.UserID)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, "ASSIGNMENT_ERROR", "Failed to assign variant", err.Error())
		return
	}
	
	_, variant := abTestManager.GetVariant(experimentID, req.UserID)
	
	SuccessResponse(w, map[string]interface{}{
		"variant_id": variantID,
		"variant":    variant,
	}, nil)
}

// TrackConversionHandler tracks a conversion event
func TrackConversionHandler(w http.ResponseWriter, r *http.Request) {
	var req ml.ConversionEvent
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request", err.Error())
		return
	}
	
	if req.Timestamp.IsZero() {
		req.Timestamp = time.Now()
	}
	
	err := abTestManager.TrackConversion(req)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, "CONVERSION_ERROR", "Failed to track conversion", err.Error())
		return
	}
	
	SuccessResponse(w, map[string]interface{}{
		"status": "recorded",
	}, nil)
}

// ListExperimentsHandler lists all experiments
func ListExperimentsHandler(w http.ResponseWriter, r *http.Request) {
	experiments := abTestManager.GetAllExperiments()
	
	SuccessResponse(w, experiments, nil)
}

// GetMLStatsHandler returns overall ML statistics
func GetMLStatsHandler(w http.ResponseWriter, r *http.Request) {
	stats := map[string]interface{}{
		"recommendations": recommendationEngine.GetStats(),
		"predictive":      predictiveAnalytics.GetMetrics(),
		"ab_testing":      abTestManager.GetStats(),
	}
	
	SuccessResponse(w, stats, nil)
}
