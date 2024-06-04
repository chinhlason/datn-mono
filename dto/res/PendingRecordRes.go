package res

import "time"

type PendingRecordRes struct {
	Id          string
	DoctorCode  string
	PatientCode string
	Fullname    string
	Phone       string
	Detail      string
	CreateAt    time.Time
}
