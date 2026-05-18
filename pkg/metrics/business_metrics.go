package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Business Metrics

	// MenusCreated tracks total menus created per restaurant
	MenusCreated = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "qr_menu_menus_created_total",
			Help: "Total number of menus created",
		},
		[]string{"restaurant_id"},
	)

	// ItemsAdded tracks total menu items added
	ItemsAdded = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "qr_menu_items_added_total",
			Help: "Total number of menu items added",
		},
		[]string{"restaurant_id"},
	)

	// QRCodesGenerated tracks QR code generation
	QRCodesGenerated = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "qr_menu_qrcodes_generated_total",
			Help: "Total number of QR codes generated",
		},
		[]string{"restaurant_id"},
	)

	// MenuViews tracks number of public menu views
	MenuViews = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "qr_menu_views_total",
			Help: "Total number of menu views",
		},
		[]string{"restaurant_id", "menu_id"},
	)

	// RestaurantsActive tracks active restaurants
	RestaurantsActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "qr_menu_restaurants_active",
			Help: "Number of active restaurants",
		},
	)

	// UsersActive tracks active users
	UsersActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "qr_menu_users_active",
			Help: "Number of active users",
		},
	)

	// DatabaseConnectionPoolSize tracks connection pool
	DatabaseConnectionPoolSize = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "qr_menu_db_connection_pool_size",
			Help: "Current size of database connection pool",
		},
	)

	// CacheHitRate tracks cache effectiveness
	CacheHitRate = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "qr_menu_cache_hits_total",
			Help: "Total number of cache hits",
		},
		[]string{"cache_type"},
	)

	// CacheMissRate tracks cache misses
	CacheMissRate = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "qr_menu_cache_misses_total",
			Help: "Total number of cache misses",
		},
		[]string{"cache_type"},
	)

	// OperationDuration tracks operation latencies by type
	OperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "qr_menu_operation_duration_seconds",
			Help:    "Duration of business operations in seconds",
			Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5},
		},
		[]string{"operation"},
	)

	// DataSize tracks data volume in system
	DataSize = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "qr_menu_data_size_bytes",
			Help: "Size of stored data in bytes",
		},
		[]string{"collection"},
	)

	// CircuitBreakerState tracks circuit breaker status
	CircuitBreakerState = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "qr_menu_circuit_breaker_state",
			Help: "Circuit breaker state (0=closed, 1=half-open, 2=open)",
		},
		[]string{"service"},
	)
)
