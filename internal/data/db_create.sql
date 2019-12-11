PRAGMA foreign_keys = ON;
PRAGMA encoding = 'UTF-8';

CREATE TABLE IF NOT EXISTS party
(
    party_id   INTEGER PRIMARY KEY NOT NULL,
    created_at TIMESTAMP           NOT NULL DEFAULT (datetime('now')) UNIQUE,
    name       TEXT                NOT NULL DEFAULT '(без имени)'
);


CREATE TABLE IF NOT EXISTS product
(
    product_id INTEGER PRIMARY KEY NOT NULL,
    party_id   INTEGER             NOT NULL,
    serial     INTEGER             NOT NULL DEFAULT 0,
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
    id       INTEGER PRIMARY KEY NOT NULL,
    party_id INTEGER             NOT NULL,
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