package logjrn

import (
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/internal/pkg"
	"github.com/jmoiron/sqlx"
	"io/ioutil"
	"strings"
	"time"
)

//go:generate go run github.com/fpawel/gotools/cmd/sqlstr/...

type J struct {
	db *sqlx.DB
}

type Entry struct {
	EntryID  int64     `db:"entry_id"`
	StoredAt time.Time `db:"stored_at"`
	Text     string    `db:"text"`
	Ok       bool      `db:"ok"`
	Indent   int       `db:"indent"`
	Stack    string    `db:"stack"`
}

func (x *J) Close() error {
	return x.db.Close()
}

func (x *J) Open(filename string) error {
	db, err := pkg.OpenSqliteDBx(filename)
	if err != nil {
		return merry.Wrap(err)
	}
	if _, err := db.Exec(SQLSchema); err != nil {
		return merry.Wrap(err)
	}
	x.db = db
	return nil
}

func (x *J) GetEntryByID(ent *Entry) error {
	return x.db.Get(ent, `
SELECT entry_id, stored_at, ok, text, indent, stack 
FROM entry 
WHERE entry_id = ? `, ent.EntryID)
}

func (x *J) GetEntriesIDsOfDay(strTm string) ([]int64, error) {
	t, err := time.Parse(layoutDate, strTm)
	if err != nil {
		return nil, err
	}
	strTm = t.Format(layoutDate)
	var r []int64
	err = x.db.Select(&r, `SELECT entry_id FROM entry WHERE STRFTIME('%Y.%m.%d', stored_at) = ? `, strTm)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (x *J) GetEntriesOfDay(strTm string) ([]*Entry, error) {
	t, err := time.Parse(layoutDate, strTm)
	if err != nil {
		return nil, err
	}
	strTm = t.Format(layoutDate)
	var r []*Entry
	err = x.db.Select(&r, `
SELECT entry_id, stored_at, ok, text, indent, stack 
FROM entry 
WHERE STRFTIME('%Y.%m.%d', stored_at) = ? `, strTm)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (x *J) DeleteDays(days []string) error {
	var xs []string
	for _, day := range days {
		xs = append(xs, fmt.Sprintf("'%s'", day))
	}
	query := `DELETE FROM entry WHERE STRFTIME('%Y.%m.%d', stored_at) IN ` + fmt.Sprintf("(%s)", strings.Join(xs, ","))
	_, err := x.db.Exec(query)
	return err
}

func (x *J) ListDays() ([]string, error) {
	var days []string
	const q = `
SELECT DISTINCT strftime('%Y.%m.%d', stored_at)
FROM entry
ORDER BY stored_at DESC`

	if err := x.db.Select(&days, q); err != nil {
		return nil, err
	}

	var f bool
	sNow := time.Now().Format(layoutDate)
	for _, s := range days {
		if s == sNow {
			f = true
		}
	}
	if !f {
		days = append([]string{sNow}, days...)
	}
	return days, nil
}

func (x *J) AddEntry(ent *Entry) error {
	r, err := x.db.Exec(
		`INSERT INTO entry(stored_at, ok, text, indent, stack) VALUES ( STRFTIME('%Y-%m-%d %H:%M:%f',?), ?,?,?,?)`,
		ent.StoredAt.Format(timeLayout),
		ent.Ok,
		ent.Text,
		ent.Indent,
		ent.Stack)
	if err != nil {
		return merry.Wrap(err)
	}
	n, err := r.RowsAffected()
	if err != nil {
		return merry.Wrap(err)
	}
	if n != 1 {
		return merry.Errorf("expected 1 rows affected, got %d", n)
	}
	if ent.EntryID, err = pkg.SqlGetNewInsertedID(r); err != nil {
		return err
	}
	return nil
}

func (x *J) AddEntries(ents []*Entry) error {
	var xs []string
	for _, ent := range ents {
		ok := 0
		if ent.Ok {
			ok = 1
		}
		xs = append(xs, fmt.Sprintf("(%s, %d, '%s', %d, '%s')",
			formatTimeAsQuery(ent.StoredAt),
			ok,
			removeQuote(ent.Text), ent.Indent,
			removeQuote(ent.Stack)))
	}
	q := `INSERT INTO entry(stored_at, ok, text, indent, stack) VALUES ` + strings.Join(xs, ",")

	r, err := x.db.Exec(q)
	if err != nil {
		_ = ioutil.WriteFile("sql.sql", []byte(q), 0644)
		return merry.Wrap(err)
	}
	n, err := r.RowsAffected()
	if err != nil {
		return merry.Wrap(err)
	}
	if n != int64(len(ents)) {
		return merry.Errorf("expected 1 rows affected, got %d", n)
	}
	return nil
}

const layoutDate = "2006.01.02"

func removeQuote(value string) string {
	replace := map[string]string{"'": ""}

	for b, a := range replace {
		value = strings.Replace(value, b, a, -1)
	}

	return value
}

func formatTimeAsQuery(t time.Time) string {
	return "STRFTIME('%Y-%m-%d %H:%M:%f','" +
		t.Format(timeLayout) + "')"
}

const timeLayout = "2006-01-02 15:04:05.000"
