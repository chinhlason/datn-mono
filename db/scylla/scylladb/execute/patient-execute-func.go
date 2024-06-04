package execute

import (
	"HospitalManager/dto/req/record_req"
	"HospitalManager/model"
	"context"
	"fmt"
	"github.com/gocql/gocql"
	"github.com/scylladb/gocqlx/v2/qb"
	"time"
)

func (q *Queries) InsertPatient(req record_req.InsertRecordReq, IdPatient gocql.UUID, patientCode string) error {
	_, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	tableName := fmt.Sprintf("%s.patients", q.keyspace)
	insert := &model.Patients{
		Id:            IdPatient,
		PatientCode:   patientCode,
		Fullname:      req.Fullname,
		Ccid:          req.Ccid,
		Address:       req.Address,
		Dob:           req.Dob,
		Gender:        req.Gender,
		Phone:         req.Phone,
		RelativeName:  req.RelativeName,
		RelativePhone: req.RelativePhone,
		Reason:        req.Reason,
		CreateAt:      time.Now(),
		UpdateAt:      time.Now(),
	}
	stmt := qb.Insert(tableName).
		Columns("id", "patient_code", "fullname", "ccid", "address", "dob", "gender", "phone",
			"relative_name", "relative_phone", "reason", "create_at", "update_at").
		Query(q.session)
	stmt.BindStruct(insert)
	if err := stmt.ExecRelease(); err != nil {
		return err
	}
	return nil
}

func (q *Queries) GetPatient(value string, option string) ([]model.Patients, error) {
	_, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	tableName := fmt.Sprintf("%s.patients", q.keyspace)
	var patients []model.Patients
	stmt, names := qb.Select(tableName).
		Where(qb.Eq(option)).
		ToCql()
	stmt += " ALLOW FILTERING"
	query := q.session.Query(stmt, names).BindMap(qb.M{
		option: value,
	})
	if err := query.SelectRelease(&patients); err != nil {
		return nil, err
	}
	return patients, nil
}
