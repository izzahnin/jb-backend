-- =============================================================================
-- SEED DATA — Demo & Portfolio
-- PT. Jalur Berlian Makassar — Fleet Management System
--
-- Cara pakai:
--   docker exec -i jbm_postgres psql -U admin -d jalur_berlian_db < migrations/seed.sql
--
-- Password semua user: demo1234
-- Hapus data lama sebelum insert ulang (idempotent)
-- =============================================================================

-- -----------------------------------------------------------------------------
-- BERSIHKAN DATA LAMA (urutan: child → parent)
-- -----------------------------------------------------------------------------
DELETE FROM locations;
DELETE FROM trips;
DELETE FROM orders;
DELETE FROM drivers;
DELETE FROM trucks;
DELETE FROM customers;
DELETE FROM users;

-- Reset sequences
ALTER SEQUENCE users_id_seq     RESTART WITH 1;
ALTER SEQUENCE customers_id_seq RESTART WITH 1;
ALTER SEQUENCE trucks_id_seq    RESTART WITH 1;
ALTER SEQUENCE drivers_id_seq   RESTART WITH 1;
ALTER SEQUENCE orders_id_seq    RESTART WITH 1;
ALTER SEQUENCE trips_id_seq     RESTART WITH 1;
ALTER SEQUENCE locations_id_seq RESTART WITH 1;

-- -----------------------------------------------------------------------------
-- USERS  (password: demo1234)
-- Hash di-generate via: SELECT crypt('demo1234', gen_salt('bf', 10));
-- Gunakan pgcrypto jika perlu reset: UPDATE users SET password_hash = crypt('demo1234', gen_salt('bf', 10));
-- -----------------------------------------------------------------------------
CREATE EXTENSION IF NOT EXISTS pgcrypto;

INSERT INTO users (username, password_hash, full_name, role, is_active) VALUES
  ('superadmin',  crypt('demo1234', gen_salt('bf', 10)), 'Rizky Aditya',   'super_admin',  true),
  ('admin.sales', crypt('demo1234', gen_salt('bf', 10)), 'Andi Kurniawan', 'admin_sales',  true),
  ('admin.ops',   crypt('demo1234', gen_salt('bf', 10)), 'Bayu Prasetyo',  'admin_ops',    true),
  ('demo',        crypt('demo1234', gen_salt('bf', 10)), 'Demo Viewer',    'demo',         true);

-- -----------------------------------------------------------------------------
-- CUSTOMERS (5 perusahaan logistik/industri Sulawesi)
-- -----------------------------------------------------------------------------
INSERT INTO customers (company_name, pic_name, phone, email, address, npwp, is_active, created_at, created_by) VALUES
  ('PT. Nusantara Logistik',    'Hendra Wijaya',   '0411-234567', 'hendra@nusantara.co.id',    'Jl. Urip Sumoharjo No. 12, Makassar',          '01.234.567.8-801.000', true, NOW() - INTERVAL '45 days', 2),
  ('CV. Maju Bersama',          'Siti Rahayu',     '0411-345678', 'siti@majubersama.com',       'Jl. Perintis Kemerdekaan KM 10, Makassar',      '02.345.678.9-802.000', true, NOW() - INTERVAL '38 days', 2),
  ('PT. Sulawesi Cargo',        'Rudi Santoso',    '0421-456789', 'rudi@sulcargo.id',           'Jl. Bau Massepe No. 45, Parepare',              '03.456.789.0-803.000', true, NOW() - INTERVAL '30 days', 2),
  ('PT. Timur Jaya Ekspres',    'Farhan Hidayat',  '0411-567890', 'farhan@timurjaya.co.id',     'Jl. Pettarani No. 78, Makassar',                '04.567.890.1-804.000', true, NOW() - INTERVAL '20 days', 2),
  ('CV. Berlian Abadi',         'Nurhayati',       '0411-678901', 'nurhayati@berlianabadi.com', 'Jl. Boulevard Raya Blok A No. 5, Makassar',     '05.678.901.2-805.000', true, NOW() - INTERVAL '10 days', 2);

-- -----------------------------------------------------------------------------
-- TRUCKS (5 armada berbagai tipe)
-- -----------------------------------------------------------------------------
INSERT INTO trucks (plate_number, truck_type, status, is_active, created_at, created_by) VALUES
  ('DD 1234 ABC', 'Fuso Box 6-Roda',    'available',   true, NOW() - INTERVAL '90 days', 1),
  ('DD 5678 BCD', 'Tronton 10-Roda',    'on_duty',     true, NOW() - INTERVAL '75 days', 1),
  ('DD 9012 CDE', 'Trailer 40-Feet',    'available',   true, NOW() - INTERVAL '60 days', 1),
  ('DD 3456 DEF', 'CDD Box 4-Roda',     'on_duty',     true, NOW() - INTERVAL '45 days', 1),
  ('DD 7890 EFG', 'Fuso Bak Terbuka',   'available',   true, NOW() - INTERVAL '30 days', 1);

-- -----------------------------------------------------------------------------
-- DRIVERS (5 pengemudi, 2 sedang bertugas)
-- -----------------------------------------------------------------------------
INSERT INTO drivers (name, license_number, phone, status, is_active, created_at, created_by) VALUES
  ('Ahmad Fauzi',     'B12345678901234', '08123456789', 'available', true, NOW() - INTERVAL '90 days', 1),
  ('Dedi Kurniawan',  'B23456789012345', '08234567890', 'on_duty',   true, NOW() - INTERVAL '75 days', 1),
  ('Eko Saputra',     'B34567890123456', '08345678901', 'available', true, NOW() - INTERVAL '60 days', 1),
  ('Fajar Ramadan',   'B45678901234567', '08456789012', 'on_duty',   true, NOW() - INTERVAL '45 days', 1),
  ('Galih Pratama',   'B56789012345678', '08567890123', 'available', true, NOW() - INTERVAL '30 days', 1);

-- -----------------------------------------------------------------------------
-- ORDERS (5 order, status bervariasi untuk demo)
-- -----------------------------------------------------------------------------

-- Order 1: completed (selesai 3 minggu lalu) — Nusantara Logistik
INSERT INTO orders (customer_id, admin_id, origin, destination,
  origin_lat, origin_lng, dest_lat, dest_lng,
  total_containers, order_date, status, is_active, created_at)
VALUES (1, 2,
  'Pelabuhan Soekarno-Hatta, Makassar', 'Pelabuhan Tanjung Perak, Surabaya',
  -5.1169, 119.4048, -7.1986, 112.7346,
  1, NOW() - INTERVAL '25 days', 'completed', true, NOW() - INTERVAL '25 days');

-- Order 2: completed (selesai 2 minggu lalu) — CV. Maju Bersama
INSERT INTO orders (customer_id, admin_id, origin, destination,
  origin_lat, origin_lng, dest_lat, dest_lng,
  total_containers, order_date, status, is_active, created_at)
VALUES (2, 2,
  'Gudang CV. Maju Bersama, Makassar', 'Pelabuhan Makassar — Ekspor',
  -5.1400, 119.4327, -5.1169, 119.4048,
  1, NOW() - INTERVAL '18 days', 'completed', true, NOW() - INTERVAL '18 days');

-- Order 3: partial (sedang berjalan) — PT. Sulawesi Cargo
INSERT INTO orders (customer_id, admin_id, origin, destination,
  origin_lat, origin_lng, dest_lat, dest_lng,
  total_containers, order_date, status, is_active, created_at)
VALUES (3, 2,
  'Pelabuhan Soekarno-Hatta, Makassar', 'Gudang PT. Sulawesi Cargo, Parepare',
  -5.1169, 119.4048, -4.0135, 119.6236,
  1, NOW() - INTERVAL '3 days', 'partial', true, NOW() - INTERVAL '3 days');

-- Order 4: pending (kemarin) — PT. Timur Jaya Ekspres
INSERT INTO orders (customer_id, admin_id, origin, destination,
  origin_lat, origin_lng, dest_lat, dest_lng,
  total_containers, order_date, status, is_active, created_at)
VALUES (4, 2,
  'Gudang PT. Timur Jaya Ekspres, Makassar', 'Pelabuhan Makassar',
  -5.1478, 119.4327, -5.1169, 119.4048,
  1, NOW() - INTERVAL '1 day', 'pending', true, NOW() - INTERVAL '1 day');

-- Order 5: pending (hari ini) — CV. Berlian Abadi
INSERT INTO orders (customer_id, admin_id, origin, destination,
  origin_lat, origin_lng, dest_lat, dest_lng,
  total_containers, order_date, status, is_active, created_at)
VALUES (5, 2,
  'Pelabuhan Makassar', 'Gudang CV. Berlian Abadi, Palopo',
  -5.1169, 119.4048, -2.9922, 120.1964,
  1, NOW(), 'pending', true, NOW());

-- -----------------------------------------------------------------------------
-- TRIPS
-- Order 1 → delivered | Order 2 → delivered | Order 3 → in_transit
-- Order 4 & 5 → belum ada trip (masih pending)
-- -----------------------------------------------------------------------------

-- Trip untuk Order 1 (delivered, truck Fuso Box, driver Ahmad)
INSERT INTO trips (order_id, truck_id, driver_id, trip_number, container_number, seal_number,
  status, is_active, created_at, start_time, end_time, created_by, started_by, completed_by)
VALUES (1, 1, 1, 'TRIP-001', 'CONT-MKS-2506-001', 'SEAL-JB-001',
  'delivered', true,
  NOW() - INTERVAL '25 days',
  NOW() - INTERVAL '24 days',
  NOW() - INTERVAL '22 days',
  3, 3, 3);

-- Trip untuk Order 2 (delivered, truck Trailer, driver Eko)
INSERT INTO trips (order_id, truck_id, driver_id, trip_number, container_number, seal_number,
  status, is_active, created_at, start_time, end_time, created_by, started_by, completed_by)
VALUES (2, 3, 3, 'TRIP-002', 'CONT-MKS-2506-002', 'SEAL-JB-002',
  'delivered', true,
  NOW() - INTERVAL '18 days',
  NOW() - INTERVAL '17 days',
  NOW() - INTERVAL '16 days',
  3, 3, 3);

-- Trip untuk Order 3 (in_transit — sedang berjalan, truck Tronton, driver Dedi)
INSERT INTO trips (order_id, truck_id, driver_id, trip_number, container_number, seal_number,
  status, is_active, created_at, start_time, created_by, started_by)
VALUES (3, 2, 2, 'TRIP-003', 'CONT-MKS-2506-003', 'SEAL-JB-003',
  'in_transit', true,
  NOW() - INTERVAL '3 days',
  NOW() - INTERVAL '2 days',
  3, 3);

-- -----------------------------------------------------------------------------
-- LOCATIONS — rute GPS realistis Makassar → Parepare untuk Trip 3 (in_transit)
-- Titik-titik sepanjang jalan Trans-Sulawesi
-- -----------------------------------------------------------------------------
INSERT INTO locations (trip_id, latitude, longitude, created_at) VALUES
  (3, -5.1169, 119.4048, NOW() - INTERVAL '47 hours'),  -- Pelabuhan Soekarno-Hatta, Makassar
  (3, -5.0980, 119.4152, NOW() - INTERVAL '46 hours'),  -- Jl. Nusantara, Makassar Utara
  (3, -5.0823, 119.4231, NOW() - INTERVAL '45 hours'),  -- Perbatasan kota Makassar
  (3, -4.9756, 119.4512, NOW() - INTERVAL '43 hours'),  -- Kabupaten Maros
  (3, -4.8502, 119.5128, NOW() - INTERVAL '41 hours'),  -- Kabupaten Pangkep
  (3, -4.7780, 119.5380, NOW() - INTERVAL '39 hours'),  -- Labakkang, Pangkep
  (3, -4.6891, 119.5612, NOW() - INTERVAL '37 hours'),  -- Barru selatan
  (3, -4.5678, 119.5890, NOW() - INTERVAL '35 hours'),  -- Kota Barru
  (3, -4.4234, 119.6012, NOW() - INTERVAL '33 hours'),  -- Tanete Rilau, Barru
  (3, -4.3012, 119.6187, NOW() - INTERVAL '20 hours'),  -- Pinggir kota Parepare
  (3, -4.1256, 119.6215, NOW() - INTERVAL '10 hours'),  -- Kota Parepare
  (3, -4.0135, 119.6236, NOW() - INTERVAL '2 hours');   -- Gudang PT. Sulawesi Cargo (posisi terkini)

-- -----------------------------------------------------------------------------
-- UPDATE STATUS ORDER sesuai trip yang selesai
-- -----------------------------------------------------------------------------
UPDATE orders SET status = 'completed', updated_at = NOW() - INTERVAL '22 days' WHERE id = 1;
UPDATE orders SET status = 'completed', updated_at = NOW() - INTERVAL '16 days' WHERE id = 2;

-- -----------------------------------------------------------------------------
-- SUMMARY (untuk verifikasi)
-- -----------------------------------------------------------------------------
DO $$
BEGIN
  RAISE NOTICE '=== SEED BERHASIL ===';
  RAISE NOTICE 'Users    : % baris', (SELECT COUNT(*) FROM users);
  RAISE NOTICE 'Customers: % baris', (SELECT COUNT(*) FROM customers);
  RAISE NOTICE 'Trucks   : % baris', (SELECT COUNT(*) FROM trucks);
  RAISE NOTICE 'Drivers  : % baris', (SELECT COUNT(*) FROM drivers);
  RAISE NOTICE 'Orders   : % baris', (SELECT COUNT(*) FROM orders);
  RAISE NOTICE 'Trips    : % baris', (SELECT COUNT(*) FROM trips);
  RAISE NOTICE 'Locations: % baris', (SELECT COUNT(*) FROM locations);
  RAISE NOTICE '====================';
  RAISE NOTICE 'Login: superadmin / admin.sales / admin.ops / demo';
  RAISE NOTICE 'Password: demo1234 (semua)';
  RAISE NOTICE 'demo = read-only access ke semua halaman';
END $$;
