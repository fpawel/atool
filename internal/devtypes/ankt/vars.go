package ankt

import (
	"github.com/fpawel/comm/modbus"
)

var (
	varCCh0    modbus.Var = 0
	varCCh1    modbus.Var = 2
	varCCh2    modbus.Var = 4
	varPkPa    modbus.Var = 6
	varPmm     modbus.Var = 8
	varTmcu    modbus.Var = 10
	varVbat    modbus.Var = 12
	varVref    modbus.Var = 14
	varVmcu    modbus.Var = 16
	varVdatP   modbus.Var = 18
	varCoutCh0 modbus.Var = 640
	varTppCh0  modbus.Var = 642
	varILOn0   modbus.Var = 644
	varILOff0  modbus.Var = 646
	varUwCh0   modbus.Var = 648
	varUrCh0   modbus.Var = 650
	varWORK0   modbus.Var = 652
	varREF0    modbus.Var = 654
	varVar1Ch0 modbus.Var = 656
	varVar2Ch0 modbus.Var = 658
	varVar3Ch0 modbus.Var = 660
	varFppCh0  modbus.Var = 662
	varCoutCh1 modbus.Var = 672
	varTppCh1  modbus.Var = 674
	varILOn1   modbus.Var = 676
	varILOff1  modbus.Var = 678
	varUwCh1   modbus.Var = 680
	varUrCh1   modbus.Var = 682
	varWORK1   modbus.Var = 684
	varREF1    modbus.Var = 686
	varVar1Ch1 modbus.Var = 688
	varVar2Ch1 modbus.Var = 690
	varVar3Ch1 modbus.Var = 692
	varFppCh1  modbus.Var = 694

	paramsNames = map[modbus.Var]string{
		varCCh0:    "Концентрация 1: электрохимия",
		varCCh1:    "Концентрация 2: электрохимия 2: оптика 1",
		varCCh2:    "Концентрация 3: оптика 1: оптика 2",
		varPkPa:    "Давление, кПа",
		varPmm:     "Давление, мм. рт. ст",
		varTmcu:    "Температура микроконтроллера, град.С",
		varVbat:    "Напряжение аккумуляторной батареи, В",
		varVref:    "Опорное напряжение для электрохимии, В",
		varVmcu:    "Напряжение питания микроконтроллера, В",
		varVdatP:   "Напряжение датчика давления, В",
		varCoutCh0: "Концентрация - первый канал оптики",
		varTppCh0:  "Температура пироприемника: оптика 1",
		varILOn0:   "Лампа ВКЛ: оптика 1",
		varILOff0:  "Лампа ВЫКЛ: оптика 1",
		varUwCh0:   "Рабочий канал АЦП: оптика 1",
		varUrCh0:   "Опорный канал АЦП: оптика 1",
		varWORK0:   "Норм. рабочий канал АЦП: оптика 1",
		varREF0:    "Норм. опроный канал АЦП: оптика 1",
		varVar1Ch0: "Дифф. сигнал:  оптика 1",
		varVar2Ch0: "Дифф. сигнал с поправкой по нулю от T: оптика 1",
		varVar3Ch0: "Дифф. сигнал с поправкой по чувст. от T: оптика 1",
		varFppCh0:  "Частота преобразования АЦП: оптика 1",
		varCoutCh1: "Концентрация: оптика 2",
		varTppCh1:  "Температура пироприемника: оптика 2",
		varILOn1:   "Лампа ВКЛ: оптика 2",
		varILOff1:  "Лампа ВЫКЛ: оптика 2",
		varUwCh1:   "Рабочий канал АЦП: оптика 2",
		varUrCh1:   "Опорный канал АЦП: оптика 2",
		varWORK1:   "Норм. рабочий канал АЦП: оптика 2",
		varREF1:    "Норм. опроный канал АЦП: оптика 2",
		varVar1Ch1: "Дифф. сигнал:  оптика 2",
		varVar2Ch1: "Дифф. сигнал с поправкой по нулю от T: оптика 2",
		varVar3Ch1: "Дифф. сигнал с поправкой по чувст. от T: оптика 2",
		varFppCh1:  "Частота преобразования АЦП: оптика 2",
	}

	varsP = []modbus.Var{
		varPkPa,
		varPmm,
		varVdatP,
	}

	varsCommon = []modbus.Var{
		varCCh0,
		varCCh1,
		varCCh2,
		varTmcu,
		varVbat,
		varVref,
		varVmcu,
	}

	varsChan1 = []modbus.Var{
		varCoutCh0,
		varTppCh0,
		varILOn0,
		varILOff0,
		varUwCh0,
		varUrCh0,
		varWORK0,
		varREF0,
		varVar1Ch0,
		varVar2Ch0,
		varVar3Ch0,
		varFppCh0,
	}
	varsChan2 = []modbus.Var{
		varCoutCh1,
		varTppCh1,
		varILOn1,
		varILOff1,
		varUwCh1,
		varUrCh1,
		varWORK1,
		varREF1,
		varVar1Ch1,
		varVar2Ch1,
		varVar3Ch1,
		varFppCh1,
	}
)
