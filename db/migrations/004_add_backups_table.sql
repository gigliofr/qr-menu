-- Create backups table
CREATE TABLE IF NOT EXISTS backups (
	id VARCHAR(36) PRIMARY KEY,
	restaurant_id VARCHAR(36),
	backup_path VARCHAR(255) NOT NULL,
	backup_size BIGINT NOT NULL,
	file_count INT,
	compress_rate DECIMAL(5, 2),
	hash VARCHAR(64),
	status VARCHAR(50) DEFAULT 'success',
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	INDEX idx_created_at (created_at),
	INDEX idx_status (status)
);

-- Create backup_schedules table
CREATE TABLE IF NOT EXISTS backup_schedules (
	id VARCHAR(36) PRIMARY KEY,
	restaurant_id VARCHAR(36),
	schedule_type VARCHAR(50) NOT NULL,
	schedule_hour INT,
	schedule_day INT,
	is_active BOOLEAN DEFAULT TRUE,
	next_run TIMESTAMP,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	INDEX idx_restaurant_id (restaurant_id)
);