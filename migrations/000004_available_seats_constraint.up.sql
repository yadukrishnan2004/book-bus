-- Add a non-negative constraint on schedules.available_seats to prevent
-- the counter from drifting below zero due to bugs or manual edits.
-- Any negative value would cause all future bookings to fail with a false
-- "no seats available" error.

ALTER TABLE schedules
    ADD CONSTRAINT chk_available_seats_non_negative
    CHECK (available_seats >= 0);
