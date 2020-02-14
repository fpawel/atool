package data

import (
	"database/sql"
	"fmt"
	"github.com/ansel1/merry"
	"strconv"
	"strings"
	"time"
)

const TimeLayout = "2006-01-02 15:04:05.000"

func formatMonth(t time.Time) string {
	months := []string{
		"Январь",
		"Февраль",
		"Март",
		"Апрель",
		"Май",
		"Июнь",
		"Июль",
		"Август",
		"Сентябрь",
		"Октябрь",
		"Ноябрь",
		"Декабрь",
	}
	n := int(t.Month())
	if n > 1 && n < 12 {
		return months[n-1]
	}
	return ""
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
