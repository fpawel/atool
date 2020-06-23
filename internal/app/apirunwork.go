package app

import (
	"context"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/atool/internal/config/appcfg"
	"github.com/fpawel/atool/internal/data"
	"github.com/fpawel/atool/internal/hardware"
	"github.com/fpawel/atool/internal/thriftgen/api"
	"github.com/fpawel/atool/internal/workgui"
	"github.com/fpawel/atool/internal/worklua"
	"github.com/fpawel/atool/internal/workparty"
	"github.com/fpawel/comm"
	"github.com/fpawel/comm/modbus"
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"
	"path/filepath"
)

type runWorkSvc struct{}

var _ api.RunWorkService = new(runWorkSvc)

func (h *runWorkSvc) RunLuaScript(_ context.Context, filename string) error {
	luaState := lua.NewState()
	imp := worklua.NewImport(log, luaState)
	luaState.SetGlobal("go", luar.New(luaState, imp))
	return runWorkFunc(filepath.Base(filename), func(log logger, ctx context.Context) error {
		defer hardware.CloseHardware(log, appCtx)
		defer luaState.Close()
		luaState.SetContext(ctx)
		return luaState.DoFile(filename)
	})
}

func (h *runWorkSvc) RunDeviceWork(context.Context) error {
	p, err := data.GetCurrentParty()
	if err != nil {
		return err
	}
	d, ok := appcfg.DeviceTypes[p.DeviceType]
	if !ok {
		return fmt.Errorf("не задан тип прибора: %s", p.DeviceType)
	}
	if d.Work == nil {
		return fmt.Errorf("тип прибора %s не поддерживает автоматическую настройку", p.DeviceType)
	}
	return runWorkFunc(
		"Автоматическая настройка: "+p.DeviceType,
		func(log comm.Logger, ctx context.Context) error {
			err := d.Work(log, ctx)
			hardware.CloseHardware(log, appCtx)
			return err
		},
	)
}

func (h *runWorkSvc) SearchProducts(ctx context.Context, comportName string) error {
	return workparty.NewWorkScanModbus(comportName).Run(log, ctx)
}

func (h *runWorkSvc) Connect(_ context.Context) error {
	return runWork(workparty.NewWorkInterrogate())
}

func (h *runWorkSvc) Interrupt(_ context.Context) error {
	if !workgui.IsConnected() {
		return nil
	}
	workgui.Interrupt()
	workgui.Wait()
	return nil
}

func (h *runWorkSvc) InterruptDelay(_ context.Context) error {
	workgui.InterruptDelay(log)
	return nil
}

func (h *runWorkSvc) Connected(_ context.Context) (bool, error) {
	return workgui.IsConnected(), nil
}

func (h *runWorkSvc) Command(_ context.Context, cmd int16, s string) error {
	b, err := parseHexBytes(s)
	if err != nil {
		return merry.New("ожидалась последовательность байтов HEX")
	}
	return runWork(workparty.NewWorkRawCmd(modbus.ProtoCmd(cmd), b))
}

func (h *runWorkSvc) SwitchGas(_ context.Context, valve int8) error {
	return runWorkFunc(
		fmt.Sprintf("переключение пневмоблока %d", valve),
		hardware.SwitchGas(byte(valve)))
}
