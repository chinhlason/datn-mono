package res

import (
	"HospitalManager/model"
	"github.com/gocql/gocql"
	"time"
)

type BedRes struct {
	Id         gocql.UUID `json:"id"`
	Name       string     `json:"name"`
	Status     string     `json:"status"`
	CreateAt   time.Time  `json:"create_at"`
	UpdateAt   time.Time  `json:"update_at"`
	Room       model.Rooms
	InuseAt    time.Time
	NotInuseAt time.Time
}

type BedRecord struct {
	BedName     string
	RoomName    string
	IdRecord    string
	PatientName string
}
