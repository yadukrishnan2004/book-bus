-- Rollback guest booking support changes
DROP INDEX IF EXISTS idx_bookings_reference;

ALTER TABLE bookings
    DROP COLUMN IF EXISTS passenger_name,
    DROP COLUMN IF EXISTS passenger_phone,
    DROP COLUMN IF EXISTS booking_reference,
    ALTER COLUMN user_id SET NOT NULL;
