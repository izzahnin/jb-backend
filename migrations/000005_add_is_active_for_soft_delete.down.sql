BEGIN;

DROP INDEX IF EXISTS idx_trips_is_active;
DROP INDEX IF EXISTS idx_orders_is_active;
DROP INDEX IF EXISTS idx_customers_is_active;

ALTER TABLE trips
  DROP COLUMN IF EXISTS is_active;

ALTER TABLE orders
  DROP COLUMN IF EXISTS is_active;

ALTER TABLE customers
  DROP COLUMN IF EXISTS is_active;

COMMIT;
