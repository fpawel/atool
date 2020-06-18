package pkg

import (
	"database/sql"
	"github.com/ansel1/merry"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func OpenSqliteDB(fileName string) (*sql.DB, error) {
	conn, err := sql.Open("sqlite3", fileName)
	if err != nil {
		return nil, err
	}
	conn.SetMaxIdleConns(1)
	conn.SetMaxOpenConns(1)
	conn.SetConnMaxLifetime(0)
	return conn, err
}

func OpenSqliteDBx(fileName string) (*sqlx.DB, error) {
	conn, err := OpenSqliteDB(fileName)
	if err != nil {
		return nil, err
	}
	return sqlx.NewDb(conn, "sqlite3"), nil
}

func SqlGetNewInsertedID(r sql.Result) (int64, error) {
	id, err := r.LastInsertId()
	if err != nil {
		return 0, err
	}
	if id <= 0 {
		return 0, merry.New("was not inserted")
	}
	return id, nil
}
