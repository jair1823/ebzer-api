ALTER TABLE orders ADD COLUMN deleted_at TEXT;
ALTER TABLE orders ADD COLUMN deleted_by INTEGER REFERENCES users(id);

ALTER TABLE income ADD COLUMN deleted_at TEXT;
ALTER TABLE income ADD COLUMN deleted_by INTEGER REFERENCES users(id);

ALTER TABLE expenses ADD COLUMN deleted_at TEXT;
ALTER TABLE expenses ADD COLUMN deleted_by INTEGER REFERENCES users(id);

ALTER TABLE agenda_items ADD COLUMN deleted_at TEXT;
ALTER TABLE agenda_items ADD COLUMN deleted_by INTEGER REFERENCES users(id);

CREATE TABLE audit_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    entity_type TEXT NOT NULL,
    entity_id INTEGER NOT NULL,
    action TEXT NOT NULL,
    actor_user_id INTEGER REFERENCES users(id),
    actor_username TEXT,
    summary TEXT,
    before_json TEXT,
    after_json TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX idx_orders_deleted_at ON orders (deleted_at);
CREATE INDEX idx_income_deleted_at ON income (deleted_at);
CREATE INDEX idx_expenses_deleted_at ON expenses (deleted_at);
CREATE INDEX idx_agenda_items_deleted_at ON agenda_items (deleted_at);
CREATE INDEX idx_audit_events_entity ON audit_events (entity_type, entity_id);
CREATE INDEX idx_audit_events_created_at ON audit_events (created_at);
