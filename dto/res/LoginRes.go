package res

import "HospitalManager/model"

type LoginRes struct {
	User         model.Users
	AccessToken  string
	RefreshToken string
}
