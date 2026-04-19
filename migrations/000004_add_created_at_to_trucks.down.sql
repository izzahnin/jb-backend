-- Remove created_at column from trucks table
DROP INDEX IF EXISTS idx_trucks_created_at;
ALTER TABLE trucks
DROP COLUMN created_at;
