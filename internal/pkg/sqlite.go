package pkg

import (
	"database/sql"
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


func OpenSqliteDBx(fileName string) (*sqlx.DB,error) {
	conn, err := OpenSqliteDB(fileName)
	if err != nil {
		return nil,err
	}
	return sqlx.NewDb(conn, "sqlite3"), nil
}