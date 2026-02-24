-- Create analytics_events table
CREATE TABLE IF NOT EXISTS analytics_events (
	id VARCHAR(36) PRIMARY KEY,
	restaurant_id VARCHAR(36) NOT NULL,
	event_type VARCHAR(100) NOT NULL,
	event_data JSON,
	user_id VARCHAR(36),
	session_id VARCHAR(36),
	ip_address VARCHAR(45),
	user_agent TEXT,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (restaurant_id) REFERENCES restaurants(id) ON DELETE CASCADE,
	INDEX idx_restaurant_id (restaurant_id),
	INDEX idx_event_type (event_type),
	INDEX idx_created_at (created_at)
);

-- Create analytics_sessions table
CREATE TABLE IF NOT EXISTS analytics_sessions (
	id VARCHAR(36) PRIMARY KEY,
	restaurant_id VARCHAR(36) NOT NULL,
	session_start TIMESTAMP NOT NULL,
	session_end TIMESTAMP,
	duration_seconds INT,
	total_events INT DEFAULT 0,
	FOREIGN KEY (restaurant_id) REFERENCES restaurants(id) ON DELETE CASCADE,
	INDEX idx_restaurant_id (restaurant_id),
	INDEX idx_session_start (session_start)
);