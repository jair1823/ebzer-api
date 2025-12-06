CREATE TYPE payment_method AS ENUM ('cash', 'sinpe', 'transfer');

CREATE TABLE income (
    id SERIAL PRIMARY KEY,
    order_id INTEGER UNIQUE NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    amount NUMERIC(10,2) NOT NULL,
    date TIMESTAMP NOT NULL DEFAULT NOW(),
    payment_method payment_method NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Index by date
CREATE INDEX idx_income_date ON income (date);