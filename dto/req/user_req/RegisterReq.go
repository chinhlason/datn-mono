package user_req

type RegisterReq struct {
	DoctorCode string `validate:"required"`
	Password   string `validate:"required"`
	Fullname   string `validate:"required"`
	Email      string `validate:"required"`
	Phone      string `validate:"required"`
}
