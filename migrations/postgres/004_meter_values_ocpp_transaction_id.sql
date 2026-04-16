ALTER TABLE meter_values
ADD COLUMN IF NOT EXISTS ocpp_transaction_id BIGINT;

CREATE INDEX IF NOT EXISTS idx_meter_values_ocpp_transaction_id
ON meter_values(ocpp_transaction_id);