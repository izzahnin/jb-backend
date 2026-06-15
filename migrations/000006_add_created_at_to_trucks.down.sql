BEGIN;

-- Drop index first (if it exists)
DROP INDEX IF EXISTS idx_trucks_created_at;

-- Remove created_at column from trucks
ALTER TABLE trucks
DROP COLUMN IF EXISTS created_at;

COMMIT;
