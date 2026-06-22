ALTER TABLE customers DROP COLUMN IF EXISTS updated_at;
ALTER TABLE trucks    DROP COLUMN IF EXISTS updated_at;
ALTER TABLE drivers   DROP COLUMN IF EXISTS created_at;
ALTER TABLE drivers   DROP COLUMN IF EXISTS updated_at;
ALTER TABLE trips     DROP COLUMN IF EXISTS updated_at;
ALTER TABLE orders    DROP COLUMN IF EXISTS updated_at;
