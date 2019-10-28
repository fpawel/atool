PRAGMA foreign_keys = ON;
PRAGMA encoding = 'UTF-8';

CREATE TABLE IF NOT EXISTS party
(
    party_id   INTEGER PRIMARY KEY NOT NULL,
    created_at TIMESTAMP           NOT NULL DEFAULT (datetime('now')) UNIQUE
);

CREATE TABLE IF NOT EXISTS device
(
    device_name TEXT NOT NULL PRIMARY KEY
);

INSERT OR IGNORE INTO device(device_name)
VALUES ('DEFAULT');

CREATE TABLE IF NOT EXISTS device_var
(
    device_name TEXT     NOT NULL,
    var_addr    SMALLINT NOT NULL,
    name        TEXT     NOT NULL,
    param_type  TEXT     NOT NULL CHECK (param_type IN
                                          ('bcd', 'float', 'uint16_little_endian', 'uint16_big_endian')),
    PRIMARY KEY (device_name, var_addr),
    FOREIGN KEY (device_name) REFERENCES device (device_name) ON DELETE CASCADE
);

INSERT OR IGNORE INTO device_var(device_name, var_addr, name, param_type)
VALUES ('DEFAULT', 0, 'Концентрация', 'bcd');

CREATE TABLE IF NOT EXISTS interrogate
(
    device_name TEXT     NOT NULL,
    var_addr    SMALLINT NOT NULL,
    count       SMALLINT NOT NULL,
    checked     BOOLEAN  NOT NULL DEFAULT 1 CHECK ( checked IN (0, 1) ),
    PRIMARY KEY (device_name, var_addr),
    FOREIGN KEY (device_name) REFERENCES device (device_name) ON DELETE CASCADE
);

INSERT OR IGNORE INTO interrogate(device_name, var_addr, count)
VALUES ('DEFAULT', 0, 2);

CREATE TABLE IF NOT EXISTS product
(
    product_id  INTEGER PRIMARY KEY NOT NULL,
    party_id    INTEGER             NOT NULL,
    created_at  TIMESTAMP           NOT NULL DEFAULT (datetime('now')) UNIQUE,
    device_name TEXT                NOT NULL DEFAULT 'DEFAULT',
    serial      INTEGER             NOT NULL CHECK (serial > 0 ),
    port        INTEGER             NOT NULL DEFAULT 1 CHECK (port >= 0 ),
    addr        INTEGER             NOT NULL DEFAULT 1 CHECK (addr >= 1 ),
    checked     BOOLEAN             NOT NULL DEFAULT 1 CHECK ( checked IN (0, 1) ),
    UNIQUE (party_id, port, addr),
    UNIQUE (party_id, device_name, serial),
    FOREIGN KEY (party_id) REFERENCES party (party_id) ON DELETE CASCADE,
    FOREIGN KEY (device_name) REFERENCES device (device_name) ON DELETE CASCADE
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
    var_addr   SMALLINT NOT NULL,
    value      REAL     NOT NULL,
    PRIMARY KEY (tm, product_id, var_addr),
    FOREIGN KEY (product_id) REFERENCES product (product_id) ON DELETE CASCADE,
    FOREIGN KEY (var_addr) REFERENCES device_var (var_addr) ON DELETE CASCADE
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