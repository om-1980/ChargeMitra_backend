CREATE TABLE IF NOT EXISTS ocpp_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    charger_id UUID REFERENCES chargers(id) ON DELETE CASCADE,
    ocpp_id VARCHAR(100) NOT NULL,
    direction VARCHAR(20) NOT NULL,
    action VARCHAR(100),
    message_id VARCHAR(100),
    payload JSONB,
    raw_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ocpp_events_charger_id
ON ocpp_events(charger_id);

CREATE INDEX IF NOT EXISTS idx_ocpp_events_ocpp_id
ON ocpp_events(ocpp_id);

CREATE INDEX IF NOT EXISTS idx_ocpp_events_action
ON ocpp_events(action);

CREATE INDEX IF NOT EXISTS idx_ocpp_events_created_at
ON ocpp_events(created_at);