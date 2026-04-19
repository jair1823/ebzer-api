-- SQLite version: Expenses table
CREATE TABLE expenses (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    description TEXT NOT NULL,
    amount REAL NOT NULL,
    date TEXT NOT NULL DEFAULT (datetime('now')),
    order_id INTEGER REFERENCES orders(id) ON DELETE SET NULL,
    category_id INTEGER REFERENCES expense_categories(id) ON DELETE SET NULL,
    type TEXT NOT NULL CHECK (type IN ('general', 'order_linked')),
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

-- Important indexes
CREATE INDEX idx_expenses_date ON expenses (date);
CREATE INDEX idx_expenses_order_id ON expenses (order_id);
CREATE INDEX idx_expenses_category_id ON expenses (category_id);