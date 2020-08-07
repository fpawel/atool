package ankt

import (
	"github.com/fpawel/atool/internal/config/devicecfg"
	"github.com/fpawel/atool/internal/devtypes/ankt/anktvar"
	"github.com/fpawel/atool/internal/devtypes/devdata"
	"github.com/fpawel/comm/modbus"
	"sort"
	"time"
)

var (
	Device = devdata.Device{
		Name:   deviceName,
		Config: deviceConfig,
		ProductTypes: func() (xs []string) {
			for _, t := range productTypesList {
				xs = append(xs, t.String())
			}
			sort.Strings(xs)
			return
		}(),
		PartyParams: []devdata.PartyParam{
			{
				Key:  "c1",
				Name: "концентрация ПГС1: начало шк.",
			},
			{
				Key:  "c2",
				Name: "концентрация ПГС2: середина шк.к.1",
			},
			{
				Key:  "c3",
				Name: "концентрация ПГС3: доп.середина шк.к.1 CO₂",
			},
			{
				Key:  "c4",
				Name: "концентрация ПГС4: конец шк.к.1",
			},
			{
				Key:  "c5",
				Name: "концентрация ПГС5: середина шк.к.2",
			},
			{
				Key:  "c6",
				Name: "концентрация ПГС6: конец шк.к.2",
			},

			{
				Key:  "d_chan1_sc_beg_t_norm",
				Name: "абс.погр. к.1: начало шк.: нормальная температура",
			},
			{
				Key:  "d_chan1_sc_mid_t_norm",
				Name: "абс.погр. к.1: середина шк.: нормальная температура",
			},
			{
				Key:  "d_chan1_sc_end_t_norm",
				Name: "абс.погр. к.1: конец шк.: нормальная температура",
			},

			{
				Key:  "d_chan1_sc_beg_t_low",
				Name: "абс.погр. к.1: начало шк.: низкая температура",
			},
			{
				Key:  "d_chan1_sc_mid_t_low",
				Name: "абс.погр. к.1: середина шк.: низкая температура",
			},
			{
				Key:  "d_chan1_sc_end_t_low",
				Name: "абс.погр. к.1: конец шк.: низкая температура",
			},

			{
				Key:  "d_chan1_sc_beg_t_high",
				Name: "абс.погр. к.1: начало шк.: высокая температура",
			},
			{
				Key:  "d_chan1_sc_mid_t_high",
				Name: "абс.погр. к.1: середина шк.: высокая температура",
			},
			{
				Key:  "d_chan1_sc_end_t_high",
				Name: "абс.погр. к.1: конец шк.: высокая температура",
			},

			{
				Key:  keyTempNorm.String(),
				Name: "уставка температуры НКУ,⁰C",
			},
			{
				Key:  keyTempLow.String(),
				Name: "уставка низкой температуры,⁰C",
			},
			{
				Key:  keyTempHigh.String(),
				Name: "уставка высокой температуры,⁰C",
			},
			{
				Key:  keySendCmdSetWorkModeIntervalSec,
				Name: "интервал отправки команды установки режима работы АНКАТ",
			},
		},
		InitParty: initParty,
		Work:      main,

		DataSections: dataSections,

		ProductTypesVars: func() []devdata.ProductTypeVars {
			xsC2 := devdata.ProductTypeVars{
				ParamsRngList: varsParamRng(anktvar.VarsChan2),
			}
			for _, t := range productTypesList {
				if t.Chan2 {
					xsC2.Names = append(xsC2.Names, t.String())
				}
			}

			xsP := devdata.ProductTypeVars{
				ParamsRngList: varsParamRng(anktvar.VarsP),
			}
			for _, t := range productTypesList {
				if t.Press {
					xsP.Names = append(xsP.Names, t.String())
				}
			}
			return []devdata.ProductTypeVars{xsC2, xsP}
		}(),

		OnReadProduct: onReadProduct,

		Calc: nil,
	}

	deviceConfig = devicecfg.Device{
		Baud:               9600,
		TimeoutGetResponse: time.Second,
		TimeoutEndResponse: 50 * time.Millisecond,
		MaxAttemptsRead:    5,
		Pause:              50 * time.Millisecond,
		NetAddr:            7,
		FloatFormat:        modbus.BCD,
		CfsList: []devicecfg.Cfs{
			{0, 50},
		},
		VarsNames: anktvar.Names,
		CfsNames:  KfsNames,
		Vars:      varsParamRng(anktvar.Vars),
	}
)

type testPt struct {
	scalePt scalePt
	chan2   bool
	keyTemp keyTemp
}

type scalePt int

const (
	scaleBeg scalePt = iota
	scaleMid
	scaleEnd
)

const (
	deviceName = "Анкат-7664МИКРО"
)

func varsParamRng(vars []modbus.Var) (xs []devicecfg.Vars) {
	for _, v := range vars {
		xs = append(xs, devicecfg.Vars{v, 1})
	}
	return
}
