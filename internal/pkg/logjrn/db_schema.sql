PRAGMA foreign_keys = ON;
PRAGMA encoding = 'UTF-8';

CREATE TABLE IF NOT EXISTS entry
(
    entry_id  INTEGER NOT NULL PRIMARY KEY,
    stored_at REAL    NOT NULL,
    ok        BOOLEAN NOT NULL CHECK ( ok IN (0, 1) ),
    text      TEXT    NOT NULL,
    indent    INTEGER NOT NULL,
    stack     TEXT    NOT NULL
);
CREATE INDEX IF NOT EXISTS index_entry_stored_at ON entry (stored_at);