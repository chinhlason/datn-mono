package model

import (
	"github.com/gocql/gocql"
	"time"
)

type Users struct {
	Id         gocql.UUID `json:"id"`
	DoctorCode string     `json:"doctor_code"`
	Password   string     `json:"-"`
	Fullname   string     `json:"fullname"`
	Email      string     `json:"email"`
	Phone      string     `json:"phone"`
	Role       string     `json:"role"`
	CreateAt   time.Time  `json:"create_at"`
	UpdateAt   time.Time  `json:"update_at"`
}

type Rooms struct {
	Id            gocql.UUID `json:"id"`
	IdDoctor      gocql.UUID `json:"id_doctor"`
	Name          string     `json:"name"`
	BedNumber     int        `json:"bed_number"`
	PatientNumber int        `json:"patient_number"`
	CreateAt      time.Time  `json:"create_at"`
	UpdateAt      time.Time  `json:"update_at"`
}

type Beds struct {
	Id       gocql.UUID `json:"id"`
	IdRoom   gocql.UUID `json:"id_room"`
	Name     string     `json:"name"`
	Status   string     `json:"status"`
	CreateAt time.Time  `json:"create_at"`
	UpdateAt time.Time  `json:"update_at"`
}

type MedicalRecords struct {
	Id        gocql.UUID `json:"id"`
	IdPatient gocql.UUID `json:"id_patient"`
	IdDoctor  gocql.UUID `json:"id_doctor"`
	IdUpdater gocql.UUID `json:"id_updater"`
	Status    string     `json:"status"`
	CreateAt  time.Time  `json:"create_at"`
	UpdateAt  time.Time  `json:"update_at"`
}

type Notes struct {
	Id       gocql.UUID `json:"id"`
	IdRecord gocql.UUID `json:"id_record"`
	IdDoctor gocql.UUID `json:"id_doctor"`
	Content  string     `json:"content"`
	ImgUrl   string     `json:"img_url"`
	CreateAt time.Time  `json:"create_at"`
	UpdateAt time.Time  `json:"update_at"`
}

type Devices struct {
	Id       gocql.UUID `json:"id"`
	Serial   string     `json:"serial"`
	Warraty  int        `json:"warraty"`
	Status   string     `json:"status"`
	CreateAt time.Time  `json:"create_at"`
	UpdateAt time.Time  `json:"update_at"`
}

type Patients struct {
	Id            gocql.UUID `json:"id"`
	PatientCode   string     `json:"patient_code"`
	Fullname      string     `json:"fullname"`
	Ccid          string     `json:"ccid"`
	Address       string     `json:"address"`
	Dob           string     `json:"dob"`
	Gender        string     `json:"gender"`
	Phone         string     `json:"phone"`
	RelativeName  string     `json:"relative_name"`
	RelativePhone string     `json:"relative_phone"`
	Reason        string     `json:"reason"`
	CreateAt      time.Time  `json:"create_at"`
	UpdateAt      time.Time  `json:"update_at"`
}

type UsageBed struct {
	Id       gocql.UUID `json:"id"`
	IdBed    gocql.UUID `json:"id_bed"`
	IdRecord gocql.UUID `json:"id_record"`
	Status   string     `json:"status"`
	CreateAt time.Time  `json:"create_at"`
	EndAt    time.Time  `json:"end_at"`
}

type UsageDevice struct {
	Id       gocql.UUID `json:"id"`
	IdDevice gocql.UUID `json:"id_device"`
	IdRecord gocql.UUID `json:"id_record"`
	Status   string     `json:"status"`
	CreateAt time.Time  `json:"create_at"`
	EndAt    time.Time  `json:"end_at"`
}

type RecordHistory struct {
	Id       gocql.UUID `json:"id"`
	IdRecord gocql.UUID `json:"id_record"`
	IdDoctor gocql.UUID `json:"id_doctor"`
	Content  string     `json:"content"`
	CreateAt time.Time  `json:"create_at"`
}

type RecordHistoryStr struct {
	Id       gocql.UUID `json:"id"`
	IdRecord string     `json:"id_record"`
	IdDoctor gocql.UUID `json:"id_doctor"`
	Content  string     `json:"content"`
	CreateAt time.Time  `json:"create_at"`
}

type UsageRoom struct {
	ID       gocql.UUID `json:"id"`
	IdRoom   gocql.UUID `json:"id_room"`
	IdRecord gocql.UUID `json:"id_record"`
	Status   string     `json:"status"`
	CreateAt time.Time  `json:"create_at"`
	EndAt    time.Time  `json:"end_at"`
}
