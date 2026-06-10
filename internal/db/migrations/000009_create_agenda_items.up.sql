-- Agenda items table for daily work management: notes, tasks, and reminders
CREATE TABLE agenda_items (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    type TEXT NOT NULL DEFAULT 'note' CHECK(type IN ('note', 'task', 'reminder')),
    title TEXT NOT NULL,
    content TEXT,
    status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending', 'done', 'archived')),
    priority TEXT NOT NULL DEFAULT 'medium' CHECK(priority IN ('low', 'medium', 'high')),
    due_date TEXT,
    completed_at TEXT,
    order_id INTEGER REFERENCES orders(id) ON DELETE SET NULL,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX idx_agenda_items_status ON agenda_items (status);
CREATE INDEX idx_agenda_items_type ON agenda_items (type);
CREATE INDEX idx_agenda_items_priority ON agenda_items (priority);
CREATE INDEX idx_agenda_items_due_date ON agenda_items (due_date);
CREATE INDEX idx_agenda_items_order_id ON agenda_items (order_id);
