package res

import (
	"HospitalManager/model"
	"time"
)

type DeviceRes struct {
	DeviceNumber int
	Devices      []model.Devices
}

type DeviceHistory struct {
	Device     model.Devices
	InuseAt    time.Time
	NotInuseAt time.Time
}
