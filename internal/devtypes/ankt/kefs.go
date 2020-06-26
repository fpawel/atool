package ankt

import "github.com/fpawel/atool/internal/config/devicecfg"

var KfsNames = map[devicecfg.Kef]string{
	0:  "Номер версии ПО",
	1:  "Номер исполнения прибора",
	2:  "Год выпуска",
	3:  "Серийный номер",
	4:  "Максимальное число регистров в таблице регистров прибора",
	5:  "Единицы измерения канала 1 ИКД",
	6:  "Величина, измеряемая каналом 1 ИКД",
	7:  "Диапазон измерений канала 1 ИКД",
	8:  "Начало шкалы канала 1 ИКД",
	9:  "Конец шкалы канала 1 ИКД",
	10: "Значение ПГС1 (начало шкалы) канала 1 ИКД",
	11: "Значение ПГС3 (конец шкалы) канала 1 ИКД",
	12: "Коэффициент калибровки нуля канала 1 ИКД",
	13: "Коэффициент калибровки чувствительности канала 1 ИКД",
	14: "Единицы измерения канала 2 ИКД",
	15: "Величина, измеряемая каналом 2 ИКД",
	16: "Диапазон измерений канала 2 ИКД",
	17: "Начало шкалы канала 2 ИКД",
	18: "Конец шкалы канала 2 ИКД",
	19: "ПГС1 (начало шкалы) канала 2 ИКД",
	20: "ПГС3 (конец шкалы) канала 2 ИКД",
	21: "Коэффициент калибровки нуля канала 2 ИКД",
	22: "Коэффициент калибровки чувствительности канала 2 ИКД",
	23: "0-ой степени кривой линеаризации канала 1 ИКД",
	24: "1-ой степени кривой линеаризации канала 1 ИКД",
	25: "2-ой степени кривой линеаризации канала 1 ИКД",
	26: "3-ей степени кривой линеаризации канала 1 ИКД",
	27: "0-ой степени полинома коррекции нуля от температуры канала 1 ИКД",
	28: "1-ой степени полинома коррекции нуля от температуры канала 1 ИКД",
	29: "2-ой степени полинома коррекции нуля от температуры канала 1 ИКД",
	30: "0-ой степени полинома кор. чувств. от температуры канала 1 ИКД",
	31: "1-ой степени полинома кор. чувств. от температуры канала 1 ИКД",
	32: "2-ой степени полинома кор. чувств. от температуры канала 1 ИКД",
	33: "0-ой степени кривой линеаризации канала 2 ИКД",
	34: "1-ой степени кривой линеаризации канала 2 ИКД",
	35: "2-ой степени кривой линеаризации канала 2 ИКД",
	36: "3-ей степени кривой линеаризации канала 2 ИКД",
	37: "0-ой степени полинома коррекции нуля от температуры канала 2 ИКД",
	38: "1-ой степени полинома коррекции нуля от температуры канала 2 ИКД",
	39: "2-ой степени полинома коррекции нуля от температуры канала 2 ИКД",
	40: "0-ой степени полинома кор. чувств. от температуры канала 2 ИКД",
	41: "1-ой степени полинома кор. чувств. от температуры канала 2 ИКД",
	42: "2-ой степени полинома кор. чувств. от температуры канала 2 ИКД",
	43: "0-ой степени полинома калибровки датчика P (в мм.рт.ст.)",
	44: "1-ой степени полинома калибровки датчика P (в мм.рт.ст.)",
	45: "0-ой степени полинома кор. нуля датчика давления от температуры",
	46: "1-ой степени полинома кор. нуля датчика давления от температуры",
	47: "2-ой степени полинома кор. нуля датчика давления от температуры",
	48: "Чувствительность датчика температуры микроконтроллера, град.С/В",
	49: "Смещение датчика температуры микроконтроллера, град.С",
}
