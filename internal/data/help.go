package data

import (
	"database/sql"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/jmoiron/sqlx"
	"strconv"
	"strings"
	"time"
)

const TimeLayout = "2006-01-02 15:04:05.000"

func formatTimeAsQuery(t time.Time) string {
	return "julianday(STRFTIME('%Y-%m-%d %H:%M:%f','" +
		t.Format(TimeLayout) + "'))"
}

func getNewInsertedID(r sql.Result) (int64, error) {
	id, err := r.LastInsertId()
	if err != nil {
		return 0, err
	}
	if id <= 0 {
		return 0, merry.New("was not inserted")
	}
	return id, nil
}

func parseTime(sqlStr string) time.Time {
	t, err := time.ParseInLocation(TimeLayout, sqlStr, time.Now().Location())
	if err != nil {
		panic(err)
	}
	return t
}

func formatIDs(ids []int64) string {
	var ss []string
	for _, id := range ids {
		ss = append(ss, strconv.FormatInt(id, 10))
	}
	return strings.Join(ss, ",")
}

func partyProductsIDsSql(db *sqlx.DB, partyID int64) (string, error) {
	var xs []int64
	if err := db.Select(&xs, `SELECT product_id FROM product WHERE party_id = ?`, partyID); err != nil {
		return "", err
	}
	return "(" + formatIDs(xs) + ")", nil
}

func formatIntSliceAsQuery(xs []int) string {
	var sx []string
	for _, n := range xs {
		sx = append(sx, strconv.Itoa(n))
	}
	return formatStrSliceAsQuery(sx)
}

func formatInt64SliceAsQuery(xs []int64) string {
	var sx []string
	for _, x := range xs {
		sx = append(sx, strconv.FormatInt(x, 10))
	}
	return formatStrSliceAsQuery(sx)
}

func formatStrSliceAsQuery(sx []string) string {
	return fmt.Sprintf("(%s)", strings.Join(sx, ","))
}
