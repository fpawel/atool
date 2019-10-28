package must

import (
	"database/sql"
	"encoding/json"
	"github.com/fpawel/atool/internal/pkg"
	"github.com/jmoiron/sqlx"
	"github.com/pelletier/go-toml"
	"io/ioutil"
	"os"
)

// PanicIf will call panic(err) in case given err is not nil.
func PanicIf(err error) {
	if err != nil {
		panic(err)
	}
}

func OpenSqliteDB(fileName string) *sql.DB {
	conn, err := pkg.OpenSqliteDB(fileName)
	PanicIf(err)
	return conn
}

func OpenSqliteDBx(fileName string) *sqlx.DB {
	return sqlx.NewDb(OpenSqliteDB(fileName), "sqlite3")
}

// WriteFile is a wrapper for ioutil.WriteFile.
func WriteFile(name string, buf []byte, perm os.FileMode) {
	err := ioutil.WriteFile(name, buf, perm)
	PanicIf(err)
}

// UnmarshalJSON is a wrapper for json.Unmarshal.
func UnmarshalJSON(data []byte, v interface{}) {
	err := json.Unmarshal(data, v)
	PanicIf(err)
}

// MarshalToml is a wrapper for toml.Marshal.
func MarshalToml(v interface{}) []byte {
	data, err := toml.Marshal(v)
	PanicIf(err)
	return data
}

// MarshalJSON is a wrapper for json.Marshal.
func MarshalJSON(v interface{}) []byte {
	data, err := json.Marshal(v)
	PanicIf(err)
	return data
}
