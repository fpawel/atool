package comports

import (
	"github.com/fpawel/atool/internal/config/devicecfg"
	"github.com/fpawel/comm"
	"github.com/fpawel/comm/comport"
	"github.com/powerman/structlog"
	"sync"
	"time"
)

//func LookupComport(name string) *comport.Port {
//	mu.Lock()
//	defer mu.Unlock()
//	p, _ := ports[name]
//	return p
//}

func Comm(comportName string, device devicecfg.Device) comm.T {
	return comm.New(Comport(comportName, device.Baud), device.CommConfig()).WithLockPort(comportName)
}

func Comport(name string, baud int) *comport.Port {
	mu.Lock()
	defer mu.Unlock()
	c := comport.Config{
		Name:        name,
		Baud:        baud,
		ReadTimeout: time.Millisecond,
	}
	if p, f := ports[c.Name]; f {
		p.SetConfig(structlog.New(), c)
	} else {
		ports[c.Name] = comport.NewPort(c)
	}
	return ports[c.Name]
}

func CloseComport(comportName string) {
	mu.Lock()
	defer mu.Unlock()
	if p, f := ports[comportName]; f {
		log.ErrIfFail(p.Close, "close_comport", comportName)
	}
}

func CloseAllComports() {
	mu.Lock()
	defer mu.Unlock()
	for comportName, port := range ports {
		log.ErrIfFail(port.Close, "close_comport", comportName)
	}
}

var (
	mu    sync.Mutex
	ports = make(map[string]*comport.Port)
	log   = structlog.New()
)
