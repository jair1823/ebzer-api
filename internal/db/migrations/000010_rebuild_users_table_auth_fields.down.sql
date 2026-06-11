CREATE TABLE users_old (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'operator' CHECK(role IN ('admin', 'operator', 'guest')),
    is_active INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

INSERT INTO users_old (id, name, email, password_hash, role, is_active, created_at, updated_at)
SELECT id, name, email, password_hash, role, is_active, created_at, updated_at
FROM users;

DROP TABLE users;

ALTER TABLE users_old RENAME TO users;
