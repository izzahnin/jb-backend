ALTER TABLE trucks    ADD COLUMN IF NOT EXISTS created_by   INT REFERENCES users(id);
ALTER TABLE customers ADD COLUMN IF NOT EXISTS created_by   INT REFERENCES users(id);
ALTER TABLE drivers   ADD COLUMN IF NOT EXISTS created_by   INT REFERENCES users(id);
ALTER TABLE trips     ADD COLUMN IF NOT EXISTS created_by   INT REFERENCES users(id);
ALTER TABLE trips     ADD COLUMN IF NOT EXISTS started_by   INT REFERENCES users(id);
ALTER TABLE trips     ADD COLUMN IF NOT EXISTS completed_by INT REFERENCES users(id);
