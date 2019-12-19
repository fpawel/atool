package comports

import (
	"github.com/fpawel/comm/comport"
	"github.com/powerman/structlog"
	"sync"
	"time"
)

func GetComport(name string, baud int) (*comport.Port, error) {
	mu.Lock()
	defer mu.Unlock()

	c := comport.Config{
		Name:        name,
		Baud:        baud,
		ReadTimeout: time.Millisecond,
	}

	if p, f := ports[c.Name]; f {
		if err := p.SetConfig(c); err != nil {
			return nil, err
		}
		return p, nil
	}
	ports[c.Name] = comport.NewPort(c)
	return ports[c.Name], nil
}

func CloseComport(comportName string) {
	mu.Lock()
	defer mu.Unlock()
	if p, f := ports[comportName]; f {
		log.ErrIfFail(p.Close)
	}
}

func CloseAllComports() {
	mu.Lock()
	defer mu.Unlock()
	for _, port := range ports {
		log.ErrIfFail(port.Close)
	}
}

var (
	mu    sync.Mutex
	ports = make(map[string]*comport.Port)
	log   = structlog.New()
)
