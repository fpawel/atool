package data

import (
	"database/sql"
	"github.com/ansel1/merry"
	"github.com/jmoiron/sqlx"
	"strconv"
	"strings"
	"time"
)

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

func currentPartyProductsIDsSql(db *sqlx.DB) (string, error) {
	var xs []int64
	if err := db.Select(&xs, `SELECT product_id FROM product WHERE party_id = (SELECT party_id FROM app_config)`); err != nil {
		return "", err
	}
	return "(" + formatIDs(xs) + ")", nil
}
