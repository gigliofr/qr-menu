package db

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

// AuditLog rappresenta un log di audit
type AuditLog struct {
	ID            string                 `bson:"_id,omitempty"`
	Action        string                 `bson:"action"`
	ResourceType  string                 `bson:"resource_type"`
	ResourceID    string                 `bson:"resource_id"`
	RestaurantID  string                 `bson:"restaurant_id"`
	UserID        string                 `bson:"user_id"`
	IPAddress     string                 `bson:"ip_address"`
	UserAgent     string                 `bson:"user_agent"`
	Status        string                 `bson:"status"` // success, failure, warning
	OldValue      map[string]interface{} `bson:"old_value,omitempty"`
	NewValue      map[string]interface{} `bson:"new_value,omitempty"`
	ErrorMessage  string                 `bson:"error_message,omitempty"`
	Timestamp     time.Time              `bson:"timestamp"`
	DurationMs    int64                  `bson:"duration_ms,omitempty"`
}

// AnalyticsEvent rappresenta un evento di analytics
type AnalyticsEvent struct {
	ID           string                 `bson:"_id,omitempty"`
	EventType    string                 `bson:"event_type"`
	RestaurantID string                 `bson:"restaurant_id"`
	UserID       string                 `bson:"user_id,omitempty"`
	SessionID    string                 `bson:"session_id,omitempty"`
	MenuID       string                 `bson:"menu_id,omitempty"`
	Data         map[string]interface{} `bson:"data"`
	IPAddress    string                 `bson:"ip_address"`
	UserAgent    string                 `bson:"user_agent"`
	Timestamp    time.Time              `bson:"timestamp"`
	DayDate      string                 `bson:"day_date"` // YYYY-MM-DD per aggregazioni
}

// ==================== AUDIT LOGS ====================

// CreateAuditLog crea un nuovo log di audit
func (m *MongoClient) CreateAuditLog(ctx context.Context, log *AuditLog) error {
	if log.Timestamp.IsZero() {
		log.Timestamp = time.Now()
	}

	coll := m.db.Collection("audit_logs")
	_, err := coll.InsertOne(ctx, log)
	if err != nil {
		return err
	}
	return nil
}

// GetAuditLogs recupera i log di audit con filtri
func (m *MongoClient) GetAuditLogs(ctx context.Context, restaurantID string, limit int64) ([]*AuditLog, error) {
	coll := m.db.Collection("audit_logs")

	opts := options.Find().
		SetSort(bson.M{"timestamp": -1}).
		SetLimit(limit)

	cursor, err := coll.Find(ctx, bson.M{"restaurant_id": restaurantID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var logs []*AuditLog
	if err = cursor.All(ctx, &logs); err != nil {
		return nil, err
	}
	return logs, nil
}

// GetAuditLogsByAction filtra i log per azione
func (m *MongoClient) GetAuditLogsByAction(ctx context.Context, restaurantID, action string, limit int64) ([]*AuditLog, error) {
	coll := m.db.Collection("audit_logs")

	opts := options.Find().
		SetSort(bson.M{"timestamp": -1}).
		SetLimit(limit)

	cursor, err := coll.Find(ctx, bson.M{
		"restaurant_id": restaurantID,
		"action":        action,
	}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var logs []*AuditLog
	if err = cursor.All(ctx, &logs); err != nil {
		return nil, err
	}
	return logs, nil
}

// GetAuditLogsByDateRange recupera i log in un intervallo di date
func (m *MongoClient) GetAuditLogsByDateRange(ctx context.Context, restaurantID string, startDate, endDate time.Time) ([]*AuditLog, error) {
	coll := m.db.Collection("audit_logs")

	cursor, err := coll.Find(ctx, bson.M{
		"restaurant_id": restaurantID,
		"timestamp": bson.M{
			"$gte": startDate,
			"$lte": endDate,
		},
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var logs []*AuditLog
	if err = cursor.All(ctx, &logs); err != nil {
		return nil, err
	}
	return logs, nil
}

// ==================== ANALYTICS ====================

// CreateAnalyticsEvent crea un nuovo evento di analytics
func (m *MongoClient) CreateAnalyticsEvent(ctx context.Context, event *AnalyticsEvent) error {
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}
	if event.DayDate == "" {
		event.DayDate = event.Timestamp.Format("2006-01-02")
	}

	coll := m.db.Collection("analytics_events")
	_, err := coll.InsertOne(ctx, event)
	if err != nil {
		return err
	}
	return nil
}

// GetAnalyticsEvents recupera gli eventi di analytics
func (m *MongoClient) GetAnalyticsEvents(ctx context.Context, restaurantID string, limit int64) ([]*AnalyticsEvent, error) {
	coll := m.db.Collection("analytics_events")

	opts := options.Find().
		SetSort(bson.M{"timestamp": -1}).
		SetLimit(limit)

	cursor, err := coll.Find(ctx, bson.M{"restaurant_id": restaurantID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var events []*AnalyticsEvent
	if err = cursor.All(ctx, &events); err != nil {
		return nil, err
	}
	return events, nil
}

// GetAnalyticsEventsByType filtra gli eventi per tipo
func (m *MongoClient) GetAnalyticsEventsByType(ctx context.Context, restaurantID, eventType string, limit int64) ([]*AnalyticsEvent, error) {
	coll := m.db.Collection("analytics_events")

	opts := options.Find().
		SetSort(bson.M{"timestamp": -1}).
		SetLimit(limit)

	cursor, err := coll.Find(ctx, bson.M{
		"restaurant_id": restaurantID,
		"event_type":    eventType,
	}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var events []*AnalyticsEvent
	if err = cursor.All(ctx, &events); err != nil {
		return nil, err
	}
	return events, nil
}

// GetAnalyticsEventsByDate recupera gli eventi di una data specifica
func (m *MongoClient) GetAnalyticsEventsByDate(ctx context.Context, restaurantID, dateStr string) ([]*AnalyticsEvent, error) {
	coll := m.db.Collection("analytics_events")

	cursor, err := coll.Find(ctx, bson.M{
		"restaurant_id": restaurantID,
		"day_date":      dateStr,
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var events []*AnalyticsEvent
	if err = cursor.All(ctx, &events); err != nil {
		return nil, err
	}
	return events, nil
}

// GetAnalyticsSummary restituisce un summary degli analytics
func (m *MongoClient) GetAnalyticsSummary(ctx context.Context, restaurantID string, days int) (map[string]interface{}, error) {
	coll := m.db.Collection("analytics_events")

	startDate := time.Now().AddDate(0, 0, -days)

	pipeline := []bson.M{
		{
			"$match": bson.M{
				"restaurant_id": restaurantID,
				"timestamp": bson.M{
					"$gte": startDate,
				},
			},
		},
		{
			"$group": bson.M{
				"_id":   "$event_type",
				"count": bson.M{"$sum": 1},
			},
		},
	}

	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	summary := make(map[string]interface{})
	summary["period_days"] = days
	summary["period_start"] = startDate
	summary["period_end"] = time.Now()
	summary["events_by_type"] = results

	return summary, nil
}

// ==================== UTILITY ====================

// DropCollections elimina tutte le collections (ATTENZIONE: data loss!)
func (m *MongoClient) DropCollections(ctx context.Context) error {
	collections := []string{"restaurants", "menus", "sessions", "audit_logs", "analytics_events"}
	for _, collName := range collections {
		if err := m.db.Collection(collName).Drop(ctx); err != nil {
			return err
		}
	}
	return nil
}

// GetDatabaseStats restituisce statistiche del database
func (m *MongoClient) GetDatabaseStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Count restaurants
	restColl := m.db.Collection("restaurants")
	restCount, _ := restColl.EstimatedDocumentCount(ctx)
	stats["restaurants"] = restCount

	// Count menus
	menuColl := m.db.Collection("menus")
	menuCount, _ := menuColl.EstimatedDocumentCount(ctx)
	stats["menus"] = menuCount

	// Count sessions
	sessColl := m.db.Collection("sessions")
	sessCount, _ := sessColl.EstimatedDocumentCount(ctx)
	stats["sessions"] = sessCount

	// Count audit logs
	auditColl := m.db.Collection("audit_logs")
	auditCount, _ := auditColl.EstimatedDocumentCount(ctx)
	stats["audit_logs"] = auditCount

	// Count analytics events
	analyticsColl := m.db.Collection("analytics_events")
	analyticsCount, _ := analyticsColl.EstimatedDocumentCount(ctx)
	stats["analytics_events"] = analyticsCount

	stats["timestamp"] = time.Now()

	return stats, nil
}
