-- Rollback: remove description and is_active from buses
ALTER TABLE buses
    DROP COLUMN IF EXISTS description,
    DROP COLUMN IF EXISTS is_active;
