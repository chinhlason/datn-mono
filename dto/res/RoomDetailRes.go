package res

import "HospitalManager/model"

type RoomDetailRes struct {
	Id            string
	PatientNumber int
	BedNumber     int
	Leader        model.Users
	Members       []model.Users
}
