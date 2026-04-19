-- SQLite version: Financial percentages table
CREATE TABLE financial_percentages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    reinvestment_percentage REAL NOT NULL,
    supplies_percentage REAL NOT NULL,
    profit_percentage REAL NOT NULL,
    effective_start_date TEXT NOT NULL
);
