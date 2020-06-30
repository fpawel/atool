package anktvar

import "github.com/fpawel/comm/modbus"

var (
	CCh0    modbus.Var = 0
	CCh1    modbus.Var = 2
	CCh2    modbus.Var = 4
	PkPa    modbus.Var = 6
	Pmm     modbus.Var = 8
	Tmcu    modbus.Var = 10
	Vbat    modbus.Var = 12
	Vref    modbus.Var = 14
	Vmcu    modbus.Var = 16
	VdatP   modbus.Var = 18
	CoutCh0 modbus.Var = 640
	TppCh0  modbus.Var = 642
	ILOn0   modbus.Var = 644
	ILOff0  modbus.Var = 646
	UwCh0   modbus.Var = 648
	UrCh0   modbus.Var = 650
	WORK0   modbus.Var = 652
	REF0    modbus.Var = 654
	Var1Ch0 modbus.Var = 656
	Var2Ch0 modbus.Var = 658
	Var3Ch0 modbus.Var = 660
	FppCh0  modbus.Var = 662
	CoutCh1 modbus.Var = 672
	TppCh1  modbus.Var = 674
	ILOn1   modbus.Var = 676
	ILOff1  modbus.Var = 678
	UwCh1   modbus.Var = 680
	UrCh1   modbus.Var = 682
	WORK1   modbus.Var = 684
	REF1    modbus.Var = 686
	Var1Ch1 modbus.Var = 688
	Var2Ch1 modbus.Var = 690
	Var3Ch1 modbus.Var = 692
	FppCh1  modbus.Var = 694

	Names = map[modbus.Var]string{
		CCh0:    "Концентрация 1: электрохимия",
		CCh1:    "Концентрация 2: электрохимия 2: оптика 1",
		CCh2:    "Концентрация 3: оптика 1: оптика 2",
		PkPa:    "Давление, кПа",
		Pmm:     "Давление, мм. рт. ст",
		Tmcu:    "Температура микроконтроллера, град.С",
		Vbat:    "Напряжение аккумуляторной батареи, В",
		Vref:    "Опорное напряжение для электрохимии, В",
		Vmcu:    "Напряжение питания микроконтроллера, В",
		VdatP:   "Напряжение датчика давления, В",
		CoutCh0: "Концентрация - первый канал оптики",
		TppCh0:  "Температура пироприемника: оптика 1",
		ILOn0:   "Лампа ВКЛ: оптика 1",
		ILOff0:  "Лампа ВЫКЛ: оптика 1",
		UwCh0:   "Рабочий канал АЦП: оптика 1",
		UrCh0:   "Опорный канал АЦП: оптика 1",
		WORK0:   "Норм. рабочий канал АЦП: оптика 1",
		REF0:    "Норм. опроный канал АЦП: оптика 1",
		Var1Ch0: "Дифф. сигнал:  оптика 1",
		Var2Ch0: "Дифф. сигнал с поправкой по нулю от T: оптика 1",
		Var3Ch0: "Дифф. сигнал с поправкой по чувст. от T: оптика 1",
		FppCh0:  "Частота преобразования АЦП: оптика 1",
		CoutCh1: "Концентрация: оптика 2",
		TppCh1:  "Температура пироприемника: оптика 2",
		ILOn1:   "Лампа ВКЛ: оптика 2",
		ILOff1:  "Лампа ВЫКЛ: оптика 2",
		UwCh1:   "Рабочий канал АЦП: оптика 2",
		UrCh1:   "Опорный канал АЦП: оптика 2",
		WORK1:   "Норм. рабочий канал АЦП: оптика 2",
		REF1:    "Норм. опроный канал АЦП: оптика 2",
		Var1Ch1: "Дифф. сигнал:  оптика 2",
		Var2Ch1: "Дифф. сигнал с поправкой по нулю от T: оптика 2",
		Var3Ch1: "Дифф. сигнал с поправкой по чувст. от T: оптика 2",
		FppCh1:  "Частота преобразования АЦП: оптика 2",
	}

	VarsP = []modbus.Var{
		PkPa,
		Pmm,
		VdatP,
	}

	Vars = []modbus.Var{
		CCh0,
		CCh1,
		CCh2,
		Tmcu,
		Vbat,
		Vref,
		Vmcu,

		CoutCh0,
		TppCh0,
		ILOn0,
		ILOff0,
		UwCh0,
		UrCh0,
		WORK0,
		REF0,
		Var1Ch0,
		Var2Ch0,
		Var3Ch0,
		FppCh0,
	}

	VarsChan2 = []modbus.Var{
		CoutCh1,
		TppCh1,
		ILOn1,
		ILOff1,
		UwCh1,
		UrCh1,
		WORK1,
		REF1,
		Var1Ch1,
		Var2Ch1,
		Var3Ch1,
		FppCh1,
	}
)
