package data

import (
	"database/sql"
	"github.com/fpawel/atool/internal/pkg"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
	"time"
)

func TestJulian(t *testing.T) {

	tm := parseTimeLayout(time.Now().Format(timeLayout))
	tmJulian := pkg.TimeToJulian(tm)
	tmReverse := pkg.JulianToTime(tmJulian)

	tmJulianSql := timeToJulianSql(tm)
	tmReverseSql := julianToTimeSql(tmJulianSql)

	tmMs := millis(tm)
	tmReverseMs := millis(tmReverse)
	tmReverseSqlMs := millis(tmReverseSql)

	assert.Equal(t, tmJulian, tmJulianSql, "julian should be the same as sql")
	assert.Equal(t, tmMs, tmReverseMs, "reverse should be the same")
	assert.Equal(t, tmMs, tmReverseSqlMs, "sql: reverse should be the same")
}

func millis(t time.Time) int64 {
	return t.UnixNano() / 1_000_000_000
}

func parseTimeLayout(s string) time.Time {
	tm, err := time.ParseInLocation(timeLayout, s, time.Local)
	if err != nil {
		panic(err)
	}
	return tm
}

const timeLayout = "2006-01-02 15:04:05.000"

func timeToJulianSql(t time.Time) (x float64) {
	err := db.Get(&x, "SELECT julianday(STRFTIME('%Y-%m-%d %H:%M:%f','"+t.Format(timeLayout)+"'))")
	if err != nil {
		panic(err)
	}
	return
}

func julianToTimeSql(j float64) time.Time {
	var s string
	err := db.Get(&s, "SELECT STRFTIME('%Y-%m-%d %H:%M:%f',"+strconv.FormatFloat(j, 'g', -1, 64)+")")
	if err != nil {
		panic(err)
	}
	return parseTimeLayout(s)
}

var db *sqlx.DB

func init() {
	conn, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		panic(err)
	}
	conn.SetMaxIdleConns(1)
	conn.SetMaxOpenConns(1)
	conn.SetConnMaxLifetime(0)
	db = sqlx.NewDb(conn, "sqlite3")
}
