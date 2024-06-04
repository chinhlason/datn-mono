package res

import (
	"HospitalManager/model"
	"time"
)

type HistoryRes struct {
	Id       string
	Doctor   model.Users
	Content  string
	CreateAt time.Time
}
