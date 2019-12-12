package app

import (
	"context"
	"github.com/fpawel/atool/internal/thriftgen/api"
	"sync/atomic"
)

type hardwareConnSvc struct{}

var _ api.HardwareConnectionService = new(hardwareConnSvc)

func (h *hardwareConnSvc) Connect(_ context.Context) error {
	if connected() {
		return nil
	}

	connect()
	return nil
}

func (h *hardwareConnSvc) Disconnect(_ context.Context) error {
	if !connected() {
		return nil
	}
	disconnect()
	wgConnect.Wait()
	return nil
}

func (h *hardwareConnSvc) Connected(_ context.Context) (bool, error) {
	return atomic.LoadInt32(&atomicConnected) != 0, nil
}
