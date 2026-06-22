ALTER TABLE trucks    DROP COLUMN IF EXISTS updated_by;
ALTER TABLE customers DROP COLUMN IF EXISTS updated_by;
ALTER TABLE drivers   DROP COLUMN IF EXISTS updated_by;
