-- SQLite version: Orders table with configurable statuses

-- 1) Create order_statuses table (configurable statuses)
CREATE TABLE order_statuses (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    display_name TEXT NOT NULL,
    color TEXT DEFAULT '#6B7280',
    order_position INTEGER NOT NULL,
    is_system_status INTEGER DEFAULT 0,
    is_final_status INTEGER DEFAULT 0,
    is_active INTEGER DEFAULT 1,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX idx_order_statuses_active ON order_statuses (is_active);
CREATE INDEX idx_order_statuses_order ON order_statuses (order_position);
CREATE INDEX idx_order_statuses_name ON order_statuses (name);

-- 2) Seed system statuses (three base statuses)
INSERT INTO order_statuses (name, display_name, color, order_position, is_system_status, is_final_status)
VALUES
  ('new', 'Nuevo', '#3B82F6', 1, 1, 0),
  ('completed', 'Completado', '#10B981', 100, 1, 1),
  ('cancelled', 'Cancelado', '#EF4444', 101, 1, 1);

-- 4) Create orders table (status_id is the only persisted order status)
CREATE TABLE orders (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    description TEXT NOT NULL,
    amount_charged REAL NOT NULL,
    status_id INTEGER NOT NULL DEFAULT 1 REFERENCES order_statuses(id),
    entry_date TEXT NOT NULL DEFAULT (datetime('now')),
    estimated_delivery_date TEXT,
    delivery_type TEXT NOT NULL DEFAULT 'pickup' CHECK(delivery_type IN ('pickup', 'shipping', 'delivery')),
    client_name TEXT,
    client_phone TEXT,
    notes TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

-- Indexes
CREATE INDEX idx_orders_status_id ON orders (status_id);
CREATE INDEX idx_orders_entry_date ON orders (entry_date);
CREATE INDEX idx_orders_estimated_delivery_date ON orders (estimated_delivery_date);
