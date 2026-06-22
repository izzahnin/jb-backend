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
-- hash: $2a$10$z/EdjLmQQfw7CCuT7PUh6eiSc2UESzezOJz8/EpQiKVItgA70u532
-- -----------------------------------------------------------------------------
INSERT INTO users (username, password_hash, full_name, role, is_active) VALUES
  ('superadmin',  '$2a$10$z/EdjLmQQfw7CCuT7PUh6eiSc2UESzezOJz8/EpQiKVItgA70u532', 'Super Admin',       'super_admin',  true),
  ('admin.sales', '$2a$10$z/EdjLmQQfw7CCuT7PUh6eiSc2UESzezOJz8/EpQiKVItgA70u532', 'Andi Kurniawan',    'admin_sales',  true),
  ('admin.ops',   '$2a$10$z/EdjLmQQfw7CCuT7PUh6eiSc2UESzezOJz8/EpQiKVItgA70u532', 'Bayu Prasetyo',     'admin_ops',    true);

-- -----------------------------------------------------------------------------
-- CUSTOMERS
-- -----------------------------------------------------------------------------
INSERT INTO customers (company_name, pic_name, phone, email, address, npwp, is_active, created_at, created_by) VALUES
  ('PT. Nusantara Logistik',   'Hendra Wijaya',   '0411-234567', 'hendra@nusantara.co.id',  'Jl. Urip Sumoharjo No. 12, Makassar',     '01.234.567.8-801.000', true, NOW() - INTERVAL '30 days', 2),
  ('CV. Maju Bersama',         'Siti Rahayu',     '0411-345678', 'siti@majubersama.com',    'Jl. Perintis Kemerdekaan KM 10, Makassar', '02.345.678.9-802.000', true, NOW() - INTERVAL '20 days', 2),
  ('PT. Sulawesi Cargo',       'Rudi Santoso',    '0411-456789', 'rudi@sulcargo.id',        'Jl. Penghibur No. 8, Makassar',            '03.456.789.0-803.000', true, NOW() - INTERVAL '10 days', 2);

-- -----------------------------------------------------------------------------
-- TRUCKS
-- -----------------------------------------------------------------------------
INSERT INTO trucks (plate_number, truck_type, status, is_active, created_at, created_by) VALUES
  ('DD 1234 ABC', 'Fuso Box 6-Roda',  'available',   true, NOW() - INTERVAL '60 days', 1),
  ('DD 5678 BCD', 'Tronton 10-Roda',  'on_duty',     true, NOW() - INTERVAL '45 days', 1),
  ('DD 9012 CDE', 'Trailer 40-Feet',  'available',   true, NOW() - INTERVAL '30 days', 1);

-- -----------------------------------------------------------------------------
-- DRIVERS
-- -----------------------------------------------------------------------------
INSERT INTO drivers (name, license_number, phone, status, is_active, created_at, created_by) VALUES
  ('Ahmad Fauzi',    'B12345678901234', '08123456789', 'available', true, NOW() - INTERVAL '60 days', 1),
  ('Dedi Kurniawan', 'B23456789012345', '08234567890', 'on_duty',   true, NOW() - INTERVAL '45 days', 1),
  ('Eko Saputra',    'B34567890123456', '08345678901', 'available', true, NOW() - INTERVAL '30 days', 1);

-- -----------------------------------------------------------------------------
-- ORDERS
-- Status bervariasi untuk demo
-- -----------------------------------------------------------------------------

-- Order 1: completed (sudah selesai 1 minggu lalu)
INSERT INTO orders (customer_id, admin_id, origin, destination,
  origin_lat, origin_lng, dest_lat, dest_lng,
  total_containers, order_date, status, is_active, created_at)
VALUES (1, 2,
  'Pelabuhan Soekarno-Hatta, Makassar', 'Pelabuhan Tanjung Perak, Surabaya',
  -5.1169, 119.4048, -7.1986, 112.7346,
  1, NOW() - INTERVAL '25 days', 'completed', true, NOW() - INTERVAL '25 days');

-- Order 2: completed (selesai 2 minggu lalu)
INSERT INTO orders (customer_id, admin_id, origin, destination,
  origin_lat, origin_lng, dest_lat, dest_lng,
  total_containers, order_date, status, is_active, created_at)
VALUES (2, 2,
  'Gudang CV. Maju Bersama, Makassar', 'Pelabuhan Makassar — Ekspor',
  -5.1400, 119.4327, -5.1169, 119.4048,
  1, NOW() - INTERVAL '18 days', 'completed', true, NOW() - INTERVAL '18 days');

-- Order 3: in_transit (sedang berjalan — untuk demo GPS live)
INSERT INTO orders (customer_id, admin_id, origin, destination,
  origin_lat, origin_lng, dest_lat, dest_lng,
  total_containers, order_date, status, is_active, created_at)
VALUES (3, 2,
  'Pelabuhan Soekarno-Hatta, Makassar', 'Gudang PT. Sulawesi Cargo, Parepare',
  -5.1169, 119.4048, -4.0135, 119.6236,
  1, NOW() - INTERVAL '3 days', 'partial', true, NOW() - INTERVAL '3 days');

-- Order 4: pending (baru dibuat kemarin)
INSERT INTO orders (customer_id, admin_id, origin, destination,
  origin_lat, origin_lng, dest_lat, dest_lng,
  total_containers, order_date, status, is_active, created_at)
VALUES (1, 2,
  'Gudang PT. Nusantara, Makassar', 'Pelabuhan Makassar',
  -5.1500, 119.4200, -5.1169, 119.4048,
  1, NOW() - INTERVAL '1 day', 'pending', true, NOW() - INTERVAL '1 day');

-- Order 5: pending (baru dibuat hari ini)
INSERT INTO orders (customer_id, admin_id, origin, destination,
  origin_lat, origin_lng, dest_lat, dest_lng,
  total_containers, order_date, status, is_active, created_at)
VALUES (2, 2,
  'Pelabuhan Makassar', 'Gudang CV. Maju Bersama, Palopo',
  -5.1169, 119.4048, -2.9922, 120.1964,
  1, NOW(), 'pending', true, NOW());

-- -----------------------------------------------------------------------------
-- TRIPS
-- Order 1 → delivered (selesai)
-- Order 2 → delivered (selesai)
-- Order 3 → in_transit (sedang berjalan — truck DD 5678 BCD, driver Dedi)
-- Order 4 & 5 → belum ada trip (masih pending)
-- -----------------------------------------------------------------------------

-- Trip untuk Order 1 (delivered)
INSERT INTO trips (order_id, truck_id, driver_id, trip_number, container_number, seal_number,
  status, is_active, created_at, start_time, end_time, created_by, started_by, completed_by)
VALUES (1, 1, 1, 'TRIP-001', 'CONT-MKS-001', 'SEAL-001',
  'delivered', true,
  NOW() - INTERVAL '25 days',
  NOW() - INTERVAL '24 days',
  NOW() - INTERVAL '22 days',
  3, 3, 3);

-- Trip untuk Order 2 (delivered)
INSERT INTO trips (order_id, truck_id, driver_id, trip_number, container_number, seal_number,
  status, is_active, created_at, start_time, end_time, created_by, started_by, completed_by)
VALUES (2, 3, 3, 'TRIP-002', 'CONT-MKS-002', 'SEAL-002',
  'delivered', true,
  NOW() - INTERVAL '18 days',
  NOW() - INTERVAL '17 days',
  NOW() - INTERVAL '16 days',
  3, 3, 3);

-- Trip untuk Order 3 (in_transit — sedang berjalan)
INSERT INTO trips (order_id, truck_id, driver_id, trip_number, container_number, seal_number,
  status, is_active, created_at, start_time, created_by, started_by)
VALUES (3, 2, 2, 'TRIP-003', 'CONT-MKS-003', 'SEAL-003',
  'in_transit', true,
  NOW() - INTERVAL '3 days',
  NOW() - INTERVAL '2 days',
  3, 3);

-- -----------------------------------------------------------------------------
-- LOCATIONS — rute Makassar → Parepare untuk Trip 3 (in_transit)
-- Titik GPS realistis sepanjang jalan Trans-Sulawesi
-- -----------------------------------------------------------------------------
INSERT INTO locations (trip_id, latitude, longitude, created_at) VALUES
  (3, -5.1169, 119.4048, NOW() - INTERVAL '47 hours'),  -- Pelabuhan Makassar (start)
  (3, -5.0823, 119.4231, NOW() - INTERVAL '46 hours'),  -- Keluar kota Makassar
  (3, -4.9756, 119.4512, NOW() - INTERVAL '44 hours'),  -- Maros
  (3, -4.8502, 119.5128, NOW() - INTERVAL '42 hours'),  -- Pangkep
  (3, -4.7234, 119.5634, NOW() - INTERVAL '40 hours'),  -- Barru (utara Pangkep)
  (3, -4.5891, 119.6012, NOW() - INTERVAL '38 hours'),  -- Mendekati Parepare
  (3, -4.3012, 119.6187, NOW() - INTERVAL '36 hours'),  -- Pinggir kota Parepare
  (3, -4.1256, 119.6215, NOW() - INTERVAL '34 hours'),  -- Kota Parepare
  (3, -4.0135, 119.6236, NOW() - INTERVAL '6 hours');   -- Tujuan (posisi terkini)

-- Update Redis-equivalent: posisi terakhir sudah ada di tabel, backend akan cache ke Redis
-- saat pertama kali ada GET request untuk trip ini.

-- -----------------------------------------------------------------------------
-- UPDATE STATUS ORDER sesuai trip yang selesai
-- -----------------------------------------------------------------------------
UPDATE orders SET status = 'completed', updated_at = NOW() - INTERVAL '22 days'
  WHERE id = 1;

UPDATE orders SET status = 'completed', updated_at = NOW() - INTERVAL '16 days'
  WHERE id = 2;

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
  RAISE NOTICE 'Login: superadmin / admin.sales / admin.ops';
  RAISE NOTICE 'Password: demo1234';
END $$;
