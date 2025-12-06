CREATE TABLE expenses (
    id SERIAL PRIMARY KEY,
    description TEXT NOT NULL,
    amount NUMERIC(10,2) NOT NULL,
    date TIMESTAMP NOT NULL DEFAULT NOW(),

    order_id INTEGER REFERENCES orders(id) ON DELETE SET NULL,
    category_id INTEGER REFERENCES expense_categories(id) ON DELETE SET NULL,

    type VARCHAR(20) NOT NULL CHECK (type IN ('general', 'order_linked')),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Important indexes
CREATE INDEX idx_expenses_date ON expenses (date);
CREATE INDEX idx_expenses_order_id ON expenses (order_id);
CREATE INDEX idx_expenses_category_id ON expenses (category_id);