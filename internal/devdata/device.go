package devdata

var Devices = make(map[string]Device)

type Device struct {
	DataSections DataSections
}
