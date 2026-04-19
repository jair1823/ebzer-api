-- SQLite version: Orders table
CREATE TABLE orders (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    description TEXT NOT NULL,
    amount_charged REAL NOT NULL,
    status TEXT NOT NULL DEFAULT 'confirmed' CHECK(status IN ('confirmed', 'in_progress', 'ready', 'shipped', 'delivered', 'cancelled')),
    entry_date TEXT NOT NULL DEFAULT (datetime('now')),
    estimated_delivery_date TEXT,
    delivery_type TEXT NOT NULL DEFAULT 'pickup' CHECK(delivery_type IN ('pickup', 'shipping', 'delivery')),
    client_name TEXT,
    client_phone TEXT,
    notes TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

-- Recommended indexes:
CREATE INDEX idx_orders_status ON orders (status);
CREATE INDEX idx_orders_entry_date ON orders (entry_date);
CREATE INDEX idx_orders_estimated_delivery_date ON orders (estimated_delivery_date);