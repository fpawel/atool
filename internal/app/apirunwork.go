package app

import (
	"context"
	"fmt"
	"github.com/fpawel/atool/internal/config/appcfg"
	"github.com/fpawel/atool/internal/data"
	"github.com/fpawel/atool/internal/hardware"
	"github.com/fpawel/atool/internal/thriftgen/api"
	"github.com/fpawel/atool/internal/thriftgen/apitypes"
	"github.com/fpawel/atool/internal/workgui"
	"github.com/fpawel/atool/internal/worklua"
	"github.com/fpawel/atool/internal/workparty"
	"github.com/fpawel/comm"
	"github.com/fpawel/comm/modbus"
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"
	"path/filepath"
	"strconv"
	"strings"
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

func (h *runWorkSvc) SendDeviceCommand(_ context.Context, req *apitypes.RequestDeviceCommand) error {
	devCmd, ok := parseDevCmdCode(req.CmdDevice)
	if !ok {
		return fmt.Errorf("invalid device command %q", req.CmdDevice)
	}
	errInvalidArg := func(err error) error {
		return fmt.Errorf("invalid argument %s %q: %w", req.Format, req.Argument, err)
	}
	strFormat := strings.ToLower(req.Format)
	if strFormat == "hex" {
		dtBytes, err := parseHexBytes(req.Argument)
		if err != nil {
			return errInvalidArg(err)
		}
		if len(dtBytes) != 4 {
			return errInvalidArg(fmt.Errorf("ожидалось 4 байта данных, получено %d", len(dtBytes)))
		}
		runSingleTask(workparty.NewWorkWrite32Bytes(modbus.ProtoCmd(req.CmdModbus), devCmd, dtBytes).Func)
		return nil
	}
	fFmt := modbus.FloatBitsFormat(strFormat)
	if err := fFmt.Validate(); err != nil {
		return errInvalidArg(err)
	}
	v, err := strconv.ParseFloat(req.Argument, 64)
	if err != nil {
		return errInvalidArg(err)
	}
	runSingleTask(workparty.Write32(devCmd, fFmt, v))
	return nil
}

func parseDevCmdCode(name string) (modbus.DevCmd, bool) {
	n, err := strconv.ParseUint(name, 10, 16)
	if err == nil {
		return modbus.DevCmd(n), true
	}
	party, err := data.GetCurrentParty()
	if err != nil {
		return 0, false
	}
	device, err := appcfg.GetDeviceByName(party.DeviceType)
	if err != nil {
		return 0, false
	}

	for _, c := range device.Commands {
		if c.Name == name {
			return c.Code, true
		}
	}
	return 0, false
}

func (h *runWorkSvc) SwitchGas(_ context.Context, valve int8) error {
	runSingleTask(hardware.SwitchGas(byte(valve)))
	return nil
}
