package logjrn

import (
	"github.com/fpawel/atool/internal/pkg"
	"github.com/jmoiron/sqlx"
)

//go:generate go run github.com/fpawel/gotools/cmd/sqlstr/...

type J struct {
	db *sqlx.DB
}

func (x *J) Open(filename string) error {
	var err error
	x.db, err = pkg.OpenSqliteDBx(filename)
	if err != nil {
		return err
	}
	if _, err := x.db.Exec(SQLSchema); err != nil {
		return err
	}
	return nil
}
