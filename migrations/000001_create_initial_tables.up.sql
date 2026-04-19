BEGIN;

-- Reset all legacy tables so schema starts clean from this migration.
DROP TABLE IF EXISTS locations CASCADE;
DROP TABLE IF EXISTS locations_legacy CASCADE;
DROP TABLE IF EXISTS trips CASCADE;
DROP TABLE IF EXISTS orders CASCADE;
DROP TABLE IF EXISTS drivers CASCADE;
DROP TABLE IF EXISTS trucks CASCADE;
DROP TABLE IF EXISTS customers CASCADE;
DROP TABLE IF EXISTS audit_logs CASCADE;
DROP TABLE IF EXISTS auth_tokens CASCADE;
DROP TABLE IF EXISTS users CASCADE;

-- 1. Users (Admin/Staff internal only)
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    full_name VARCHAR(100),
    role VARCHAR(20) NOT NULL CHECK (role IN ('super_admin', 'admin_sales', 'admin_ops')),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 2. Customers (business entities, no login)
CREATE TABLE customers (
    id SERIAL PRIMARY KEY,
    company_name VARCHAR(150) NOT NULL,
    pic_name VARCHAR(100),
    phone VARCHAR(20),
    email VARCHAR(100),
    address TEXT,
    npwp VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 3. Drivers
CREATE TABLE drivers (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    license_number VARCHAR(50) UNIQUE,
    phone VARCHAR(20),
    status VARCHAR(20) DEFAULT 'available' CHECK (status IN ('available', 'on_duty', 'off')),
    is_active BOOLEAN DEFAULT true
);

-- 4. Trucks
CREATE TABLE trucks (
    id SERIAL PRIMARY KEY,
    plate_number VARCHAR(20) UNIQUE NOT NULL,
    truck_type VARCHAR(50),
    status VARCHAR(20) DEFAULT 'available' CHECK (status IN ('available', 'on_duty', 'maintenance')),
    is_active BOOLEAN DEFAULT true
);

-- 5. Orders (global customer order)
CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    order_number VARCHAR(50) UNIQUE NOT NULL,
    customer_id INT REFERENCES customers(id),
    admin_id INT REFERENCES users(id),
    origin TEXT NOT NULL,
    destination TEXT NOT NULL,
    total_containers INT DEFAULT 1,
    order_date TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'partial', 'completed', 'cancelled'))
);

-- 6. Trips / Surat Jalan (operational execution per truck)
CREATE TABLE trips (
    id SERIAL PRIMARY KEY,
    order_id INT REFERENCES orders(id),
    truck_id INT REFERENCES trucks(id),
    driver_id INT REFERENCES drivers(id),
    trip_number VARCHAR(50) UNIQUE NOT NULL,
    container_number VARCHAR(50),
    seal_number VARCHAR(50),
    status VARCHAR(20) DEFAULT 'pickup' CHECK (status IN ('pickup', 'in_transit', 'delivered', 'cancelled')),
    start_time TIMESTAMP WITH TIME ZONE,
    end_time TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 7. Locations (GPS tracking per trip) with partitioning.
CREATE TABLE locations (
    id BIGSERIAL,
    trip_id INT REFERENCES trips(id),
    latitude DOUBLE PRECISION NOT NULL,
    longitude DOUBLE PRECISION NOT NULL,
    speed FLOAT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
) PARTITION BY RANGE (created_at);

-- Default partition keeps writes safe before monthly partitions are created.
CREATE TABLE locations_default PARTITION OF locations DEFAULT;

-- 8. Audit logs for compliance and traceability.
CREATE TABLE audit_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id),
    action VARCHAR(20) NOT NULL CHECK (action IN ('CREATE', 'UPDATE', 'DELETE')),
    table_name VARCHAR(50) NOT NULL,
    record_id INT NOT NULL,
    old_values TEXT,
    new_values TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_customers_company_name ON customers(company_name);
CREATE INDEX idx_drivers_status ON drivers(status);
CREATE INDEX idx_trucks_plate_number ON trucks(plate_number);
CREATE INDEX idx_trips_order_id ON trips(order_id);
CREATE INDEX idx_trips_truck_id ON trips(truck_id);
CREATE INDEX idx_trips_driver_id ON trips(driver_id);
CREATE INDEX idx_locations_trip_id_created_at ON locations(trip_id, created_at DESC);
CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_table_record ON audit_logs(table_name, record_id);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at DESC);

COMMIT;