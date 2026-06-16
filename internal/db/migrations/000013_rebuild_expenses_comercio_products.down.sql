DROP TABLE IF EXISTS expense_items;
DROP TABLE IF EXISTS expenses;
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS comercios;

CREATE TABLE expense_categories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE expenses (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    description TEXT NOT NULL,
    amount REAL NOT NULL CHECK (amount > 0),
    date TEXT NOT NULL DEFAULT (datetime('now')),
    order_id INTEGER REFERENCES orders(id) ON DELETE SET NULL,
    category_id INTEGER REFERENCES expense_categories(id) ON DELETE SET NULL,
    type TEXT NOT NULL DEFAULT 'general',
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX idx_expenses_date ON expenses (date);
CREATE INDEX idx_expenses_order_id ON expenses (order_id);
CREATE INDEX idx_expenses_category_id ON expenses (category_id);
