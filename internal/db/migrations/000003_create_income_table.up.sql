-- SQLite version: Income table
CREATE TABLE income (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    order_id INTEGER NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    amount REAL NOT NULL,
    date TEXT NOT NULL DEFAULT (datetime('now')),
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

-- Index by order_id for efficient lookup of multiple incomes per order
CREATE INDEX idx_income_order_id ON income (order_id);

-- Index by date
CREATE INDEX idx_income_date ON income (date);