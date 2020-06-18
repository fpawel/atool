PRAGMA foreign_keys = ON;
PRAGMA encoding = 'UTF-8';

CREATE TABLE IF NOT EXISTS entry
(
    entry_id  INTEGER   NOT NULL PRIMARY KEY,
    stored_at TIMESTAMP NOT NULL,
    ok        BOOLEAN   NOT NULL CHECK (ok IN (0, 1)),
    text      TEXT      NOT NULL,
    indent    INTEGER   NOT NULL,
    stack     TEXT      NOT NULL
);
CREATE INDEX IF NOT EXISTS index_entry_stored_at ON entry (stored_at);

CREATE VIEW IF NOT EXISTS last_entry AS
SELECT DISTINCT CAST(strftime('%Y', created_at) AS INTEGER) AS year,
                CAST(strftime('%m', created_at) AS INTEGER) AS month,
                CAST(strftime('%d', created_at) AS INTEGER) AS day
FROM party
ORDER BY created_at DESC LIMIT 1