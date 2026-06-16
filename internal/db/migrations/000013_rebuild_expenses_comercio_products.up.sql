DROP TABLE IF EXISTS expense_items;
DROP TABLE IF EXISTS expenses;
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS comercios;
DROP TABLE IF EXISTS expense_categories;

CREATE TABLE comercios (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL COLLATE NOCASE UNIQUE,
    description TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE products (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    comercio_id INTEGER NOT NULL REFERENCES comercios(id) ON DELETE CASCADE,
    name TEXT NOT NULL COLLATE NOCASE,
    default_price REAL NOT NULL CHECK (default_price > 0),
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    UNIQUE (comercio_id, name)
);

CREATE TABLE expenses (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    comercio_id INTEGER NOT NULL REFERENCES comercios(id) ON DELETE RESTRICT,
    date TEXT NOT NULL DEFAULT (date('now')),
    description TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE expense_items (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    expense_id INTEGER NOT NULL REFERENCES expenses(id) ON DELETE CASCADE,
    product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    product_name TEXT NOT NULL,
    quantity REAL NOT NULL CHECK (quantity > 0),
    unit_price REAL NOT NULL CHECK (unit_price > 0),
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX idx_products_comercio_id ON products (comercio_id);
CREATE INDEX idx_expenses_comercio_id ON expenses (comercio_id);
CREATE INDEX idx_expenses_date ON expenses (date);
CREATE INDEX idx_expense_items_expense_id ON expense_items (expense_id);
CREATE INDEX idx_expense_items_product_id ON expense_items (product_id);
