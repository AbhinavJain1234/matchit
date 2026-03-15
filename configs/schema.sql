-- MatchIt database schema
-- Run with: psql -U postgres -d matchit -f configs/schema.sql

CREATE TABLE IF NOT EXISTS riders (
    id         TEXT PRIMARY KEY,
    name       TEXT NOT NULL,
    phone      TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS drivers (
    id           TEXT PRIMARY KEY,
    name         TEXT NOT NULL,
    vehicle_type TEXT NOT NULL,
    rating       NUMERIC(3, 2) NOT NULL DEFAULT 5.0,
    status       TEXT NOT NULL DEFAULT 'offline', -- offline | available | on_trip
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS rides (
    id         TEXT PRIMARY KEY,
    rider_id   TEXT NOT NULL,
    driver_id  TEXT,                          -- NULL until a driver accepts
    pickup_lat DOUBLE PRECISION NOT NULL,
    pickup_lon DOUBLE PRECISION NOT NULL,
    dest_lat   DOUBLE PRECISION NOT NULL,
    dest_lon   DOUBLE PRECISION NOT NULL,
    status     TEXT NOT NULL DEFAULT 'REQUESTED',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index for looking up active rides by rider (used for duplicate prevention in Stage 5a)
CREATE INDEX IF NOT EXISTS idx_rides_rider_status ON rides (rider_id, status);
