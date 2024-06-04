package patient_req

type InsertPatientReq struct {
	Fullname      string `json:"fullname"`
	Ccid          string `json:"ccid"`
	Address       string `json:"address"`
	Dob           string `json:"dob"`
	Gender        string `json:"gender"`
	Phone         string `json:"phone"`
	RelativeName  string `json:"relative_name"`
	RelativePhone string `json:"relative_phone"`
	Reason        string `json:"reason"`
}
