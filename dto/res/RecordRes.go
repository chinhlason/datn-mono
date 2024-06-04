package res

import (
	"HospitalManager/model"
	"time"
)

type RecordRes struct {
	Id         string
	Status     string
	Patient    model.Patients
	Doctor     model.Users
	Updater    model.Users
	Notes      []Note
	Beds       []BedRes
	Devices    []DeviceHistory
	History    []HistoryRes
	CurrBed    BedRes
	CurrDevice model.Devices
	CreateAt   time.Time
	UpdateAt   time.Time
}
