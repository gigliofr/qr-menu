-- Create notification_preferences table
CREATE TABLE IF NOT EXISTS notification_preferences (
	id VARCHAR(36) PRIMARY KEY,
	restaurant_id VARCHAR(36) NOT NULL,
	enable_push BOOLEAN DEFAULT TRUE,
	enable_email BOOLEAN DEFAULT TRUE,
	enable_sms BOOLEAN DEFAULT FALSE,
	order_notifications BOOLEAN DEFAULT TRUE,
	reservation_notifications BOOLEAN DEFAULT TRUE,
	promo_notifications BOOLEAN DEFAULT TRUE,
	quiet_hours_start VARCHAR(5),
	quiet_hours_end VARCHAR(5),
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (restaurant_id) REFERENCES restaurants(id) ON DELETE CASCADE,
	UNIQUE KEY uniq_restaurant_id (restaurant_id)
);

-- Create notification_history table
CREATE TABLE IF NOT EXISTS notification_history (
	id VARCHAR(36) PRIMARY KEY,
	restaurant_id VARCHAR(36) NOT NULL,
	type VARCHAR(50) NOT NULL,
	title VARCHAR(255) NOT NULL,
	body TEXT,
	status VARCHAR(50) DEFAULT 'sent',
	read_at TIMESTAMP,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (restaurant_id) REFERENCES restaurants(id) ON DELETE CASCADE,
	INDEX idx_restaurant_id (restaurant_id),
	INDEX idx_created_at (created_at)
);