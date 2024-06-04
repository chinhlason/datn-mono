package res

import (
	"HospitalManager/model"
	"github.com/gocql/gocql"
	"time"
)

type Note struct {
	Id       gocql.UUID `json:"id"`
	IdRecord gocql.UUID `json:"id_record"`
	Content  string     `json:"content"`
	ImgUrl   string     `json:"img_url"`
	CreateAt time.Time  `json:"create_at"`
	UpdateAt time.Time  `json:"update_at"`
	Doctor   model.Users
}
