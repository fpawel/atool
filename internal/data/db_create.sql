PRAGMA foreign_keys = ON;
PRAGMA encoding = 'UTF-8';

CREATE TABLE IF NOT EXISTS party
(
    party_id   INTEGER PRIMARY KEY NOT NULL,
    created_at TIMESTAMP           NOT NULL DEFAULT (datetime('now')) UNIQUE,
    note       TEXT                NOT NULL DEFAULT '(без примечания)'
);

CREATE TABLE IF NOT EXISTS hardware
(
    device                TEXT    NOT NULL PRIMARY KEY,
    baud                  INTEGER NOT NULL DEFAULT 9600,
    timeout_get_responses INTEGER NOT NULL DEFAULT 1000000000, -- в наносекундах = 1 секунда
    timeout_end_response  INTEGER NOT NULL DEFAULT 50000000,   -- в наносекундах = 50 миллисекунд
    pause                 INTEGER NOT NULL DEFAULT 0,          -- в наносекундах
    max_attempts_read     INTEGER NOT NULL DEFAULT 0,
    CHECK ( timeout_get_responses >= 0 ),
    CHECK ( timeout_end_response >= 0 ),
    CHECK ( pause >= 0 ),
    CHECK ( baud >= 0 ),
    CHECK ( max_attempts_read >= 0 )
);

INSERT OR IGNORE INTO hardware(device)
VALUES ('DEFAULT');

CREATE TABLE IF NOT EXISTS param
(
    device TEXT    NOT NULL DEFAULT 'DEFAULT',
    var    INTEGER NOT NULL,
    count  INTEGER NOT NULL,
    format TEXT    NOT NULL,
    FOREIGN KEY (device) REFERENCES hardware (device) ON DELETE CASCADE,
    PRIMARY KEY (device, var),
    CHECK (var >= 0 ),
    CHECK (count >= 0 ),
    CHECK (format IN ('bcd', 'float', 'int'))
);

INSERT OR IGNORE INTO param(device, var, count, format)
VALUES ('DEFAULT', 0, 2, 'bcd');

CREATE TABLE IF NOT EXISTS product
(
    product_id INTEGER   NOT NULL,
    party_id   INTEGER   NOT NULL,
    device     TEXT      NOT NULL DEFAULT 'DEFAULT',
    comport    TEXT      NOT NULL DEFAULT 'COM1',
    addr       INTEGER   NOT NULL DEFAULT 1,
    checked    BOOLEAN   NOT NULL DEFAULT 1,
    PRIMARY KEY (product_id),
    UNIQUE (party_id, comport, addr),
    FOREIGN KEY (party_id) REFERENCES party (party_id) ON DELETE CASCADE,
    FOREIGN KEY (device) REFERENCES hardware (device) ON DELETE CASCADE,
    CHECK (addr >= 1 ),
    CHECK (checked IN (0, 1) )
);

CREATE TABLE IF NOT EXISTS chart
(
    product_id INTEGER NOT NULL,
    var        INTEGER NOT NULL,
    chart_id   INTEGER NOT NULL DEFAULT 1,
    checked    BOOLEAN NOT NULL DEFAULT 1,
    left_axis  BOOLEAN NOT NULL DEFAULT 1,
    color      TEXT    NOT NULL DEFAULT '',
    UNIQUE (product_id, var),
    FOREIGN KEY (product_id) REFERENCES product (product_id) ON DELETE CASCADE,
    CHECK (chart_id >= 1 ),
    CHECK (checked IN (0, 1)),
    CHECK (left_axis IN (0, 1))
);



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
    tm         REAL    NOT NULL,
    product_id INTEGER NOT NULL,
    var        INTEGER NOT NULL,
    value      REAL    NOT NULL,
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
            FROM app_config) != (SELECT party_id
                                 FROM last_bucket)
BEGIN
    INSERT INTO bucket (created_at, updated_at, party_id)
    VALUES (datetime(new.tm), datetime(new.tm), (SELECT party_id FROM app_config));
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

CREATE TABLE IF NOT EXISTS app_config
(
    id          INTEGER PRIMARY KEY NOT NULL,
    party_id    INTEGER             NOT NULL,
    log_comport BOOLEAN             NOT NULL DEFAULT 0,
    FOREIGN KEY (party_id) REFERENCES party (party_id)
);

INSERT INTO party (party_id)
SELECT 1
WHERE NOT EXISTS(SELECT * FROM party);

INSERT INTO product (product_id, party_id)
SELECT 1, 1
WHERE NOT EXISTS(SELECT * FROM product);

INSERT OR IGNORE INTO app_config (id, party_id)
VALUES (1, 1);