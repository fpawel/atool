package must

import (
	"database/sql"
	"encoding/json"
	"github.com/fpawel/atool/internal/pkg"
	"github.com/jmoiron/sqlx"
	"gopkg.in/yaml.v3"
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

// UnmarshalYAML is a wrapper for json.Unmarshal.
func UnmarshalYaml(data []byte, v interface{}) {
	err := yaml.Unmarshal(data, v)
	PanicIf(err)
}

// MarshalYaml is a wrapper for toml.Marshal.
func MarshalYaml(v interface{}) []byte {
	data, err := yaml.Marshal(v)
	PanicIf(err)
	return data
}

func MarshalJson(v interface{}) []byte {
	data, err := json.Marshal(v)
	PanicIf(err)
	return data
}
