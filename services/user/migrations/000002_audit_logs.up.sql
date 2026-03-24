CREATE TABLE audit_logs (
    id          UUID        PRIMARY KEY,
    service     VARCHAR(100) NOT NULL,
    action      VARCHAR(255) NOT NULL,
    actor_id    TEXT        NOT NULL DEFAULT '',
    target_id   TEXT        NOT NULL DEFAULT '',
    status      VARCHAR(50) NOT NULL,
    metadata    JSONB       NOT NULL DEFAULT '{}',
    error       TEXT        NOT NULL DEFAULT '',
    timestamp   TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_audit_logs_service_action ON audit_logs (service, action, timestamp DESC);
CREATE INDEX idx_audit_logs_actor_id       ON audit_logs (actor_id,  timestamp DESC);
CREATE INDEX idx_audit_logs_target_id      ON audit_logs (target_id, timestamp DESC);
