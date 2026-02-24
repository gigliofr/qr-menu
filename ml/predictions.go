package ml

import (
	"math"
	"sort"
	"sync"
	"time"
)

// PredictiveAnalytics provides forecasting and predictive capabilities
type PredictiveAnalytics struct {
	mu              sync.RWMutex
	timeSeries      map[string][]TimeSeriesPoint // metric -> data points
	forecasts       map[string][]ForecastPoint    // metric -> forecasts
	seasonalPatterns map[string]SeasonalPattern   // metric -> pattern
}

// TimeSeriesPoint represents a data point in time series
type TimeSeriesPoint struct {
	Timestamp time.Time
	Value     float64
	Metadata  map[string]interface{}
}

// ForecastPoint represents a forecasted value
type ForecastPoint struct {
	Timestamp      time.Time
	PredictedValue float64
	ConfidenceLow  float64
	ConfidenceHigh float64
	Method         string
}

// SeasonalPattern represents detected seasonal patterns
type SeasonalPattern struct {
	Period        time.Duration
	Amplitude     float64
	Phase         float64
	TrendSlope    float64
	BaselineValue float64
	Detected      bool
}

// Trend represents a trend analysis result
type Trend struct {
	Direction string  // up, down, stable
	Slope     float64
	R2        float64 // R-squared goodness of fit
	StartDate time.Time
	EndDate   time.Time
}

// NewPredictiveAnalytics creates a new predictive analytics engine
func NewPredictiveAnalytics() *PredictiveAnalytics {
	return &PredictiveAnalytics{
		timeSeries:       make(map[string][]TimeSeriesPoint),
		forecasts:        make(map[string][]ForecastPoint),
		seasonalPatterns: make(map[string]SeasonalPattern),
	}
}

// AddDataPoint adds a new data point to a time series
func (pa *PredictiveAnalytics) AddDataPoint(metric string, point TimeSeriesPoint) {
	pa.mu.Lock()
	defer pa.mu.Unlock()
	
	if _, exists := pa.timeSeries[metric]; !exists {
		pa.timeSeries[metric] = make([]TimeSeriesPoint, 0)
	}
	
	pa.timeSeries[metric] = append(pa.timeSeries[metric], point)
	
	// Keep only last 1000 points to prevent unbounded growth
	if len(pa.timeSeries[metric]) > 1000 {
		pa.timeSeries[metric] = pa.timeSeries[metric][len(pa.timeSeries[metric])-1000:]
	}
}

// ForecastDemand predicts future demand using exponential smoothing
func (pa *PredictiveAnalytics) ForecastDemand(metric string, periods int) ([]ForecastPoint, error) {
	pa.mu.RLock()
	data, exists := pa.timeSeries[metric]
	pa.mu.RUnlock()
	
	if !exists || len(data) < 3 {
		return nil, nil // Not enough data
	}
	
	// Use Holt-Winters exponential smoothing
	forecasts := pa.holtWinters(data, periods)
	
	// Store forecasts
	pa.mu.Lock()
	pa.forecasts[metric] = forecasts
	pa.mu.Unlock()
	
	return forecasts, nil
}

// holtWinters implements Holt-Winters exponential smoothing
func (pa *PredictiveAnalytics) holtWinters(data []TimeSeriesPoint, periods int) []ForecastPoint {
	if len(data) < 3 {
		return []ForecastPoint{}
	}
	
	// Parameters (typically optimized, but using defaults here)
	alpha := 0.3 // Level smoothing
	beta := 0.1  // Trend smoothing
	
	// Initialize
	level := data[0].Value
	trend := (data[len(data)-1].Value - data[0].Value) / float64(len(data))
	
	// Smooth the series
	for _, point := range data[1:] {
		prevLevel := level
		level = alpha*point.Value + (1-alpha)*(level+trend)
		trend = beta*(level-prevLevel) + (1-beta)*trend
	}
	
	// Generate forecasts
	forecasts := make([]ForecastPoint, periods)
	lastTimestamp := data[len(data)-1].Timestamp
	
	// Estimate time interval between points
	var interval time.Duration
	if len(data) > 1 {
		interval = data[len(data)-1].Timestamp.Sub(data[len(data)-2].Timestamp)
	} else {
		interval = 24 * time.Hour // Default to daily
	}
	
	for i := 0; i < periods; i++ {
		forecast := level + trend*float64(i+1)
		
		// Calculate confidence interval (simplified)
		stdDev := pa.calculateStdDev(data)
		margin := 1.96 * stdDev // 95% confidence
		
		forecasts[i] = ForecastPoint{
			Timestamp:      lastTimestamp.Add(interval * time.Duration(i+1)),
			PredictedValue: forecast,
			ConfidenceLow:  forecast - margin,
			ConfidenceHigh: forecast + margin,
			Method:         "holt-winters",
		}
	}
	
	return forecasts
}

// DetectSeasonality analyzes data for seasonal patterns
func (pa *PredictiveAnalytics) DetectSeasonality(metric string) SeasonalPattern {
	pa.mu.RLock()
	data, exists := pa.timeSeries[metric]
	pa.mu.RUnlock()
	
	if !exists || len(data) < 14 {
		return SeasonalPattern{Detected: false}
	}
	
	// Simple autocorrelation-based detection
	// Check for daily (24h) and weekly (7 day) patterns
	
	dailyCorr := pa.autocorrelation(data, 24)
	weeklyCorr := pa.autocorrelation(data, 7*24)
	
	pattern := SeasonalPattern{
		Detected: false,
	}
	
	// If correlation is strong, we have seasonality
	if math.Abs(dailyCorr) > 0.5 {
		pattern.Detected = true
		pattern.Period = 24 * time.Hour
		pattern.Amplitude = dailyCorr
	} else if math.Abs(weeklyCorr) > 0.5 {
		pattern.Detected = true
		pattern.Period = 7 * 24 * time.Hour
		pattern.Amplitude = weeklyCorr
	}
	
	// Calculate trend
	if len(data) > 1 {
		pattern.TrendSlope = (data[len(data)-1].Value - data[0].Value) / float64(len(data))
		
		// Baseline is average
		sum := 0.0
		for _, point := range data {
			sum += point.Value
		}
		pattern.BaselineValue = sum / float64(len(data))
	}
	
	pa.mu.Lock()
	pa.seasonalPatterns[metric] = pattern
	pa.mu.Unlock()
	
	return pattern
}

// AnalyzeTrend analyzes the trend in a time series
func (pa *PredictiveAnalytics) AnalyzeTrend(metric string) Trend {
	pa.mu.RLock()
	data, exists := pa.timeSeries[metric]
	pa.mu.RUnlock()
	
	if !exists || len(data) < 2 {
		return Trend{Direction: "unknown"}
	}
	
	// Linear regression
	slope, r2 := pa.linearRegression(data)
	
	trend := Trend{
		Slope:     slope,
		R2:        r2,
		StartDate: data[0].Timestamp,
		EndDate:   data[len(data)-1].Timestamp,
	}
	
	// Determine direction
	if slope > 0.01 {
		trend.Direction = "up"
	} else if slope < -0.01 {
		trend.Direction = "down"
	} else {
		trend.Direction = "stable"
	}
	
	return trend
}

// PredictPeakTimes predicts when demand will peak
func (pa *PredictiveAnalytics) PredictPeakTimes(metric string, lookAhead time.Duration) []time.Time {
	pa.mu.RLock()
	data, exists := pa.timeSeries[metric]
	pattern := pa.seasonalPatterns[metric]
	pa.mu.RUnlock()
	
	if !exists || !pattern.Detected {
		return []time.Time{}
	}
	
	// Find historical peaks
	peaks := pa.findPeaks(data)
	if len(peaks) == 0 {
		return []time.Time{}
	}
	
	// Project peaks forward based on seasonal period
	peakTimes := make([]time.Time, 0)
	lastPeak := peaks[len(peaks)-1].Timestamp
	
	for currentTime := lastPeak.Add(pattern.Period); currentTime.Before(time.Now().Add(lookAhead)); currentTime = currentTime.Add(pattern.Period) {
		peakTimes = append(peakTimes, currentTime)
	}
	
	return peakTimes
}

// OptimizeInventory suggests optimal inventory levels
func (pa *PredictiveAnalytics) OptimizeInventory(itemID string, leadTime time.Duration) map[string]float64 {
	// Forecast demand over lead time
	forecasts, _ := pa.ForecastDemand("item_demand_"+itemID, int(leadTime.Hours()/24))
	
	if len(forecasts) == 0 {
		return map[string]float64{
			"recommended_stock": 0,
			"safety_stock":      0,
		}
	}
	
	// Calculate expected demand
	expectedDemand := 0.0
	for _, forecast := range forecasts {
		expectedDemand += forecast.PredictedValue
	}
	
	// Safety stock (for 95% service level)
	pa.mu.RLock()
	data := pa.timeSeries["item_demand_"+itemID]
	pa.mu.RUnlock()
	
	stdDev := pa.calculateStdDev(data)
	safetyStock := 1.65 * stdDev * math.Sqrt(float64(len(forecasts))) // Z-score for 95%
	
	return map[string]float64{
		"expected_demand":   expectedDemand,
		"safety_stock":      safetyStock,
		"recommended_stock": expectedDemand + safetyStock,
		"service_level":     0.95,
	}
}

// Statistical helper functions

func (pa *PredictiveAnalytics) calculateStdDev(data []TimeSeriesPoint) float64 {
	if len(data) == 0 {
		return 0
	}
	
	// Calculate mean
	sum := 0.0
	for _, point := range data {
		sum += point.Value
	}
	mean := sum / float64(len(data))
	
	// Calculate variance
	variance := 0.0
	for _, point := range data {
		diff := point.Value - mean
		variance += diff * diff
	}
	variance /= float64(len(data))
	
	return math.Sqrt(variance)
}

func (pa *PredictiveAnalytics) autocorrelation(data []TimeSeriesPoint, lag int) float64 {
	if len(data) < lag+1 {
		return 0
	}
	
	// Calculate mean
	sum := 0.0
	for _, point := range data {
		sum += point.Value
	}
	mean := sum / float64(len(data))
	
	// Calculate autocorrelation
	numerator := 0.0
	denominator := 0.0
	
	for i := 0; i < len(data)-lag; i++ {
		numerator += (data[i].Value - mean) * (data[i+lag].Value - mean)
	}
	
	for _, point := range data {
		diff := point.Value - mean
		denominator += diff * diff
	}
	
	if denominator == 0 {
		return 0
	}
	
	return numerator / denominator
}

func (pa *PredictiveAnalytics) linearRegression(data []TimeSeriesPoint) (slope, r2 float64) {
	n := float64(len(data))
	if n < 2 {
		return 0, 0
	}
	
	// Convert timestamps to numeric values (hours since first point)
	var sumX, sumY, sumXY, sumX2, sumY2 float64
	
	firstTime := data[0].Timestamp
	for _, point := range data {
		x := point.Timestamp.Sub(firstTime).Hours()
		y := point.Value
		
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
		sumY2 += y * y
	}
	
	// Calculate slope
	denominator := n*sumX2 - sumX*sumX
	if denominator == 0 {
		return 0, 0
	}
	
	slope = (n*sumXY - sumX*sumY) / denominator
	
	// Calculate R-squared
	meanY := sumY / n
	ssTotal := sumY2 - n*meanY*meanY
	intercept := (sumY - slope*sumX) / n
	ssResidual := 0.0
	
	for _, point := range data {
		x := point.Timestamp.Sub(firstTime).Hours()
		predicted := slope*x + intercept
		residual := point.Value - predicted
		ssResidual += residual * residual
	}
	
	if ssTotal == 0 {
		r2 = 0
	} else {
		r2 = 1 - (ssResidual / ssTotal)
	}
	
	return slope, r2
}

func (pa *PredictiveAnalytics) findPeaks(data []TimeSeriesPoint) []TimeSeriesPoint {
	if len(data) < 3 {
		return []TimeSeriesPoint{}
	}
	
	peaks := make([]TimeSeriesPoint, 0)
	
	for i := 1; i < len(data)-1; i++ {
		if data[i].Value > data[i-1].Value && data[i].Value > data[i+1].Value {
			peaks = append(peaks, data[i])
		}
	}
	
	// Sort by value descending
	sort.Slice(peaks, func(i, j int) bool {
		return peaks[i].Value > peaks[j].Value
	})
	
	return peaks
}

// GetMetrics returns analytics metrics
func (pa *PredictiveAnalytics) GetMetrics() map[string]interface{} {
	pa.mu.RLock()
	defer pa.mu.RUnlock()
	
	metrics := make(map[string]interface{})
	
	for metric, data := range pa.timeSeries {
		if len(data) > 0 {
			trend := pa.AnalyzeTrend(metric)
			metrics[metric] = map[string]interface{}{
				"data_points": len(data),
				"latest_value": data[len(data)-1].Value,
				"trend":       trend.Direction,
				"trend_slope": trend.Slope,
			}
		}
	}
	
	return metrics
}
