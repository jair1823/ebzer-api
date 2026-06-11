CREATE TABLE users_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    username TEXT UNIQUE NOT NULL,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'operator' CHECK(role IN ('admin', 'operator', 'guest')),
    is_active INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

INSERT INTO users_new (id, name, username, email, password_hash, role, is_active, created_at, updated_at)
SELECT
    id,
    name,
    LOWER(substr(email, 1, instr(email, '@') - 1)),
    email,
    password_hash,
    role,
    is_active,
    created_at,
    updated_at
FROM users;

DROP TABLE users;

ALTER TABLE users_new RENAME TO users;
