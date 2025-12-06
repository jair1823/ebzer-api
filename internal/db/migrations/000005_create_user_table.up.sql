CREATE TYPE user_role AS ENUM ('admin', 'employee');

CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(150) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    role user_role NOT NULL DEFAULT 'employee',
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);