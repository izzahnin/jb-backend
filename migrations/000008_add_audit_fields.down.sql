ALTER TABLE trips     DROP COLUMN IF EXISTS completed_by;
ALTER TABLE trips     DROP COLUMN IF EXISTS started_by;
ALTER TABLE trips     DROP COLUMN IF EXISTS created_by;
ALTER TABLE drivers   DROP COLUMN IF EXISTS created_by;
ALTER TABLE customers DROP COLUMN IF EXISTS created_by;
ALTER TABLE trucks    DROP COLUMN IF EXISTS created_by;
