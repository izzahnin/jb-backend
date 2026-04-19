-- Add created_at column to trucks table
ALTER TABLE trucks
ADD COLUMN created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;

-- Create index for ordering by created_at
CREATE INDEX idx_trucks_created_at ON trucks(created_at DESC);
