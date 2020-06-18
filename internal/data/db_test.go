package data

import (
	"github.com/fpawel/atool/internal/pkg/must"
	"os"
	"path/filepath"
	"testing"
)

func TestCopyCurrentParty(t *testing.T) {
	dbFilename := filepath.Join(os.Getenv("GOPATH"), "src", "github.com", "fpawel", "atool", "build", "atool.sqlite")
	err := Open(dbFilename)
	must.PanicIf(err)
	partyID, err := GetCurrentPartyID()
	must.PanicIf(err)
	must.PanicIf(CopyParty(partyID))
	must.PanicIf(DB.Close())
}
