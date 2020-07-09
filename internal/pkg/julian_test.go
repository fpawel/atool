package pkg

import (
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

func Test1(t *testing.T) {
	fmt.Println(time.Now().Format(time.RFC3339))
}

func TestJulian(t *testing.T) {

	for i := 0; i < 1_000; i++ {
		tm := randDate()
		tmJulian := TimeToJulian(tm)
		tmReverse := JulianToTime(tmJulian)
		testTimeDiff(t, tm, tmReverse, 100*time.Microsecond)
	}

	for i := 0; i < 1_000; i++ {
		tm := randDate()
		tmJulian := TimeToJulian(tm)
		tmJulianSql := timeToJulianSql(tm)
		tmReverseSql := julianToTimeSql(tmJulianSql)
		assert.Equal(t, tmJulian, tmJulianSql, "julian should be the same as sql")
		testTimeDiff(t, tm, tmReverseSql, 100*time.Microsecond)
	}
}

func randDate() time.Time {
	min := time.Date(1970, 1, 0, 0, 0, 0, 0, time.UTC).Unix()
	max := time.Date(2070, 1, 0, 0, 0, 0, 0, time.UTC).Unix()
	delta := max - min

	sec := rand.Int63n(delta) + min
	return time.Unix(sec, 0)
}

func testTimeDiff(t *testing.T, t1, t2 time.Time, maxDiff time.Duration) {
	assert.Less(t, int64(t1.Sub(t2)), int64(maxDiff), fmt.Sprintf("%s AND %s", t1, t2))
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
	tm, err := time.ParseInLocation(timeLayout, s, time.Local)
	if err != nil {
		panic(err)
	}
	return tm
}

var db *sqlx.DB

func init() {
	rand.Seed(time.Now().Unix())
	conn, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		panic(err)
	}
	db = sqlx.NewDb(conn, "sqlite3")
}
