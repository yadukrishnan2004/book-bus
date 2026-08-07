-- Update bookings table to support guest booking (no auth yet)
-- and add booking_reference to group multi-seat bookings together.

ALTER TABLE bookings
    ALTER COLUMN user_id DROP NOT NULL,
    ADD COLUMN IF NOT EXISTS passenger_name    VARCHAR(255) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS passenger_phone   VARCHAR(20)  NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS booking_reference UUID         NOT NULL DEFAULT uuid_generate_v4();

CREATE INDEX IF NOT EXISTS idx_bookings_reference ON bookings(booking_reference);
