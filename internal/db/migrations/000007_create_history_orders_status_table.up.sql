CREATE TABLE history_order_status (
    id SERIAL PRIMARY KEY,
    order_id INTEGER NOT NULL REFERENCES orders(id) ON DELETE CASCADE,

    previous_status order_status,
    new_status order_status NOT NULL,

    change_date TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Index for queries by order
CREATE INDEX idx_history_order_id ON history_order_status (order_id);