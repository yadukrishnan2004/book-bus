-- Enable UUID extension (useful for primary keys later)
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name       VARCHAR(255)        NOT NULL,
    email      VARCHAR(255) UNIQUE NOT NULL,
    phone      VARCHAR(20)         NOT NULL,
    password   TEXT                NOT NULL,
    role       VARCHAR(20)         NOT NULL DEFAULT 'passenger',
    created_at TIMESTAMPTZ         NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ         NOT NULL DEFAULT NOW()
);

-- Buses table
CREATE TABLE IF NOT EXISTS buses (
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name         VARCHAR(255) NOT NULL,
    number_plate VARCHAR(50)  UNIQUE NOT NULL,
    total_seats  INT          NOT NULL,
    bus_type     VARCHAR(50)  NOT NULL DEFAULT 'standard',   -- e.g. standard, sleeper, ac
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Routes table
CREATE TABLE IF NOT EXISTS routes (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    origin      VARCHAR(255) NOT NULL,
    destination VARCHAR(255) NOT NULL,
    distance_km NUMERIC(8,2),
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Schedules table (a bus running on a route at a specific date/time)
CREATE TABLE IF NOT EXISTS schedules (
    id             UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    bus_id         UUID         NOT NULL REFERENCES buses(id)  ON DELETE CASCADE,
    route_id       UUID         NOT NULL REFERENCES routes(id) ON DELETE CASCADE,
    departure_time TIMESTAMPTZ  NOT NULL,
    arrival_time   TIMESTAMPTZ  NOT NULL,
    price          NUMERIC(10,2) NOT NULL,
    available_seats INT          NOT NULL,
    status         VARCHAR(20)  NOT NULL DEFAULT 'scheduled',  -- scheduled, cancelled, completed
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Bookings table
CREATE TABLE IF NOT EXISTS bookings (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id     UUID          NOT NULL REFERENCES users(id)     ON DELETE CASCADE,
    schedule_id UUID          NOT NULL REFERENCES schedules(id) ON DELETE CASCADE,
    seat_number INT           NOT NULL,
    status      VARCHAR(20)   NOT NULL DEFAULT 'pending',      -- pending, confirmed, cancelled
    total_price NUMERIC(10,2) NOT NULL,
    booked_at   TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    created_at  TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ   NOT NULL DEFAULT NOW(),

    UNIQUE (schedule_id, seat_number)  -- no two bookings for the same seat on same schedule
);

-- Indexes for common query patterns
CREATE INDEX IF NOT EXISTS idx_schedules_departure    ON schedules(departure_time);
CREATE INDEX IF NOT EXISTS idx_schedules_route        ON schedules(route_id);
CREATE INDEX IF NOT EXISTS idx_bookings_user          ON bookings(user_id);
CREATE INDEX IF NOT EXISTS idx_bookings_schedule      ON bookings(schedule_id);
