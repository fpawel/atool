PRAGMA foreign_keys = ON;
PRAGMA encoding = 'UTF-8';

CREATE TABLE IF NOT EXISTS party
(
    party_id     INTEGER PRIMARY KEY NOT NULL,
    created_at   TIMESTAMP           NOT NULL DEFAULT (datetime('now')) UNIQUE,
    name         TEXT                NOT NULL DEFAULT '(без имени)',
    device_type  TEXT                NOT NULL DEFAULT '',
    product_type TEXT                NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS work_log
(
    record_id   INTEGER PRIMARY KEY NOT NULL,
    party_id    INTEGER             NOT NULL,
    started_at  TIMESTAMP           NOT NULL,
    complete_at TIMESTAMP DEFAULT NULL,
    work_name   TEXT                NOT NULL,
    CONSTRAINT product_party_id_foreign_key
        FOREIGN KEY (party_id) REFERENCES party (party_id)
            ON DELETE CASCADE
);

CREATE TRIGGER IF NOT EXISTS trigger_work_log_init_new_record
    AFTER INSERT
    ON work_log
BEGIN
    UPDATE work_log
    SET started_at = datetime('now'),
        party_id   = (SELECT app_config.party_id FROM app_config)
    WHERE record_id = new.record_id;
END;

CREATE TABLE IF NOT EXISTS product
(
    product_id    INTEGER PRIMARY KEY NOT NULL,
    party_id      INTEGER             NOT NULL,
    serial        INTEGER             NOT NULL DEFAULT 0,
    comport       TEXT                NOT NULL DEFAULT 'COM1',
    addr          INTEGER             NOT NULL DEFAULT 1,
    active        BOOLEAN             NOT NULL DEFAULT 1,
    created_at    DATETIME            NOT NULL DEFAULT (DATETIME('now', 'localtime')),
    created_order INTEGER             NOT NULL,
    CONSTRAINT product_party_id_comport_addr_unique
        UNIQUE (party_id, comport, addr),
    CONSTRAINT product_party_id_created_at_created_order_unique
        UNIQUE (party_id, created_at, created_order),
    CONSTRAINT product_party_id_foreign_key
        FOREIGN KEY (party_id) REFERENCES party (party_id)
            ON DELETE CASCADE,
    CONSTRAINT product_addr_more_then_1
        CHECK (addr >= 1),
    CONSTRAINT product_active_bool
        CHECK (active IN (0, 1) )
);

DROP VIEW IF EXISTS product_enumerated;
CREATE VIEW IF NOT EXISTS product_enumerated AS
SELECT *,
       (SELECT count()
        FROM product p
        WHERE p.party_id = a.party_id
          AND (p.created_at, p.created_order) < (a.created_at, a.created_order)) AS place
FROM product a;


CREATE TABLE IF NOT EXISTS party_value
(
    party_id INTEGER NOT NULL,
    key      TEXT    NOT NULL,
    value    REAL    NOT NULL,
    CONSTRAINT party_value_primary_key PRIMARY KEY (party_id, key),
    CONSTRAINT party_value_key_not_empty CHECK ( key != '' ),
    CONSTRAINT party_value_party_id_foreign_key
        FOREIGN KEY (party_id) REFERENCES party (party_id)
            ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS product_value
(
    product_id INTEGER NOT NULL,
    key        TEXT    NOT NULL,
    value      REAL    NOT NULL,
    CONSTRAINT product_value_primary_key PRIMARY KEY (product_id, key),
    CONSTRAINT product_value_key_not_empty CHECK ( key != '' ),
    CONSTRAINT product_value_party_id_foreign_key
        FOREIGN KEY (product_id) REFERENCES product (product_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS product_param
(
    product_id    INTEGER  NOT NULL,
    param_addr    SMALLINT NOT NULL,
    chart         TEXT     NOT NULL,
    series_active BOOLEAN  NOT NULL DEFAULT 1,

    CONSTRAINT product_param_primary_key
        PRIMARY KEY (product_id, param_addr),
    CONSTRAINT product_param_chart_not_empty
        CHECK (chart != '' ),
    CONSTRAINT product_param_series_active_bool
        CHECK (series_active IN (0, 1) ),
    CONSTRAINT product_param_product_id_foreign_key
        FOREIGN KEY (product_id) REFERENCES product (product_id)
            ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS measurement
(
    tm         INTEGER  NOT NULL,
    product_id INTEGER  NOT NULL,
    param_addr SMALLINT NOT NULL,
    value      REAL     NOT NULL,
    CONSTRAINT measurement_primary_key
        PRIMARY KEY (tm, product_id, param_addr),
    CONSTRAINT measurement_product_id_foreign_key
        FOREIGN KEY (product_id) REFERENCES product (product_id)
            ON DELETE CASCADE
);

CREATE VIEW IF NOT EXISTS measurement_ext AS
SELECT STRFTIME('%Y-%m-%d %H:%M:%f', tm / 1e9, 'unixepoch')   AS stored_at,
       cast(strftime('%Y', tm / 1e9, 'unixepoch') AS INTEGER) AS year,
       cast(strftime('%m', tm / 1e9, 'unixepoch') AS INTEGER) AS month,
       *
FROM measurement;

CREATE TABLE IF NOT EXISTS app_config
(
    id       INTEGER PRIMARY KEY NOT NULL,
    party_id INTEGER             NOT NULL,
    CONSTRAINT app_config_party_id_foreign_key
        FOREIGN KEY (party_id) REFERENCES party (party_id)
);

INSERT INTO party (party_id)
SELECT 1
WHERE NOT EXISTS(SELECT * FROM party);

INSERT INTO product (product_id, party_id, created_order)
SELECT 1, 1, 1
WHERE NOT EXISTS(SELECT * FROM product);

INSERT OR IGNORE INTO app_config (id, party_id)
VALUES (1, 1);