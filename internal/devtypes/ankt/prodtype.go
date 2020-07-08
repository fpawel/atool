package ankt

import (
	"fmt"
)

type productType struct {
	N     int
	Chan2 bool
	Press bool
	Chan  [2]chanNfo
}

func (x productType) String() string {
	if !x.Chan2 {
		return fmt.Sprintf("%d %s", x.N, x.Chan[0])
	}
	return fmt.Sprintf("%d %s %s", x.N, x.Chan[0], x.Chan[1])
}
