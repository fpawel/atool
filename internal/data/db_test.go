package data

import (
	"github.com/fpawel/atool/internal/pkg/must"
	"os"
	"path/filepath"
	"testing"
)

func TestCopyCurrentParty(t *testing.T) {
	dbFilename := filepath.Join(os.Getenv("GOPATH"), "src", "github.com", "fpawel", "atool", "build", "atool.sqlite")
	db, err := Open(dbFilename)
	must.PanicIf(err)
	must.PanicIf(CopyCurrentParty(db))
	must.PanicIf(db.Close())
}
