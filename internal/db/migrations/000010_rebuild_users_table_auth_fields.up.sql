CREATE TABLE users_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'operator' CHECK(role IN ('admin', 'operator', 'guest')),
    is_active INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

INSERT INTO users_new (id, name, email, password_hash, role)
SELECT
    id,
    name,
    LOWER(email),
    password_hash,
    CASE role
        WHEN 'admin' THEN 'admin'
        WHEN 'guest' THEN 'guest'
        WHEN 'operator' THEN 'operator'
        WHEN 'employee' THEN 'operator'
        ELSE 'operator'
    END
FROM users;

DROP TABLE users;

ALTER TABLE users_new RENAME TO users;
