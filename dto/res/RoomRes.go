package res

import (
	"HospitalManager/model"
	"github.com/gocql/gocql"
	"time"
)

type Room struct {
	Id            gocql.UUID `json:"id"`
	Name          string     `json:"name"`
	BedNumber     int        `json:"bed_number"`
	PatientNumber int        `json:"patient_number"`
	CreateAt      time.Time  `json:"create_at"`
	UpdateAt      time.Time  `json:"update_at"`
	Doctor        model.Users
}
