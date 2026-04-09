CREATE TABLE IF NOT EXISTS meter_values (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES charging_sessions(id) ON DELETE CASCADE,
    charger_id UUID NOT NULL REFERENCES chargers(id) ON DELETE CASCADE,
    sampled_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    measurand VARCHAR(100) NOT NULL DEFAULT 'Energy.Active.Import.Register',
    value NUMERIC(12,3) NOT NULL,
    unit VARCHAR(20) NOT NULL DEFAULT 'Wh',
    context VARCHAR(100),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_meter_values_session_id ON meter_values(session_id);
CREATE INDEX IF NOT EXISTS idx_meter_values_charger_id ON meter_values(charger_id);
CREATE INDEX IF NOT EXISTS idx_meter_values_sampled_at ON meter_values(sampled_at);