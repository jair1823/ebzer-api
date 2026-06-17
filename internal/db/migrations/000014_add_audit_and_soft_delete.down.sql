DROP INDEX IF EXISTS idx_audit_events_created_at;
DROP INDEX IF EXISTS idx_audit_events_entity;
DROP INDEX IF EXISTS idx_agenda_items_deleted_at;
DROP INDEX IF EXISTS idx_expenses_deleted_at;
DROP INDEX IF EXISTS idx_income_deleted_at;
DROP INDEX IF EXISTS idx_orders_deleted_at;

DROP TABLE IF EXISTS audit_events;
