ALTER TABLE expense_items RENAME TO expense_items_integer;

CREATE TABLE expense_items (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    expense_id INTEGER NOT NULL REFERENCES expenses(id) ON DELETE CASCADE,
    product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    product_name TEXT NOT NULL,
    quantity REAL NOT NULL CHECK (quantity > 0),
    unit_price REAL NOT NULL CHECK (unit_price > 0),
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

INSERT INTO expense_items (id, expense_id, product_id, product_name, quantity, unit_price, created_at)
SELECT id, expense_id, product_id, product_name, quantity, unit_price, created_at
FROM expense_items_integer;

DROP TABLE expense_items_integer;

CREATE INDEX idx_expense_items_expense_id ON expense_items (expense_id);
CREATE INDEX idx_expense_items_product_id ON expense_items (product_id);
