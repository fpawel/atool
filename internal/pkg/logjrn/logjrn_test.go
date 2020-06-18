package logjrn

import (
	"fmt"
	"github.com/fpawel/atool/internal/pkg/must"
	"testing"
)

func TestJ_ListDays(t *testing.T) {
	j := new(J)
	must.PanicIf(j.Open(`C:\GOPATH\src\github.com\fpawel\atool\build\logjrn.sqlite`))
	days, err := j.ListDays()
	must.PanicIf(err)
	for _, day := range days {
		fmt.Println(day)
	}

	ents, err := j.GetEntriesOfDay(days[0])
	must.PanicIf(err)
	for _, ent := range ents {
		fmt.Println(ent)
	}

	must.PanicIf(j.db.Close())
}
