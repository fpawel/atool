PRAGMA foreign_keys = ON;
PRAGMA encoding = 'UTF-8';

CREATE TABLE IF NOT EXISTS party
(
    party_id   INTEGER PRIMARY KEY NOT NULL,
    created_at TIMESTAMP           NOT NULL DEFAULT (datetime('now')) UNIQUE
);

CREATE TABLE IF NOT EXISTS hardware
(
    device                TEXT    NOT NULL PRIMARY KEY,
    baud                  INTEGER NOT NULL DEFAULT 9600,
    timeout_get_responses INTEGER NOT NULL DEFAULT 1000000000, -- в наносекундах = 1 секунда
    timeout_end_response  INTEGER NOT NULL DEFAULT 50000000,   -- в наносекундах = 50 миллисекунд
    pause                 INTEGER NOT NULL DEFAULT 0,          -- в наносекундах
    max_attempts_read     INTEGER NOT NULL DEFAULT 0,
    CHECK (timeout_get_responses >= 0),
    CHECK (timeout_end_response >= 0),
    CHECK (pause >= 0),
    CHECK (baud >= 0),
    CHECK (max_attempts_read >= 0 )
);

CREATE TABLE IF NOT EXISTS params
(
    device   TEXT     NOT NULL,
    reg_addr SMALLINT NOT NULL,
    format   TEXT     NOT NULL,
    count    SMALLINT NOT NULL,
    PRIMARY KEY (device, reg_addr),
    FOREIGN KEY (device) REFERENCES hardware (device) ON DELETE CASCADE,
    CHECK (reg_addr >= 0 ),
    CHECK (count > 0 ),
    CHECK (format IN ('bcd',
                      'float_big_endian', 'float_little_endian',
                      'int_big_endian', 'int_little_endian'))
);

INSERT OR IGNORE INTO hardware(device)
VALUES ('default');
INSERT OR IGNORE INTO params(device, reg_addr, format, count)
VALUES ('default', 0, 'bcd', 1);

CREATE TABLE IF NOT EXISTS product
(
    product_id INTEGER PRIMARY KEY NOT NULL,
    party_id   INTEGER             NOT NULL,
    device     TEXT                NOT NULL DEFAULT 'default',
    comport    TEXT                NOT NULL DEFAULT 'COM1',
    addr       INTEGER             NOT NULL DEFAULT 1,
    active     BOOLEAN             NOT NULL DEFAULT 1,
    UNIQUE (party_id, comport, addr),
    FOREIGN KEY (party_id) REFERENCES party (party_id)
        ON DELETE CASCADE,
    CHECK (addr >= 1),
    CHECK (active IN (0, 1) )
);

CREATE TABLE IF NOT EXISTS product_param
(
    product_id    INTEGER  NOT NULL,
    param_addr    SMALLINT NOT NULL,
    chart         TEXT     NOT NULL,
    series_active BOOLEAN  NOT NULL DEFAULT 1,

    PRIMARY KEY (product_id, param_addr),
    CHECK (chart != '' ),
    CHECK (series_active IN (0, 1) ),

    FOREIGN KEY (product_id) REFERENCES product (product_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS measurement
(
    tm         REAL     NOT NULL,
    product_id INTEGER  NOT NULL,
    param_addr SMALLINT NOT NULL,
    value      REAL     NOT NULL,
    PRIMARY KEY (tm, product_id, param_addr),
    FOREIGN KEY (product_id) REFERENCES product (product_id) ON DELETE CASCADE
);

CREATE VIEW IF NOT EXISTS measurement_ext AS
SELECT STRFTIME('%Y-%m-%d %H:%M:%f', tm)   AS stored_at,
       cast(strftime('%Y', tm) AS INTEGER) AS year,
       cast(strftime('%m', tm) AS INTEGER) AS month,
       *
FROM measurement;

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