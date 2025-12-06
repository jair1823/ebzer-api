CREATE TYPE order_status AS ENUM (
    'new',
    'design',
    'pending_client',
    'production',
    'ready',
    'delivered'
);

CREATE TYPE delivery_type AS ENUM (
    'pickup',
    'shipping',
    'delivery',
    'other'
);
CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    description TEXT NOT NULL,
    amount_charged NUMERIC(10,2) NOT NULL,
    status order_status NOT NULL DEFAULT 'new',
    entry_date TIMESTAMP NOT NULL DEFAULT NOW(),
    estimated_delivery_date TIMESTAMP NULL,
    delivery_type delivery_type NOT NULL DEFAULT 'pickup',
    notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Recommended indexes:
CREATE INDEX idx_orders_status ON orders (status);
CREATE INDEX idx_orders_entry_date ON orders (entry_date);
CREATE INDEX idx_orders_estimated_delivery_date ON orders (estimated_delivery_date);