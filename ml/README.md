# Machine Learning & Analytics Module

## Overview

This module provides advanced machine learning and analytics capabilities for the QR Menu System, including personalized recommendations, predictive analytics, and A/B testing framework.

## Components

### 1. Recommendation Engine (`recommendations.go`)

Collaborative filtering-based recommendation system using item-item similarity.

**Features:**
- Multiple similarity metrics (cosine, Pearson, Jaccard)
- Collaborative filtering (item-based)
- Cold start handling with popularity-based recommendations
- Trending items detection
- Real-time interaction tracking

**Similarity Metrics:**
- **Cosine Similarity**: Measures angle between user rating vectors
- **Pearson Correlation**: Measures linear correlation between ratings
- **Jaccard Similarity**: Measures overlap between user sets

**Usage:**
```go
// Initialize
config := ml.RecommendationConfig{
    MinTrainingData:    20,
    SimilarityMetric:   "cosine",
    MaxRecommendations: 10,
}
engine := ml.NewRecommendationEngine(config)

// Track interactions
engine.RecordInteraction("user123", "item456", "view", 1.0)
engine.RecordInteraction("user123", "item789", "order", 1.0)

// Train model
engine.Train()

// Get recommendations
recommendations := engine.GetRecommendations("user123", []string{}, 10)
for _, rec := range recommendations {
    fmt.Printf("Item: %s, Score: %.2f, Reason: %s\n", 
        rec.ItemID, rec.Score, rec.Reason)
}
```

**Interaction Types & Weights:**
- `view`: 1.0
- `click`: 2.0
- `add_to_cart`: 5.0
- `favorite`: 8.0
- `order`: 10.0

**Cold Start Strategy:**
When a user has no interaction history, the system falls back to popular items ranked by:
- View count
- Conversion rate
- Recent trends

### 2. Predictive Analytics (`predictions.go`)

Time series forecasting and trend analysis using statistical methods.

**Features:**
- Holt-Winters exponential smoothing for forecasting
- Seasonal pattern detection
- Trend analysis (linear regression)
- Peak time prediction
- Inventory optimization

**Forecasting Methods:**
- **Holt-Winters**: Captures level and trend
- **Seasonality Detection**: Identifies daily/weekly patterns
- **Confidence Intervals**: 95% confidence bounds

**Usage:**
```go
// Initialize
pa := ml.NewPredictiveAnalytics()

// Add historical data
for _, dataPoint := range historicalData {
    point := ml.TimeSeriesPoint{
        Timestamp: dataPoint.Time,
        Value:     dataPoint.Value,
        Metadata:  dataPoint.Meta,
    }
    pa.AddDataPoint("orders", point)
}

// Forecast demand
forecasts, _ := pa.ForecastDemand("orders", 7) // 7 periods ahead
for _, f := range forecasts {
    fmt.Printf("Time: %s, Predicted: %.2f (%.2f - %.2f)\n",
        f.Timestamp, f.PredictedValue, f.ConfidenceLow, f.ConfidenceHigh)
}

// Detect seasonality
pattern := pa.DetectSeasonality("orders")
if pattern.Detected {
    fmt.Printf("Seasonal period: %s, Amplitude: %.2f\n",
        pattern.Period, pattern.Amplitude)
}

// Analyze trend
trend := pa.AnalyzeTrend("revenue")
fmt.Printf("Direction: %s, Slope: %.4f, R²: %.4f\n",
    trend.Direction, trend.Slope, trend.R2)
```

**Inventory Optimization:**
```go
// Optimize inventory for an item
optimization := pa.OptimizeInventory("item123", 7*24*time.Hour) // 7-day lead time

fmt.Printf("Expected Demand: %.0f units\n", optimization["expected_demand"])
fmt.Printf("Safety Stock: %.0f units\n", optimization["safety_stock"])
fmt.Printf("Recommended Stock: %.0f units\n", optimization["recommended_stock"])
fmt.Printf("Service Level: %.0f%%\n", optimization["service_level"]*100)
```

### 3. A/B Testing Framework (`abtesting.go`)

Complete A/B testing system with statistical significance calculation.

**Features:**
- Multi-variant testing (A/B/n)
- Traffic allocation control
- Conversion tracking
- Statistical significance testing (Z-test)
- Experiment lifecycle management

**Experiment Lifecycle:**
1. **Draft**: Initial creation
2. **Running**: Active experiment
3. **Paused**: Temporarily stopped
4. **Completed**: Ended with results

**Usage:**
```go
// Initialize
abTest := ml.NewABTestManager()

// Create experiment
experiment := ml.Experiment{
    Name:        "Menu Layout Test",
    Description: "Test two menu layouts",
    Metric:      "conversion_rate",
    Variants: []ml.Variant{
        {
            ID:          "control",
            Name:        "Original Layout",
            Traffic:     0.5,
            IsControl:   true,
            Config:      map[string]interface{}{"layout": "grid"},
        },
        {
            ID:          "variant_a",
            Name:        "List Layout",
            Traffic:     0.5,
            IsControl:   false,
            Config:      map[string]interface{}{"layout": "list"},
        },
    },
}
exp, _ := abTest.CreateExperiment(experiment)

// Start experiment
abTest.StartExperiment(exp.ID)

// Assign user to variant
variantID, _ := abTest.AssignVariant(exp.ID, "user123")

// Track conversion
event := ml.ConversionEvent{
    UserID:       "user123",
    ExperimentID: exp.ID,
    EventType:    "purchase",
    Value:        25.99,
    Timestamp:    time.Now(),
}
abTest.TrackConversion(event)

// Get results
results := abTest.GetExperimentResults(exp.ID)
if results.StatSignificant {
    fmt.Printf("Winner: %s (p-value: %.4f)\n", results.Winner, results.PValue)
    fmt.Printf("Recommendation: %s\n", results.Recommendation)
} else {
    fmt.Println("Not enough data for statistical significance")
}
```

**Statistical Significance:**
- Uses Z-test for proportions
- p-value < 0.05 for significance
- Minimum 30 samples per variant
- Two-tailed test

## API Endpoints

### Recommendation Endpoints

#### `GET /api/v1/ml/recommendations`
Get personalized recommendations.

**Query Params:**
- `limit`: Max results (default: 10)
- `exclude`: Item IDs to exclude

**Response:**
```json
[
  {
    "item_id": "item123",
    "score": 0.87,
    "reason": "Similar to items you liked"
  }
]
```

#### `GET /api/v1/ml/items/{id}/similar`
Get items similar to a specific item.

**Path Params:**
- `id`: Item ID

**Query Params:**
- `limit`: Max results (default: 5)

#### `GET /api/v1/ml/items/trending`
Get trending items.

**Query Params:**
- `window`: Time window in hours (default: 24)
- `limit`: Max results (default: 10)

#### `POST /api/v1/ml/interactions`
Track user interaction.

**Body:**
```json
{
  "user_id": "user123",
  "item_id": "item456",
  "interaction_type": "view",
  "weight": 1.0
}
```

#### `POST /api/v1/ml/recommendations/train`
Trigger model training.

**Response:**
```json
{
  "status": "trained",
  "stats": {
    "total_users": 150,
    "total_items": 320,
    "total_interactions": 4200
  }
}
```

### Predictive Analytics Endpoints

#### `GET /api/v1/ml/forecast`
Forecast demand.

**Query Params:**
- `metric`: Metric to forecast (default: "orders")
- `periods`: Number of periods ahead (default: 7)

**Response:**
```json
[
  {
    "timestamp": "2026-02-25T00:00:00Z",
    "predicted_value": 125.5,
    "confidence_low": 110.2,
    "confidence_high": 140.8,
    "method": "holt-winters"
  }
]
```

#### `GET /api/v1/ml/seasonality`
Detect seasonal patterns.

**Query Params:**
- `metric`: Metric to analyze (default: "orders")

**Response:**
```json
{
  "detected": true,
  "period": "24h0m0s",
  "amplitude": 0.65,
  "baseline_value": 100.5,
  "trend_slope": 0.25
}
```

#### `GET /api/v1/ml/trend`
Analyze trend.

**Query Params:**
- `metric`: Metric to analyze (default: "revenue")

**Response:**
```json
{
  "direction": "up",
  "slope": 0.045,
  "r2": 0.82,
  "start_date": "2026-01-01T00:00:00Z",
  "end_date": "2026-02-24T00:00:00Z"
}
```

#### `GET /api/v1/ml/peak-times`
Predict peak demand times.

**Query Params:**
- `metric`: Metric to analyze (default: "orders")
- `lookahead_hours`: Hours to look ahead (default: 168)

**Response:**
```json
{
  "peak_times": [
    "2026-02-25T12:00:00Z",
    "2026-02-25T19:00:00Z",
    "2026-02-26T12:00:00Z"
  ],
  "metric": "orders",
  "lookahead": "168h0m0s"
}
```

#### `GET /api/v1/ml/inventory/{item_id}/optimize`
Optimize inventory for an item.

**Path Params:**
- `item_id`: Item ID

**Query Params:**
- `lead_time_days`: Lead time in days (default: 7)

**Response:**
```json
{
  "expected_demand": 42.0,
  "safety_stock": 12.5,
  "recommended_stock": 54.5,
  "service_level": 0.95
}
```

#### `POST /api/v1/ml/data-points`
Add data point to time series.

**Body:**
```json
{
  "metric": "orders",
  "value": 125.0,
  "timestamp": "2026-02-24T12:00:00Z",
  "metadata": {
    "restaurant_id": "rest123"
  }
}
```

### A/B Testing Endpoints

#### `POST /api/v1/ml/experiments`
Create A/B test experiment.

**Body:**
```json
{
  "name": "Menu Layout Test",
  "description": "Test two menu layouts",
  "metric": "conversion_rate",
  "variants": [
    {
      "id": "control",
      "name": "Original",
      "traffic": 0.5,
      "is_control": true,
      "config": {"layout": "grid"}
    },
    {
      "id": "variant_a",
      "name": "New Layout",
      "traffic": 0.5,
      "is_control": false,
      "config": {"layout": "list"}
    }
  ]
}
```

#### `GET /api/v1/ml/experiments`
List all experiments.

#### `POST /api/v1/ml/experiments/{id}/start`
Start an experiment.

#### `POST /api/v1/ml/experiments/{id}/stop`
Stop an experiment.

#### `GET /api/v1/ml/experiments/{id}/results`
Get experiment results.

**Response:**
```json
{
  "experiment_id": "exp123",
  "winner": "variant_a",
  "winner_variant": {
    "variant_id": "variant_a",
    "impressions": 1250,
    "conversions": 156,
    "conversion_rate": 0.1248,
    "revenue": 3850.75,
    "average_revenue": 24.68
  },
  "stat_significant": true,
  "p_value": 0.023,
  "confidence_level": 0.977,
  "recommendation": "variant_a performs better"
}
```

#### `POST /api/v1/ml/experiments/{id}/assign`
Assign user to variant.

**Body:**
```json
{
  "user_id": "user123"
}
```

**Response:**
```json
{
  "variant_id": "variant_a",
  "variant": {
    "id": "variant_a",
    "name": "New Layout",
    "config": {"layout": "list"}
  }
}
```

#### `POST /api/v1/ml/experiments/conversions`
Track conversion.

**Body:**
```json
{
  "user_id": "user123",
  "experiment_id": "exp123",
  "event_type": "purchase",
  "value": 25.99,
  "timestamp": "2026-02-24T12:00:00Z"
}
```

#### `GET /api/v1/ml/stats`
Get overall ML statistics.

**Response:**
```json
{
  "recommendations": {
    "total_users": 150,
    "total_items": 320,
    "total_interactions": 4200
  },
  "predictive": {
    "orders": {
      "data_points": 500,
      "latest_value": 125.0,
      "trend": "up",
      "trend_slope": 0.25
    }
  },
  "ab_testing": {
    "total_experiments": 5,
    "running_experiments": 2,
    "completed_experiments": 3
  }
}
```

## Use Cases

### 1. Personalized Menu Recommendations

```javascript
// Frontend: Get personalized recommendations
const recommendations = await fetch('/api/v1/ml/recommendations?limit=5')
  .then(r => r.json());

// Display recommended items
recommendations.forEach(rec => {
  console.log(`Recommend: ${rec.item_id} (${rec.reason})`);
});

// Track when user views an item
await fetch('/api/v1/ml/interactions', {
  method: 'POST',
  body: JSON.stringify({
    user_id: currentUser.id,
    item_id: item.id,
    interaction_type: 'view'
  })
});
```

### 2. Demand Forecasting for Inventory

```javascript
// Get 7-day demand forecast
const forecast = await fetch('/api/v1/ml/forecast?metric=orders&periods=7')
  .then(r => r.json());

// Display forecast chart
forecast.forEach(day => {
  console.log(`${day.timestamp}: ${day.predicted_value} orders`);
});

// Optimize inventory
const optimization = await fetch('/api/v1/ml/inventory/item123/optimize?lead_time_days=5')
  .then(r => r.json());

console.log(`Order ${optimization.recommended_stock} units`);
```

### 3. A/B Testing Menu Layouts

```javascript
// Get variant for current user
const assignment = await fetch('/api/v1/ml/experiments/exp123/assign', {
  method: 'POST',
  body: JSON.stringify({ user_id: currentUser.id })
}).then(r => r.json());

// Apply variant configuration
if (assignment.variant.config.layout === 'list') {
  renderListLayout();
} else {
  renderGridLayout();
}

// Track conversion
await fetch('/api/v1/ml/experiments/conversions', {
  method: 'POST',
  body: JSON.stringify({
    user_id: currentUser.id,
    experiment_id: 'exp123',
    event_type: 'purchase',
    value: orderTotal
  })
});
```

## Performance Considerations

- **Recommendation Engine**: O(n²) training complexity, run training periodically (not on every request)
- **Predictive Analytics**: Lightweight forecasting, suitable for real-time queries
- **A/B Testing**: Constant-time variant assignment with hash-based bucketing
- **Memory**: In-memory storage for demo; use Redis/database for production

## Best Practices

1. **Training Frequency**: Train recommendation model daily or when significant data changes
2. **Data Quality**: Ensure clean, validated time series data for accurate forecasts
3. **Sample Size**: Wait for 100+ samples per variant before declaring A/B test winner
4. **Confidence**: Use 95% confidence level (p < 0.05) for A/B testing
5. **Cold Start**: Always have fallback recommendations (popular items)
6. **Monitoring**: Track model performance and retrain when accuracy drops

## Production Deployment

1. **Database Integration**: Replace in-memory storage with persistent database
2. **Batch Training**: Schedule periodic training jobs for recommendation engine
3. **Caching**: Cache recommendations and forecasts for frequently requested items
4. **Monitoring**: Track model accuracy, latency, and resource usage
5. **A/B Testing**: Implement proper randomization and avoid selection bias
6. **Data Pipeline**: Set up automated data ingestion for real-time analytics

## License

Part of the QR Menu System - Enterprise Edition
