package ankt

import "github.com/fpawel/atool/internal/config/devicecfg"

type kef = devicecfg.Kef

const (
	kefSoftVersion kef = 0
	kefDeviceType  kef = 1
	kefYEAR        kef = 2
	kefSerial      kef = 3
	kefKef4        kef = 4
	kefUnits1      kef = 5
	kefGasType1    kef = 6
	kefScale1      kef = 7
	kefScale1Begin kef = 8
	kefScale1End   kef = 9
	kefScale1C1    kef = 10
	kefScale1C3    kef = 11
	kefKNull1      kef = 12
	kefKSens1      kef = 13
	kefUnits2      kef = 14
	kefGasType2    kef = 15
	kefScale2      kef = 16
	kefScale2Begin kef = 17
	kefScale2End   kef = 18
	kefScale2C1    kef = 19
	kefScale2C3    kef = 20
	kefKNull2      kef = 21
	kefKSens2      kef = 22
	kefCh1Lin0     kef = 23
	kefCh1Lin1     kef = 24
	kefCh1Lin2     kef = 25
	kefCh1Lin3     kef = 26
	kefCh1T0v0     kef = 27
	kefCh1T0v1     kef = 28
	kefCh1T0v2     kef = 29
	kefCh1TKv0     kef = 30
	kefCh1TKv1     kef = 31
	kefCh1TKv2     kef = 32
	kefCh2Lin0     kef = 33
	kefCh2Lin1     kef = 34
	kefCh2Lin2     kef = 35
	kefCh2Lin3     kef = 36
	kefCh2T0v0     kef = 37
	kefCh2T0v1     kef = 38
	kefCh2T0v2     kef = 39
	kefCh2TKv0     kef = 40
	kefCh2TKv1     kef = 41
	kefCh2TKv2     kef = 42
	kefP0          kef = 43
	kefP1          kef = 44
	kefPT0         kef = 45
	kefPT1         kef = 46
	kefPT2         kef = 47
	kefKdFt        kef = 48
	kefKFt         kef = 49
)

var KfsNames = map[kef]string{
	kefSoftVersion: "Номер версии ПО",
	kefDeviceType:  "Номер исполнения прибора",
	kefYEAR:        "Год выпуска",
	kefSerial:      "Серийный номер",
	kefKef4:        "Максимальное число регистров в таблице регистров прибора",
	kefUnits1:      "Единицы измерения канала 1 ИКД",
	kefGasType1:    "Величина, измеряемая каналом 1 ИКД",
	kefScale1:      "Диапазон измерений канала 1 ИКД",
	kefScale1Begin: "Начало шкалы канала 1 ИКД",
	kefScale1End:   "Конец шкалы канала 1 ИКД",
	kefScale1C1:    "Значение ПГС1 (начало шкалы) канала 1 ИКД",
	kefScale1C3:    "Значение ПГС3 (конец шкалы) канала 1 ИКД",
	kefKNull1:      "Коэффициент калибровки нуля канала 1 ИКД",
	kefKSens1:      "Коэффициент калибровки чувствительности канала 1 ИКД",
	kefUnits2:      "Единицы измерения канала 2 ИКД",
	kefGasType2:    "Величина, измеряемая каналом 2 ИКД",
	kefScale2:      "Диапазон измерений канала 2 ИКД",
	kefScale2Begin: "Начало шкалы канала 2 ИКД",
	kefScale2End:   "Конец шкалы канала 2 ИКД",
	kefScale2C1:    "ПГС1 (начало шкалы) канала 2 ИКД",
	kefScale2C3:    "ПГС3 (конец шкалы) канала 2 ИКД",
	kefKNull2:      "Коэффициент калибровки нуля канала 2 ИКД",
	kefKSens2:      "Коэффициент калибровки чувствительности канала 2 ИКД",
	kefCh1Lin0:     "0-ой степени кривой линеаризации канала 1 ИКД",
	kefCh1Lin1:     "1-ой степени кривой линеаризации канала 1 ИКД",
	kefCh1Lin2:     "2-ой степени кривой линеаризации канала 1 ИКД",
	kefCh1Lin3:     "3-ей степени кривой линеаризации канала 1 ИКД",
	kefCh1T0v0:     "0-ой степени полинома коррекции нуля от температуры канала 1 ИКД",
	kefCh1T0v1:     "1-ой степени полинома коррекции нуля от температуры канала 1 ИКД",
	kefCh1T0v2:     "2-ой степени полинома коррекции нуля от температуры канала 1 ИКД",
	kefCh1TKv0:     "0-ой степени полинома кор. чувств. от температуры канала 1 ИКД",
	kefCh1TKv1:     "1-ой степени полинома кор. чувств. от температуры канала 1 ИКД",
	kefCh1TKv2:     "2-ой степени полинома кор. чувств. от температуры канала 1 ИКД",
	kefCh2Lin0:     "0-ой степени кривой линеаризации канала 2 ИКД",
	kefCh2Lin1:     "1-ой степени кривой линеаризации канала 2 ИКД",
	kefCh2Lin2:     "2-ой степени кривой линеаризации канала 2 ИКД",
	kefCh2Lin3:     "3-ей степени кривой линеаризации канала 2 ИКД",
	kefCh2T0v0:     "0-ой степени полинома коррекции нуля от температуры канала 2 ИКД",
	kefCh2T0v1:     "1-ой степени полинома коррекции нуля от температуры канала 2 ИКД",
	kefCh2T0v2:     "2-ой степени полинома коррекции нуля от температуры канала 2 ИКД",
	kefCh2TKv0:     "0-ой степени полинома кор. чувств. от температуры канала 2 ИКД",
	kefCh2TKv1:     "1-ой степени полинома кор. чувств. от температуры канала 2 ИКД",
	kefCh2TKv2:     "2-ой степени полинома кор. чувств. от температуры канала 2 ИКД",
	kefP0:          "0-ой степени полинома калибровки датчика P (в мм.рт.ст.)",
	kefP1:          "1-ой степени полинома калибровки датчика P (в мм.рт.ст.)",
	kefPT0:         "0-ой степени полинома кор. нуля датчика давления от температуры",
	kefPT1:         "1-ой степени полинома кор. нуля датчика давления от температуры",
	kefPT2:         "2-ой степени полинома кор. нуля датчика давления от температуры",
	kefKdFt:        "Чувствительность датчика температуры микроконтроллера, град.С/В",
	kefKFt:         "Смещение датчика температуры микроконтроллера, град.С",
}
