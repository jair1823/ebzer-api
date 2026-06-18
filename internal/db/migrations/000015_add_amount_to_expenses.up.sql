ALTER TABLE expenses
ADD COLUMN amount REAL CHECK (amount IS NULL OR amount > 0);
