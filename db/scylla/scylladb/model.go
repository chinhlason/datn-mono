package scylladb

import "time"

type UpdateProfileReq struct {
	Fullname string
	Email    string
	Phone    string
	Id       string
	UpdateAt time.Time
}

type UpdateRole struct {
	Role string
	Id   string
}

type ChangePsw struct {
	Password string
	Id       string
}

type UpdateToken struct {
	Value string
	Id    string
}

type UpdateRoomReq struct {
	Name          string
	BedNumber     int
	PatientNumber int
	Id            string
	UpdateAt      time.Time
}

type UpdateBedNumber struct {
	BedNumber int
	Id        string
	UpdateAt  time.Time
}

type UpdatePatientNumber struct {
	PatientNumber int
	Id            string
	UpdateAt      time.Time
}

type HandOver struct {
	IdDoctor string
	UpdateAt time.Time
	Id       string
}

type UpdateBedReq struct {
	Name     string
	IdRoom   string
	Status   string
	Id       string
	UpdateAt time.Time
}

type UpdateBedStt struct {
	Status   string
	Id       string
	UpdateAt time.Time
}

type UpdateDevice struct {
	Serial   string
	Id       string
	UpdateAt time.Time
}

type UpdateDeviceStt struct {
	Status   string
	Id       string
	UpdateAt time.Time
}

type UpdateUpdaterRecord struct {
	IdUpdater string
	Id        string
	UpdateAt  time.Time
}
type UpdateUsageTable struct {
	Status string
	EndAt  time.Time
	Id     string
}

type UpdateRecordStt struct {
	Status   string
	Id       string
	UpdateAt time.Time
}

type DisableOrEnable struct {
	Status   string
	Id       string
	UpdateAt time.Time
}
