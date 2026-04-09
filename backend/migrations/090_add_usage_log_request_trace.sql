ALTER TABLE usage_logs
    ADD COLUMN IF NOT EXISTS request_body TEXT,
    ADD COLUMN IF NOT EXISTS request_body_truncated BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS request_body_bytes INT,
    ADD COLUMN IF NOT EXISTS response_body TEXT,
    ADD COLUMN IF NOT EXISTS response_body_truncated BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS response_body_bytes INT;
