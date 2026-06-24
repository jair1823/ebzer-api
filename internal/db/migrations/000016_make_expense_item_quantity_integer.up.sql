ALTER TABLE expense_items RENAME TO expense_items_old;

CREATE TABLE expense_items (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    expense_id INTEGER NOT NULL REFERENCES expenses(id) ON DELETE CASCADE,
    product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    product_name TEXT NOT NULL,
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    unit_price REAL NOT NULL CHECK (unit_price > 0),
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

INSERT INTO expense_items (id, expense_id, product_id, product_name, quantity, unit_price, created_at)
SELECT
    id,
    expense_id,
    product_id,
    product_name,
    CAST(quantity AS INTEGER) + CASE WHEN quantity > CAST(quantity AS INTEGER) THEN 1 ELSE 0 END,
    unit_price,
    created_at
FROM expense_items_old;

DROP TABLE expense_items_old;

CREATE INDEX idx_expense_items_expense_id ON expense_items (expense_id);
CREATE INDEX idx_expense_items_product_id ON expense_items (product_id);
