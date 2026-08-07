-- Add description and is_active columns to buses table
ALTER TABLE buses
    ADD COLUMN IF NOT EXISTS description TEXT        NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS is_active   BOOLEAN     NOT NULL DEFAULT true;
