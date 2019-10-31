PRAGMA foreign_keys = ON;
PRAGMA encoding = 'UTF-8';

CREATE TABLE IF NOT EXISTS party
(
    party_id   INTEGER PRIMARY KEY NOT NULL,
    created_at TIMESTAMP           NOT NULL DEFAULT (datetime('now')) UNIQUE
);

CREATE TABLE IF NOT EXISTS hardware
(
    device TEXT NOT NULL DEFAULT 'DEFAULT_DEVICE',
    baud INTEGER NOT NULL DEFAULT 9600,
    timeout_get_response INTEGER NOT NULL DEFAULT 1000000000, -- в наносекундах = 1 секунда
    timeout_end_response INTEGER NOT NULL DEFAULT 50000000, -- в наносекундах = 50 миллисекунд
    pause INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS product
(
    product_id INTEGER PRIMARY KEY NOT NULL,
    party_id   INTEGER             NOT NULL,
    created_at TIMESTAMP           NOT NULL DEFAULT (datetime('now')) UNIQUE,
    device     TEXT                NOT NULL DEFAULT 'DEFAULT',
    serial     INTEGER             NOT NULL CHECK (serial > 0 ),
    comport    TEXT             NOT NULL DEFAULT 'COM1' CHECK (comport >= 1 ),
    addr       INTEGER             NOT NULL DEFAULT 1 CHECK (addr >= 1 ),
    checked    BOOLEAN             NOT NULL DEFAULT 1 CHECK (checked IN (0, 1) ),
    UNIQUE (party_id, comport, addr),
    UNIQUE (party_id, device, serial),
    FOREIGN KEY (party_id) REFERENCES party (party_id) ON DELETE CASCADE,
    FOREIGN KEY (device) REFERENCES hardware (device) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS index_product_serial ON product (serial);

DROP VIEW IF EXISTS last_party;
CREATE VIEW IF NOT EXISTS last_party AS
SELECT *
FROM party
ORDER BY created_at DESC
LIMIT 1;

DROP VIEW IF EXISTS last_party_products;
CREATE VIEW IF NOT EXISTS last_party_products AS
SELECT *
FROM product
WHERE party_id = (SELECT party_id FROM last_party)
ORDER BY created_at;

CREATE TABLE IF NOT EXISTS bucket
(
    bucket_id  INTEGER PRIMARY KEY NOT NULL,
    created_at TIMESTAMP           NOT NULL UNIQUE DEFAULT (datetime('now')),
    updated_at TIMESTAMP           NOT NULL        DEFAULT (datetime('now')),
    party_id   INTEGER             NOT NULL,
    FOREIGN KEY (party_id) REFERENCES party (party_id) ON DELETE CASCADE
);

CREATE VIEW IF NOT EXISTS last_bucket AS
SELECT *
FROM bucket
ORDER BY created_at DESC
LIMIT 1;

CREATE TABLE IF NOT EXISTS measurement
(
    tm         REAL     NOT NULL,
    product_id INTEGER  NOT NULL,
    var        SMALLINT NOT NULL,
    value      REAL     NOT NULL,
    PRIMARY KEY (tm, product_id, var),
    FOREIGN KEY (product_id) REFERENCES product (product_id) ON DELETE CASCADE
);

CREATE TRIGGER IF NOT EXISTS trigger_bucket_insert
    AFTER INSERT
    ON measurement
    WHEN NOT EXISTS(SELECT created_at
                    FROM bucket) OR
         (new.tm - julianday((SELECT updated_at
                              FROM last_bucket))) * 86400. / 60. > 5
        OR (SELECT party_id
            FROM last_party) != (SELECT party_id
                                 FROM last_bucket)
BEGIN
    INSERT INTO bucket (created_at, updated_at, party_id)
    VALUES (datetime(new.tm), datetime(new.tm), (SELECT party_id FROM last_party));
END;

CREATE TRIGGER IF NOT EXISTS trigger_bucket_update
    AFTER INSERT
    ON measurement
    WHEN (new.tm - julianday((SELECT updated_at
                              FROM last_bucket))) * 86400. / 60. < 5
BEGIN
    UPDATE bucket
    SET updated_at = DATETIME(new.tm)
    WHERE bucket_id = (SELECT bucket_id FROM last_bucket);
END;

CREATE VIEW IF NOT EXISTS measurement_ext AS
SELECT STRFTIME('%Y-%m-%d %H:%M:%f', tm)   AS stored_at,
       cast(strftime('%Y', tm) AS INTEGER) AS year,
       cast(strftime('%m', tm) AS INTEGER) AS month,
       *
FROM measurement;

CREATE VIEW IF NOT EXISTS bucket_ext AS
SELECT bucket.*,
       party.created_at                                   AS party_created_at,
       cast(strftime('%Y', bucket.created_at) AS INTEGER) AS year,
       cast(strftime('%m', bucket.created_at) AS INTEGER) AS month,
       bucket_id = (SELECT bucket_id FROM last_bucket)    AS is_last
FROM bucket
         INNER JOIN party USING (party_id)
ORDER BY bucket.created_at;