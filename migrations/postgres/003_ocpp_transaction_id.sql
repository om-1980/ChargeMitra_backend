ALTER TABLE charging_sessions
ADD COLUMN IF NOT EXISTS ocpp_transaction_id BIGSERIAL;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'uq_charging_sessions_ocpp_transaction_id'
    ) THEN
        ALTER TABLE charging_sessions
        ADD CONSTRAINT uq_charging_sessions_ocpp_transaction_id UNIQUE (ocpp_transaction_id);
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_charging_sessions_ocpp_transaction_id
ON charging_sessions(ocpp_transaction_id);