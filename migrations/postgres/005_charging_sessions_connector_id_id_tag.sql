ALTER TABLE charging_sessions
ADD COLUMN IF NOT EXISTS connector_id INT;

ALTER TABLE charging_sessions
ADD COLUMN IF NOT EXISTS id_tag VARCHAR(100);

CREATE INDEX IF NOT EXISTS idx_charging_sessions_connector_id
ON charging_sessions(connector_id);

CREATE INDEX IF NOT EXISTS idx_charging_sessions_id_tag
ON charging_sessions(id_tag);