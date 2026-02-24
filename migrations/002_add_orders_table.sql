-- Create orders table
CREATE TABLE IF NOT EXISTS orders (
	id VARCHAR(36) PRIMARY KEY,
	restaurant_id VARCHAR(36) NOT NULL,
	customer_name VARCHAR(255),
	customer_phone VARCHAR(20),
	customer_email VARCHAR(255),
	total_amount DECIMAL(10, 2) NOT NULL,
	status VARCHAR(50) NOT NULL DEFAULT 'pending',
	notes TEXT,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (restaurant_id) REFERENCES restaurants(id) ON DELETE CASCADE,
	INDEX idx_restaurant_id (restaurant_id),
	INDEX idx_status (status),
	INDEX idx_created_at (created_at)
);

-- Create order_items table
CREATE TABLE IF NOT EXISTS order_items (
	id VARCHAR(36) PRIMARY KEY,
	order_id VARCHAR(36) NOT NULL,
	menu_item_id VARCHAR(36),
	item_name VARCHAR(255) NOT NULL,
	quantity INT NOT NULL,
	unit_price DECIMAL(10, 2) NOT NULL,
	total_price DECIMAL(10, 2) NOT NULL,
	FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE,
	FOREIGN KEY (menu_item_id) REFERENCES menu_items(id),
	INDEX idx_order_id (order_id)
);