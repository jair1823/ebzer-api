CREATE TABLE financial_percentages (
    id SERIAL PRIMARY KEY,
    reinvestment_percentage NUMERIC(5,2) NOT NULL,
    supplies_percentage NUMERIC(5,2) NOT NULL,
    profit_percentage NUMERIC(5,2) NOT NULL,
    effective_start_date DATE NOT NULL
);
