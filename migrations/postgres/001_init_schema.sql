CREATE EXTENSION IF NOT EXISTS "pgcrypto";

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'user_role') THEN
        CREATE TYPE user_role AS ENUM ('customer', 'operator', 'owner', 'admin');
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'charger_status') THEN
        CREATE TYPE charger_status AS ENUM (
            'available',
            'preparing',
            'charging',
            'finishing',
            'faulted',
            'offline',
            'unavailable'
        );
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'session_status') THEN
        CREATE TYPE session_status AS ENUM (
            'requested',
            'authorized',
            'in_progress',
            'completed',
            'stopped',
            'failed',
            'cancelled'
        );
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    full_name VARCHAR(120) NOT NULL,
    email VARCHAR(120) UNIQUE NOT NULL,
    mobile VARCHAR(20) UNIQUE,
    password_hash TEXT NOT NULL,
    role user_role NOT NULL DEFAULT 'customer',
    wallet_balance NUMERIC(12,2) NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS stations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID REFERENCES users(id) ON DELETE SET NULL,
    name VARCHAR(150) NOT NULL,
    address_line1 TEXT NOT NULL,
    address_line2 TEXT,
    city VARCHAR(80) NOT NULL,
    district VARCHAR(80),
    state VARCHAR(80) NOT NULL,
    country VARCHAR(80) NOT NULL DEFAULT 'India',
    pincode VARCHAR(20),
    latitude DOUBLE PRECISION,
    longitude DOUBLE PRECISION,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS chargers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    station_id UUID NOT NULL REFERENCES stations(id) ON DELETE CASCADE,
    ocpp_id VARCHAR(100) UNIQUE NOT NULL,
    vendor VARCHAR(100),
    model VARCHAR(100),
    firmware_version VARCHAR(100),
    connector_count INT NOT NULL DEFAULT 1,
    status charger_status NOT NULL DEFAULT 'offline',
    last_seen_at TIMESTAMPTZ,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS charging_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    charger_id UUID NOT NULL REFERENCES chargers(id) ON DELETE CASCADE,
    station_id UUID NOT NULL REFERENCES stations(id) ON DELETE CASCADE,
    status session_status NOT NULL DEFAULT 'requested',
    meter_start NUMERIC(12,3),
    meter_stop NUMERIC(12,3),
    energy_kwh NUMERIC(12,3) NOT NULL DEFAULT 0,
    amount NUMERIC(12,2) NOT NULL DEFAULT 0,
    started_at TIMESTAMPTZ,
    ended_at TIMESTAMPTZ,
    stop_reason VARCHAR(100),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_stations_owner_id ON stations(owner_id);
CREATE INDEX IF NOT EXISTS idx_chargers_station_id ON chargers(station_id);
CREATE INDEX IF NOT EXISTS idx_chargers_ocpp_id ON chargers(ocpp_id);
CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON charging_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_charger_id ON charging_sessions(charger_id);