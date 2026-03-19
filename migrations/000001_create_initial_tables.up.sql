-- 1. Tabel Users (Untuk Admin & Customer)
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    role VARCHAR(20) NOT NULL CHECK (role IN ('admin', 'customer')),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 2. Tabel Trucks (Armada Tronton)
CREATE TABLE trucks (
    id SERIAL PRIMARY KEY,
    plate_number VARCHAR(20) UNIQUE NOT NULL,
    driver_name VARCHAR(100),
    is_active BOOLEAN DEFAULT true
);

-- 3. Tabel Orders (Pesanan Logistik)
CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    order_number VARCHAR(50) UNIQUE NOT NULL,
    customer_id INT REFERENCES users(id),
    truck_id INT REFERENCES trucks(id),
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending','pickup','in_transit','delivered','cancelled')),
    origin TEXT NOT NULL,
    destination TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 4. Tabel Locations (Riwayat Perjalanan/History)
CREATE TABLE locations (
    id BIGSERIAL PRIMARY KEY,
    truck_id INT REFERENCES trucks(id),
    latitude DOUBLE PRECISION NOT NULL,
    longitude DOUBLE PRECISION NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);