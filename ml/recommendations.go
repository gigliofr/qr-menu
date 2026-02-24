package ml

import (
	"math"
	"sort"
	"sync"
	"time"
)

// RecommendationEngine provides item recommendations based on collaborative filtering
type RecommendationEngine struct {
	mu                sync.RWMutex
	userItemMatrix    map[string]map[string]float64 // user -> item -> score
	itemSimilarity    map[string]map[string]float64 // item -> item -> similarity
	itemPopularity    map[string]int                // item -> view count
	itemConversions   map[string]int                // item -> order count
	lastTrainingTime  time.Time
	minTrainingData   int
	similarityMetric  string // cosine, pearson, jaccard
}

// RecommendationConfig configures the recommendation engine
type RecommendationConfig struct {
	MinTrainingData  int
	SimilarityMetric string
	MaxRecommendations int
}

// ItemScore represents an item with its recommendation score
type ItemScore struct {
	ItemID string
	Score  float64
	Reason string
}

// NewRecommendationEngine creates a new recommendation engine
func NewRecommendationEngine(config RecommendationConfig) *RecommendationEngine {
	if config.MinTrainingData <= 0 {
		config.MinTrainingData = 10
	}
	if config.SimilarityMetric == "" {
		config.SimilarityMetric = "cosine"
	}
	
	return &RecommendationEngine{
		userItemMatrix:   make(map[string]map[string]float64),
		itemSimilarity:   make(map[string]map[string]float64),
		itemPopularity:   make(map[string]int),
		itemConversions:  make(map[string]int),
		minTrainingData:  config.MinTrainingData,
		similarityMetric: config.SimilarityMetric,
	}
}

// RecordInteraction records a user interaction with an item
func (re *RecommendationEngine) RecordInteraction(userID, itemID string, interactionType string, weight float64) {
	re.mu.Lock()
	defer re.mu.Unlock()
	
	// Initialize user if not exists
	if _, exists := re.userItemMatrix[userID]; !exists {
		re.userItemMatrix[userID] = make(map[string]float64)
	}
	
	// Update interaction score based on type
	switch interactionType {
	case "view":
		re.userItemMatrix[userID][itemID] += 1.0 * weight
		re.itemPopularity[itemID]++
	case "click":
		re.userItemMatrix[userID][itemID] += 2.0 * weight
	case "add_to_cart":
		re.userItemMatrix[userID][itemID] += 5.0 * weight
	case "order":
		re.userItemMatrix[userID][itemID] += 10.0 * weight
		re.itemConversions[itemID]++
	case "favorite":
		re.userItemMatrix[userID][itemID] += 8.0 * weight
	}
}

// Train computes item similarities based on user interactions
func (re *RecommendationEngine) Train() error {
	re.mu.Lock()
	defer re.mu.Unlock()
	
	// Check if we have enough data
	if len(re.userItemMatrix) < re.minTrainingData {
		return nil // Not enough data yet
	}
	
	// Build item-item similarity matrix
	items := re.getAllItems()
	re.itemSimilarity = make(map[string]map[string]float64)
	
	for i, item1 := range items {
		re.itemSimilarity[item1] = make(map[string]float64)
		for j := i + 1; j < len(items); j++ {
			item2 := items[j]
			
			similarity := re.calculateSimilarity(item1, item2)
			re.itemSimilarity[item1][item2] = similarity
			
			// Symmetric
			if _, exists := re.itemSimilarity[item2]; !exists {
				re.itemSimilarity[item2] = make(map[string]float64)
			}
			re.itemSimilarity[item2][item1] = similarity
		}
	}
	
	re.lastTrainingTime = time.Now()
	return nil
}

// GetRecommendations returns personalized recommendations for a user
func (re *RecommendationEngine) GetRecommendations(userID string, excludeItems []string, limit int) []ItemScore {
	re.mu.RLock()
	defer re.mu.RUnlock()
	
	if limit <= 0 {
		limit = 10
	}
	
	// Get user's interaction history
	userItems, exists := re.userItemMatrix[userID]
	if !exists || len(userItems) == 0 {
		// Cold start: return popular items
		return re.getPopularItems(excludeItems, limit)
	}
	
	// Calculate scores for all items
	scores := make(map[string]float64)
	reasons := make(map[string]string)
	
	// Collaborative filtering: find similar items
	for itemID, userScore := range userItems {
		if similarities, exists := re.itemSimilarity[itemID]; exists {
			for similarItemID, similarity := range similarities {
				if !contains(excludeItems, similarItemID) {
					scores[similarItemID] += similarity * userScore
					reasons[similarItemID] = "Similar to items you liked"
				}
			}
		}
	}
	
	// Convert to ItemScore slice
	results := make([]ItemScore, 0, len(scores))
	for itemID, score := range scores {
		if _, interacted := userItems[itemID]; !interacted {
			results = append(results, ItemScore{
				ItemID: itemID,
				Score:  score,
				Reason: reasons[itemID],
			})
		}
	}
	
	// Sort by score descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})
	
	// Limit results
	if len(results) > limit {
		results = results[:limit]
	}
	
	// If not enough collaborative filtering results, add popular items
	if len(results) < limit {
		popular := re.getPopularItems(append(excludeItems, getItemIDs(results)...), limit-len(results))
		results = append(results, popular...)
	}
	
	return results
}

// GetSimilarItems returns items similar to a given item
func (re *RecommendationEngine) GetSimilarItems(itemID string, limit int) []ItemScore {
	re.mu.RLock()
	defer re.mu.RUnlock()
	
	if limit <= 0 {
		limit = 5
	}
	
	similarities, exists := re.itemSimilarity[itemID]
	if !exists {
		return []ItemScore{}
	}
	
	// Convert to ItemScore slice
	results := make([]ItemScore, 0, len(similarities))
	for similarItemID, similarity := range similarities {
		results = append(results, ItemScore{
			ItemID: similarItemID,
			Score:  similarity,
			Reason: "Similar items",
		})
	}
	
	// Sort by similarity descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})
	
	// Limit results
	if len(results) > limit {
		results = results[:limit]
	}
	
	return results
}

// GetTrendingItems returns trending items based on recent interactions
func (re *RecommendationEngine) GetTrendingItems(timeWindow time.Duration, limit int) []ItemScore {
	re.mu.RLock()
	defer re.mu.RUnlock()
	
	// This is a simplified version - in production you'd track timestamps
	// For now, return items with high conversion rate
	
	results := make([]ItemScore, 0)
	for itemID, conversions := range re.itemConversions {
		views := re.itemPopularity[itemID]
		if views > 0 {
			conversionRate := float64(conversions) / float64(views)
			results = append(results, ItemScore{
				ItemID: itemID,
				Score:  conversionRate * float64(views), // Weighted by popularity
				Reason: "Trending now",
			})
		}
	}
	
	// Sort by score descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})
	
	// Limit results
	if len(results) > limit {
		results = results[:limit]
	}
	
	return results
}

// calculateSimilarity computes similarity between two items
func (re *RecommendationEngine) calculateSimilarity(item1, item2 string) float64 {
	switch re.similarityMetric {
	case "cosine":
		return re.cosineSimilarity(item1, item2)
	case "pearson":
		return re.pearsonCorrelation(item1, item2)
	case "jaccard":
		return re.jaccardSimilarity(item1, item2)
	default:
		return re.cosineSimilarity(item1, item2)
	}
}

// cosineSimilarity computes cosine similarity between two items
func (re *RecommendationEngine) cosineSimilarity(item1, item2 string) float64 {
	// Find users who interacted with both items
	vec1 := make(map[string]float64)
	vec2 := make(map[string]float64)
	
	for userID, items := range re.userItemMatrix {
		if score1, exists := items[item1]; exists {
			vec1[userID] = score1
		}
		if score2, exists := items[item2]; exists {
			vec2[userID] = score2
		}
	}
	
	// Calculate dot product and magnitudes
	var dotProduct, mag1, mag2 float64
	
	for userID, score1 := range vec1 {
		mag1 += score1 * score1
		if score2, exists := vec2[userID]; exists {
			dotProduct += score1 * score2
		}
	}
	
	for _, score2 := range vec2 {
		mag2 += score2 * score2
	}
	
	// Avoid division by zero
	if mag1 == 0 || mag2 == 0 {
		return 0
	}
	
	return dotProduct / (math.Sqrt(mag1) * math.Sqrt(mag2))
}

// pearsonCorrelation computes Pearson correlation coefficient
func (re *RecommendationEngine) pearsonCorrelation(item1, item2 string) float64 {
	// Find common users
	commonUsers := make([]string, 0)
	
	for userID, items := range re.userItemMatrix {
		if _, has1 := items[item1]; has1 {
			if _, has2 := items[item2]; has2 {
				commonUsers = append(commonUsers, userID)
			}
		}
	}
	
	if len(commonUsers) < 2 {
		return 0
	}
	
	// Calculate means
	var sum1, sum2 float64
	for _, userID := range commonUsers {
		sum1 += re.userItemMatrix[userID][item1]
		sum2 += re.userItemMatrix[userID][item2]
	}
	mean1 := sum1 / float64(len(commonUsers))
	mean2 := sum2 / float64(len(commonUsers))
	
	// Calculate correlation
	var numerator, denom1, denom2 float64
	for _, userID := range commonUsers {
		diff1 := re.userItemMatrix[userID][item1] - mean1
		diff2 := re.userItemMatrix[userID][item2] - mean2
		numerator += diff1 * diff2
		denom1 += diff1 * diff1
		denom2 += diff2 * diff2
	}
	
	if denom1 == 0 || denom2 == 0 {
		return 0
	}
	
	return numerator / (math.Sqrt(denom1) * math.Sqrt(denom2))
}

// jaccardSimilarity computes Jaccard similarity coefficient
func (re *RecommendationEngine) jaccardSimilarity(item1, item2 string) float64 {
	users1 := make(map[string]bool)
	users2 := make(map[string]bool)
	
	for userID, items := range re.userItemMatrix {
		if _, exists := items[item1]; exists {
			users1[userID] = true
		}
		if _, exists := items[item2]; exists {
			users2[userID] = true
		}
	}
	
	// Count intersection and union
	intersection := 0
	for userID := range users1 {
		if users2[userID] {
			intersection++
		}
	}
	
	union := len(users1) + len(users2) - intersection
	if union == 0 {
		return 0
	}
	
	return float64(intersection) / float64(union)
}

// getPopularItems returns most popular items
func (re *RecommendationEngine) getPopularItems(excludeItems []string, limit int) []ItemScore {
	results := make([]ItemScore, 0)
	
	for itemID, count := range re.itemPopularity {
		if !contains(excludeItems, itemID) {
			results = append(results, ItemScore{
				ItemID: itemID,
				Score:  float64(count),
				Reason: "Popular item",
			})
		}
	}
	
	// Sort by popularity descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})
	
	if len(results) > limit {
		results = results[:limit]
	}
	
	return results
}

// getAllItems returns all item IDs
func (re *RecommendationEngine) getAllItems() []string {
	itemSet := make(map[string]bool)
	
	for _, items := range re.userItemMatrix {
		for itemID := range items {
			itemSet[itemID] = true
		}
	}
	
	items := make([]string, 0, len(itemSet))
	for itemID := range itemSet {
		items = append(items, itemID)
	}
	
	return items
}

// Helper functions

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func getItemIDs(scores []ItemScore) []string {
	ids := make([]string, len(scores))
	for i, score := range scores {
		ids[i] = score.ItemID
	}
	return ids
}

// GetStats returns statistics about the recommendation engine
func (re *RecommendationEngine) GetStats() map[string]interface{} {
	re.mu.RLock()
	defer re.mu.RUnlock()
	
	totalInteractions := 0
	for _, items := range re.userItemMatrix {
		totalInteractions += len(items)
	}
	
	return map[string]interface{}{
		"total_users":        len(re.userItemMatrix),
		"total_items":        len(re.getAllItems()),
		"total_interactions": totalInteractions,
		"last_training":      re.lastTrainingTime,
		"similarity_metric":  re.similarityMetric,
	}
}
