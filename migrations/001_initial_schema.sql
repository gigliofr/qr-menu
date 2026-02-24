-- Create restaurants table
CREATE TABLE IF NOT EXISTS restaurants (
	id VARCHAR(36) PRIMARY KEY,
	username VARCHAR(255) UNIQUE NOT NULL,
	email VARCHAR(255) UNIQUE NOT NULL,
	password_hash VARCHAR(255) NOT NULL,
	restaurant_name VARCHAR(255) NOT NULL,
	restaurant_type VARCHAR(100),
	phone VARCHAR(20),
	address TEXT,
	city VARCHAR(100),
	postal_code VARCHAR(20),
	country VARCHAR(100),
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	INDEX idx_username (username),
	INDEX idx_email (email)
);

-- Create menus table
CREATE TABLE IF NOT EXISTS menus (
	id VARCHAR(36) PRIMARY KEY,
	restaurant_id VARCHAR(36) NOT NULL,
	menu_name VARCHAR(255) NOT NULL,
	menu_description TEXT,
	is_active BOOLEAN DEFAULT FALSE,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (restaurant_id) REFERENCES restaurants(id) ON DELETE CASCADE,
	INDEX idx_restaurant_id (restaurant_id),
	INDEX idx_is_active (is_active)
);

-- Create categories table
CREATE TABLE IF NOT EXISTS categories (
	id VARCHAR(36) PRIMARY KEY,
	menu_id VARCHAR(36) NOT NULL,
	category_name VARCHAR(255) NOT NULL,
	category_description TEXT,
	display_order INT DEFAULT 0,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (menu_id) REFERENCES menus(id) ON DELETE CASCADE,
	INDEX idx_menu_id (menu_id)
);

-- Create menu_items table
CREATE TABLE IF NOT EXISTS menu_items (
	id VARCHAR(36) PRIMARY KEY,
	category_id VARCHAR(36) NOT NULL,
	item_name VARCHAR(255) NOT NULL,
	item_description TEXT,
	price DECIMAL(10, 2) NOT NULL,
	image_url VARCHAR(255),
	is_available BOOLEAN DEFAULT TRUE,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE,
	INDEX idx_category_id (category_id),
	INDEX idx_is_available (is_available)
);