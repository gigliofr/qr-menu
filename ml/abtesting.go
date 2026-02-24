package ml

import (
	"crypto/rand"
	"encoding/hex"
	"math"
	"sync"
	"time"
)

// ABTestManager manages A/B testing experiments
type ABTestManager struct {
	mu          sync.RWMutex
	experiments map[string]*Experiment
	assignments map[string]string // userID -> variantID
	results     map[string]map[string]*VariantStats // experimentID -> variantID -> stats
}

// Experiment represents an A/B test configuration
type Experiment struct {
	ID          string
	Name        string
	Description string
	Status      string // draft, running, paused, completed
	StartDate   time.Time
	EndDate     time.Time
	Variants    []Variant
	Metric      string // Primary metric to optimize
	SampleSize  int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Variant represents a test variant
type Variant struct {
	ID          string
	Name        string
	Description string
	Traffic     float64 // Percentage of traffic (0-1)
	IsControl   bool
	Config      map[string]interface{} // Variant configuration
}

// VariantStats tracks statistics for a variant
type VariantStats struct {
	VariantID       string
	Impressions     int64
	Conversions     int64
	ConversionRate  float64
	Revenue         float64
	AverageRevenue  float64
	Participants    int64
	LastUpdated     time.Time
}

// ABTestResult represents the result of an A/B test
type ABTestResult struct {
	ExperimentID   string
	Winner         string
	WinnerVariant  *VariantStats
	AllVariants    []*VariantStats
	StatSignificant bool
	PValue         float64
	ConfidenceLevel float64
	Recommendation  string
}

// ConversionEvent represents a conversion in an experiment
type ConversionEvent struct {
	UserID       string
	ExperimentID string
	VariantID    string
	EventType    string
	Value        float64
	Timestamp    time.Time
	Metadata     map[string]interface{}
}

// NewABTestManager creates a new A/B test manager
func NewABTestManager() *ABTestManager {
	return &ABTestManager{
		experiments: make(map[string]*Experiment),
		assignments: make(map[string]string),
		results:     make(map[string]map[string]*VariantStats),
	}
}

// CreateExperiment creates a new A/B test experiment
func (ab *ABTestManager) CreateExperiment(exp Experiment) (*Experiment, error) {
	ab.mu.Lock()
	defer ab.mu.Unlock()
	
	// Generate ID if not provided
	if exp.ID == "" {
		exp.ID = generateID()
	}
	
	// Validate traffic allocation
	totalTraffic := 0.0
	for _, variant := range exp.Variants {
		totalTraffic += variant.Traffic
	}
	if math.Abs(totalTraffic-1.0) > 0.01 {
		// Auto-normalize
		for i := range exp.Variants {
			exp.Variants[i].Traffic /= totalTraffic
		}
	}
	
	// Set timestamps
	exp.CreatedAt = time.Now()
	exp.UpdatedAt = time.Now()
	exp.Status = "draft"
	
	// Initialize stats
	ab.results[exp.ID] = make(map[string]*VariantStats)
	for _, variant := range exp.Variants {
		ab.results[exp.ID][variant.ID] = &VariantStats{
			VariantID:   variant.ID,
			LastUpdated: time.Now(),
		}
	}
	
	ab.experiments[exp.ID] = &exp
	return &exp, nil
}

// StartExperiment starts an experiment
func (ab *ABTestManager) StartExperiment(experimentID string) error {
	ab.mu.Lock()
	defer ab.mu.Unlock()
	
	exp, exists := ab.experiments[experimentID]
	if !exists {
		return nil
	}
	
	exp.Status = "running"
	exp.StartDate = time.Now()
	exp.UpdatedAt = time.Now()
	
	return nil
}

// StopExperiment stops an experiment
func (ab *ABTestManager) StopExperiment(experimentID string) error {
	ab.mu.Lock()
	defer ab.mu.Unlock()
	
	exp, exists := ab.experiments[experimentID]
	if !exists {
		return nil
	}
	
	exp.Status = "completed"
	exp.EndDate = time.Now()
	exp.UpdatedAt = time.Now()
	
	return nil
}

// AssignVariant assigns a user to a variant
func (ab *ABTestManager) AssignVariant(experimentID, userID string) (string, error) {
	ab.mu.Lock()
	defer ab.mu.Unlock()
	
	// Check if already assigned
	key := experimentID + ":" + userID
	if variantID, exists := ab.assignments[key]; exists {
		return variantID, nil
	}
	
	exp, exists := ab.experiments[experimentID]
	if !exists || exp.Status != "running" {
		return "", nil
	}
	
	// Random assignment based on traffic allocation
	r := randomFloat()
	cumulative := 0.0
	
	for _, variant := range exp.Variants {
		cumulative += variant.Traffic
		if r <= cumulative {
			ab.assignments[key] = variant.ID
			
			// Increment impressions
			if stats, exists := ab.results[experimentID][variant.ID]; exists {
				stats.Impressions++
				stats.Participants++
			}
			
			return variant.ID, nil
		}
	}
	
	// Fallback to first variant
	if len(exp.Variants) > 0 {
		variantID := exp.Variants[0].ID
		ab.assignments[key] = variantID
		return variantID, nil
	}
	
	return "", nil
}

// GetVariant retrieves the assigned variant for a user
func (ab *ABTestManager) GetVariant(experimentID, userID string) (string, *Variant) {
	ab.mu.RLock()
	defer ab.mu.RUnlock()
	
	key := experimentID + ":" + userID
	variantID, exists := ab.assignments[key]
	if !exists {
		return "", nil
	}
	
	exp, exists := ab.experiments[experimentID]
	if !exists {
		return "", nil
	}
	
	for _, variant := range exp.Variants {
		if variant.ID == variantID {
			return variantID, &variant
		}
	}
	
	return variantID, nil
}

// TrackConversion records a conversion event
func (ab *ABTestManager) TrackConversion(event ConversionEvent) error {
	ab.mu.Lock()
	defer ab.mu.Unlock()
	
	// Get user's variant assignment
	key := event.ExperimentID + ":" + event.UserID
	variantID, exists := ab.assignments[key]
	if !exists {
		return nil // User not in experiment
	}
	
	// Update stats
	if stats, exists := ab.results[event.ExperimentID][variantID]; exists {
		stats.Conversions++
		stats.Revenue += event.Value
		stats.ConversionRate = float64(stats.Conversions) / float64(stats.Impressions)
		if stats.Conversions > 0 {
			stats.AverageRevenue = stats.Revenue / float64(stats.Conversions)
		}
		stats.LastUpdated = time.Now()
	}
	
	return nil
}

// GetExperimentResults calculates and returns experiment results
func (ab *ABTestManager) GetExperimentResults(experimentID string) *ABTestResult {
	ab.mu.RLock()
	defer ab.mu.RUnlock()
	
	exp, exists := ab.experiments[experimentID]
	if !exists {
		return nil
	}
	
	stats, exists := ab.results[experimentID]
	if !exists {
		return nil
	}
	
	// Collect all variant stats
	allVariants := make([]*VariantStats, 0, len(stats))
	var controlStats *VariantStats
	
	for _, s := range stats {
		allVariants = append(allVariants, s)
		
		// Find control variant
		for _, v := range exp.Variants {
			if v.ID == s.VariantID && v.IsControl {
				controlStats = s
			}
		}
	}
	
	// Find winner (highest conversion rate)
	var winner *VariantStats
	for _, s := range allVariants {
		if winner == nil || s.ConversionRate > winner.ConversionRate {
			winner = s
		}
	}
	
	result := &ABTestResult{
		ExperimentID:  experimentID,
		AllVariants:   allVariants,
		WinnerVariant: winner,
	}
	
	if winner != nil {
		result.Winner = winner.VariantID
	}
	
	// Calculate statistical significance if we have control and winner
	if controlStats != nil && winner != nil && winner.VariantID != controlStats.VariantID {
		pValue := ab.calculatePValue(controlStats, winner)
		result.PValue = pValue
		result.StatSignificant = pValue < 0.05
		result.ConfidenceLevel = 1 - pValue
		
		if result.StatSignificant {
			improvement := ((winner.ConversionRate - controlStats.ConversionRate) / controlStats.ConversionRate) * 100
			result.Recommendation = formatRecommendation(winner.VariantID, improvement)
		} else {
			result.Recommendation = "Not enough data for statistical significance. Continue running experiment."
		}
	}
	
	return result
}

// calculatePValue calculates p-value using Z-test for proportions
func (ab *ABTestManager) calculatePValue(control, variant *VariantStats) float64 {
	// Check sample sizes
	if control.Impressions < 30 || variant.Impressions < 30 {
		return 1.0 // Not enough data
	}
	
	p1 := control.ConversionRate
	p2 := variant.ConversionRate
	n1 := float64(control.Impressions)
	n2 := float64(variant.Impressions)
	
	// Pooled proportion
	pooled := (float64(control.Conversions) + float64(variant.Conversions)) / (n1 + n2)
	
	// Standard error
	se := math.Sqrt(pooled * (1 - pooled) * ((1 / n1) + (1 / n2)))
	
	if se == 0 {
		return 1.0
	}
	
	// Z-score
	z := (p2 - p1) / se
	
	// Two-tailed p-value (simplified approximation)
	pValue := 2 * (1 - normalCDF(math.Abs(z)))
	
	return pValue
}

// GetAllExperiments returns all experiments
func (ab *ABTestManager) GetAllExperiments() []*Experiment {
	ab.mu.RLock()
	defer ab.mu.RUnlock()
	
	experiments := make([]*Experiment, 0, len(ab.experiments))
	for _, exp := range ab.experiments {
		experiments = append(experiments, exp)
	}
	
	return experiments
}

// GetExperiment retrieves a specific experiment
func (ab *ABTestManager) GetExperiment(experimentID string) *Experiment {
	ab.mu.RLock()
	defer ab.mu.RUnlock()
	
	return ab.experiments[experimentID]
}

// GetVariantStats retrieves stats for a specific variant
func (ab *ABTestManager) GetVariantStats(experimentID, variantID string) *VariantStats {
	ab.mu.RLock()
	defer ab.mu.RUnlock()
	
	if stats, exists := ab.results[experimentID]; exists {
		return stats[variantID]
	}
	
	return nil
}

// Helper functions

func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func randomFloat() float64 {
	b := make([]byte, 8)
	rand.Read(b)
	
	// Convert to uint64
	var n uint64
	for i := 0; i < 8; i++ {
		n = (n << 8) | uint64(b[i])
	}
	
	// Convert to float64 in range [0, 1)
	return float64(n) / float64(^uint64(0))
}

// normalCDF approximates the normal cumulative distribution function
func normalCDF(x float64) float64 {
	// Approximation using error function
	return 0.5 * (1 + math.Erf(x/math.Sqrt(2)))
}

func formatRecommendation(variantID string, improvement float64) string {
	if improvement > 0 {
		return variantID + " performs better"
	}
	return "No significant improvement"
}

// GetStats returns overall A/B testing statistics
func (ab *ABTestManager) GetStats() map[string]interface{} {
	ab.mu.RLock()
	defer ab.mu.RUnlock()
	
	runningCount := 0
	completedCount := 0
	
	for _, exp := range ab.experiments {
		switch exp.Status {
		case "running":
			runningCount++
		case "completed":
			completedCount++
		}
	}
	
	return map[string]interface{}{
		"total_experiments":     len(ab.experiments),
		"running_experiments":   runningCount,
		"completed_experiments": completedCount,
		"total_assignments":     len(ab.assignments),
	}
}
