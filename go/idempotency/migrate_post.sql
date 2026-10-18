CREATE INDEX IF NOT EXISTS idx_idempotency_expiry ON idempotency.tbl_idempotency_records (dt_expires_at)
WHERE str_status = 'COMPLETED';
