package res

import (
	"HospitalManager/model"
	"time"
)

type DeviceInUse struct {
	Device   model.Devices
	IdRecord string
	Room     string
	Bed      string
	InUseAt  time.Time
}
