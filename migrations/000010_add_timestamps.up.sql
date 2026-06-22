ALTER TABLE customers ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE trucks    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE drivers   ADD COLUMN IF NOT EXISTS created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE drivers   ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE trips     ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE orders    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE;

-- Backfill existing records so they have timestamps
UPDATE customers SET updated_at = created_at WHERE updated_at IS NULL AND created_at IS NOT NULL;
UPDATE trucks    SET updated_at = created_at WHERE updated_at IS NULL AND created_at IS NOT NULL;
UPDATE trips     SET updated_at = created_at WHERE updated_at IS NULL AND created_at IS NOT NULL;
UPDATE orders    SET updated_at = order_date  WHERE updated_at IS NULL;
